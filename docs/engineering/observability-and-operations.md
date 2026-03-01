# Observability and Operations

## Current implementation

1. Gin logger middleware is enabled.
2. Gin recovery middleware is enabled.
3. Health endpoint exists at `GET /health`.
4. Runtime startup logs include:
- selected store backend (memory or Postgres)
- listening port
5. Unexpected handler errors are logged internally with operation + request metadata (method/path) while API responses remain generic.
6. Startup fail-fast behavior for Postgres path:
- app pings DB at startup when `DATABASE_URL` is set
- startup exits on ping failure
7. Local env auto-load behavior:
- app attempts to load `.env` on startup via `godotenv`
- shell/cloud-provided env vars still have precedence
8. Graceful shutdown is implemented with signal handling and timeout-based HTTP shutdown.

## Current code references

1. `internal/httpapi/router.go`
2. `cmd/api/main.go`
3. `internal/httpapi/health.go`
4. `internal/store/postgres/store.go`

## Current gaps

1. No request ID propagation yet.
2. No structured logger abstraction yet.
3. No metrics instrumentation yet.
4. No tracing yet.
5. No protected profiling endpoint setup yet.

## Planned direction

1. Phase 6:
- structured logs
- request ID middleware
- key service metrics
- async analytics with safe backpressure
2. Phase 7:
- profiling, tuning, alerts, runbooks, and release hardening

## Staff-level practice to keep

1. Keep startup and shutdown behavior explicit and deterministic.
2. Add observability before scale pain appears.
3. Keep hot path instrumentation lightweight.
4. Prefer explicit operational behavior over defaults.
