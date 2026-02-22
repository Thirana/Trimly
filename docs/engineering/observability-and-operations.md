# Observability and Operations

## Current implementation

1. Gin logger middleware is enabled.
2. Gin recovery middleware is enabled.
3. Startup logs include listening port.
4. Health endpoint exists at `GET /health`.

## Current code references

1. `internal/httpapi/router.go`
2. `cmd/api/main.go`
3. `internal/httpapi/health.go`

## Current gaps

1. No request ID propagation yet.
2. No structured logger abstraction yet.
3. No metrics instrumentation yet.
4. No graceful shutdown orchestration yet.
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

1. Add observability before scale pain appears.
2. Keep hot path instrumentation lightweight.
3. Prefer explicit operational behavior over defaults.
