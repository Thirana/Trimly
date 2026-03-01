# Local Development Setup

## Prerequisites

1. Go toolchain installed.
2. Version compatible with `go.mod` (`go 1.25.0`).
3. Optional but recommended for Phase 3 DB path:
- `psql` CLI
- `migrate` CLI (`golang-migrate`)

## Install dependencies

```bash
go mod download
```

## Local `.env` setup

Create `.env` in repo root:

```bash
cat > .env <<'EOF'
PORT=8080
BASE_URL=http://localhost:8080
DATABASE_URL=postgres://<user>:<password>@<host>/<db>?sslmode=require
REDIS_URL=rediss://default:password@<host>:6379
REDIS_ENABLED=false
REDIS_POSITIVE_TTL=10m
REDIS_MISS_TTL=45s
REDIS_CONNECT_TIMEOUT=3s
REDIS_OP_TIMEOUT=150ms
CACHE_METRICS_LOG_INTERVAL=30s
EOF
```

Notes:

1. App auto-loads `.env` on startup using `github.com/joho/godotenv`.
2. Existing environment variables already set in the shell take precedence over `.env`.

## Run the API (in-memory mode)

```bash
go run ./cmd/api
```

Server defaults:

1. Listens on `:8080` when `PORT` is not set.
2. Uses in-memory storage when `DATABASE_URL` is not set.

## Run the API (Postgres mode)

```bash
DATABASE_URL='postgres://<user>:<password>@<host>/<db>?sslmode=require' \
BASE_URL='http://localhost:8080' \
PORT=8080 \
go run ./cmd/api
```

Behavior:

1. App chooses Postgres store when `DATABASE_URL` is set.
2. Startup pings DB and fails fast if unreachable.
3. You can use `.env` without running `source` manually.

## Migration workflow (Phase 3)

1. Install migration CLI:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

2. Apply migrations:

```bash
migrate -path internal/store/postgres/migrations -database "$DATABASE_URL" up
```

3. Roll back one migration:

```bash
migrate -path internal/store/postgres/migrations -database "$DATABASE_URL" down 1
```

## Environment variables

1. `PORT`
- Purpose: HTTP port binding for the API process.
- Default: `8080`
- Example: `PORT=9090`

2. `BASE_URL`
- Purpose: base URL used to build `short_url` in `POST /v1/links` response.
- Default: `http://localhost:8080`
- Example: `BASE_URL=http://localhost:9090`

3. `DATABASE_URL`
- Purpose: enables Postgres store path when set.
- Example: `postgres://user:pass@host:5432/dbname?sslmode=require`

4. `REDIS_URL`
- Purpose: Redis connection URL (Upstash `rediss://...`) used when cache is enabled.

5. `REDIS_ENABLED`
- Purpose: feature flag for Redis cache path.
- Default: `false`
- Rule: when `true`, `REDIS_URL` must be set and reachable; startup validates with Redis `PING`.

6. `REDIS_POSITIVE_TTL`
- Purpose: TTL for positive cache keys.
- Default: `10m`

7. `REDIS_MISS_TTL`
- Purpose: TTL for negative cache keys.
- Default: `45s`

8. `REDIS_CONNECT_TIMEOUT`
- Purpose: startup connectivity timeout for Redis `PING`.
- Default: `3s`

9. `REDIS_OP_TIMEOUT`
- Purpose: per-request Redis operation timeout.
- Default: `150ms`
- Backward compatibility: app will use `REDIS_TIMEOUT` when `REDIS_OP_TIMEOUT` is unset.

10. `CACHE_METRICS_LOG_INTERVAL`
- Purpose: interval for periodic cache metric logs from startup process.
- Default: `30s`
- Set `0` to disable.

11. `GIN_MODE`
- Note: current code calls `gin.SetMode(gin.ReleaseMode)` on startup, so release mode is forced by code.
- Practical impact: setting `GIN_MODE` alone will not switch runtime mode unless code path changes.

## Quick health check

```bash
curl -i http://localhost:8080/health
```

Expected:

1. `200 OK`
2. JSON body: `{"status":"ok"}`
