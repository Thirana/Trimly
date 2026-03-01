# Neon + Postgres Migration Runbook (Phase 3)

This document summarizes the exact setup flow we used in this project and the issues we hit.
Use this as a reusable checklist for future Go projects.

## What this runbook covers

1. Environment setup with `.env`.
2. Required Go/tool dependencies.
3. Running SQL migrations with `golang-migrate`.
4. Verifying DB state with `psql`.
5. Common errors, root causes, and fixes.

## Stack used in this step

1. Go app runtime config loader: `github.com/joho/godotenv`
2. Postgres driver/pool: `github.com/jackc/pgx/v5` (`pgxpool`)
3. Migration tool (CLI): `github.com/golang-migrate/migrate/v4/cmd/migrate`
4. Managed DB: Neon Postgres

## 1) Create `.env`

Create `.env` in repo root:

```bash
cat > .env <<'EOF'
PORT=8080
BASE_URL='http://localhost:8080'
DATABASE_URL='postgres://<user>:<password>@<host>/<db>?sslmode=require&channel_binding=require'
REDIS_URL='redis://default:<password>@<host>:6379'
EOF
```

Important:

1. Quote URL values.
2. If URL contains `&` and you do `source .env` without quotes, shell parsing fails.

## 2) Install/refresh project dependencies

Go equivalent of `npm install`:

```bash
go mod tidy
go mod download
```

Validate build/tests:

```bash
go test ./...
```

## 3) Install migration CLI (`migrate`)

Install once:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Verify:

```bash
migrate -version
```

If command not found:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
migrate -version
```

## 4) Why `set -a; source .env; set +a` is needed

`go run ./cmd/api` auto-loads `.env` (via `godotenv`).

But external CLI tools like `migrate` and `psql` do **not** auto-load `.env`.
They only see exported shell environment variables.

Use this once per terminal session:

```bash
set -a
source .env
set +a
```

## 5) Run migrations

Apply migrations:

```bash
migrate -path internal/store/postgres/migrations -database "$DATABASE_URL" up
```

Rollback one step:

```bash
migrate -path internal/store/postgres/migrations -database "$DATABASE_URL" down 1
```

If migration already applied, you may see `no change`.

## 6) What is `psql` and why use it

`psql` is the official Postgres command-line client.
Use it to verify the exact DB your app/CLI is using.

Common checks:

```bash
psql "$DATABASE_URL" -c "select current_database(), current_user, current_schema();"
psql "$DATABASE_URL" -c "\dt"
psql "$DATABASE_URL" -c "\d+ links"
```

This is critical when Neon UI and local CLI appear inconsistent.

## 7) Start app and smoke test

Run API:

```bash
go run ./cmd/api
```

Expected startup log (when DB URL exists):

1. `using postgres store`
2. `starting api on :8080`

Smoke test:

```bash
curl -i http://localhost:8080/health

curl -i -X POST http://localhost:8080/v1/links \
  -H 'Content-Type: application/json' \
  -d '{"long_url":"https://example.com/some/path","expires_at":"2026-03-01T10:00:00Z"}'
```

## 8) Errors we hit and fixes

1. `.env: parse error near '&'`

- Cause: unquoted URL with `&` during `source .env`.
- Fix: quote URL values in `.env`.

1. `migrate: command not found`

- Cause: Go bin directory not in `PATH`.
- Fix: add `$(go env GOPATH)/bin` to shell `PATH`.

1. `500 internal` on `POST /v1/links` with table confusion

- Cause pattern: app/CLI pointing to different DB target or schema not migrated in active URL.
- Fix: verify active target with `psql`, run migration on the same `DATABASE_URL`.

1. `500 internal` after table existed

- Root cause in code (fixed): idempotency key used NUL byte separator (`\x00`), Postgres `TEXT` rejects NUL bytes.
- Fix: changed key format in `internal/store/intent.go` to a Postgres-safe deterministic format.

## 9) Fast copy/paste checklist for future projects

```bash
# from repo root
cat > .env <<'EOF'
PORT=8080
BASE_URL='http://localhost:8080'
DATABASE_URL='postgres://<user>:<password>@<host>/<db>?sslmode=require'
EOF

go mod tidy
go mod download
go test ./...

go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

set -a
source .env
set +a

migrate -path internal/store/postgres/migrations -database "$DATABASE_URL" up
psql "$DATABASE_URL" -c "\dt"

go run ./cmd/api
```

## Related files in this repo

1. `cmd/api/main.go`
2. `internal/store/postgres/store.go`
3. `internal/store/postgres/migrations/000001_create_links_table.up.sql`
4. `internal/store/postgres/migrations/000001_create_links_table.down.sql`
5. `internal/store/intent.go`
6. `docs/plan/phase-3-persistence-migrations.md`

