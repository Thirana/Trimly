# Phase 3: Persistence and Migrations (Postgres)

## Goal

Move from in-memory storage to durable Postgres persistence with clean migration workflow.

## Provider baseline for this phase

1. Managed Postgres target: `Neon`.
2. Backend hosting target for API runtime: `Render`.
3. Cache provider (`Redis Cloud`) is documented now and activated in Phase 4.
4. Frontend hosting (`Vercel`) is independent of store migration but part of deployment baseline.

## Tooling baseline for this phase

1. Migration tool: `golang-migrate` CLI (`migrate`).
2. Migration files: versioned SQL (`up/down`) under repo source control.
3. DB connection source: `DATABASE_URL` (Neon connection string in cloud environments).

## Current implementation milestone (2026-02-28)

1. Added Postgres store skeleton at `internal/store/postgres/store.go`.
2. Added first migration scaffold at `internal/store/postgres/migrations`.
3. Added runtime store selection in `cmd/api/main.go`:
- `DATABASE_URL` set -> Postgres
- otherwise -> in-memory fallback
4. Added startup DB ping fail-fast and graceful shutdown wiring.
5. Added local `.env` auto-loading via `godotenv`.

## Completion milestone (2026-03-01)

1. Added Postgres integration tests for `Save`, `Get`, `FindByIntent`, and unique constraint mapping.
2. Added migration verification integration flow (`up` from empty schema, `down`, then `up`).
3. Added safe internal handler logging for unexpected errors with generic client `500` responses.

## Scope

1. Postgres-backed link store implementation.
2. Schema migrations and local developer workflow.
3. Config and startup wiring for DB connectivity.
4. Repository-level tests covering store behavior.

## Implementation approach

1. Define schema for links:
- `id` (UUID or bigserial)
- `code` (unique)
- `original_url`
- `created_at`
- `updated_at`
- `expires_at` (nullable)
- ownership fields for future auth integration (`user_id`, nullable initially)
2. Introduce migration tooling:
- versioned `up/down` SQL files
- migration command integrated into dev workflow
3. Implement `internal/store` Postgres adapter behind existing interfaces.
4. Keep service layer unchanged where possible by preserving interface contracts.
5. Add connection handling:
- DSN from `DATABASE_URL`
- startup ping and fail-fast behavior
- graceful shutdown with context deadlines

## Data and API notes

1. Add unique indexes on `code`.
2. Add indexes needed for planned queries:
- `created_at`
- `user_id` (for future "my links" endpoints)
3. Keep API contract stable while switching storage backend.

## Testing plan

1. Store integration tests against disposable Postgres instance.
2. Migration test:
- apply all migrations from empty DB
- verify schema shape and constraints
3. Regression tests for existing handler/service behavior.

## Risks and mitigations

1. Migration drift across environments:
- enforce migration order and checksum validation where supported.
2. DB-specific behavior differences:
- test for unique violation and map to domain conflict error.
3. Local setup friction:
- provide `docker compose` or one-command setup for Postgres.
4. Managed DB assumptions differ from local:
- keep local migration workflow provider-agnostic
- validate on Neon-compatible configuration before rollout

## Exit criteria

1. DB schema versioned and reproducible.
2. Postgres store is the primary implementation path.
3. Duplicate code conflicts handled correctly.
4. CI/local tests run reliably with persistence enabled.

## Execution checklist (command-level)

Use this checklist before and during implementation.

1. Install migration CLI locally:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

2. Confirm `migrate` is on PATH:

```bash
migrate -version
```

3. Export Neon connection string for current shell:

```bash
export DATABASE_URL='postgres://<user>:<password>@<host>/<db>?sslmode=require'
```

4. Create migration folder:

```bash
mkdir -p internal/store/postgres/migrations
```

5. Create first migration pair:

```bash
migrate create -ext sql -dir internal/store/postgres/migrations -seq create_links_table
```

6. Fill generated SQL files with links schema and indexes:
- `code` unique index required
- include `expires_at`, timestamps, and future-proof columns planned in this phase

7. Apply migrations to Neon:

```bash
migrate -path internal/store/postgres/migrations -database "$DATABASE_URL" up
```

8. Verify schema quickly:

```bash
psql "$DATABASE_URL" -c '\dt'
psql "$DATABASE_URL" -c '\d+ links'
```

9. Add DB wiring and Postgres store implementation in codebase:
- read `DATABASE_URL`
- startup ping/fail-fast
- graceful close path
- preserve `LinkStore` behavior contracts

10. Run app with DB configuration:

```bash
DATABASE_URL="$DATABASE_URL" BASE_URL=http://localhost:8080 PORT=8080 go run ./cmd/api
```

11. Run regression and persistence tests:

```bash
go test ./...
go test -race ./...
```

12. Test rollback path for safety:

```bash
migrate -path internal/store/postgres/migrations -database "$DATABASE_URL" down 1
migrate -path internal/store/postgres/migrations -database "$DATABASE_URL" up
```
