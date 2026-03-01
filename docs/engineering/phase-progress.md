# Phase Progress

This file tracks implementation status against the phase roadmap.
Update it whenever phase scope or completion status changes.

## Legend

1. `done`: phase outcomes implemented and verified.
2. `in_progress`: currently being implemented.
3. `planned`: not started.
4. `blocked`: cannot proceed until dependencies are resolved.

## Phase status

1. Phase 2 (correctness + reliability): `done`
2. Phase 3 (persistence + migrations): `done`
3. Phase 4 (Redis hot path + load tests): `done`
4. Phase 5 (auth + profiles + link management): `planned`
5. Phase 6 (abuse protection + observability + async analytics): `planned`
6. Phase 7 (performance engineering + deployment hardening): `planned`

## Phase 2 changelog

1. 2026-02-26:
- Added collision retries with bounded attempts and explicit `ErrCollision`.
- Added idempotent create behavior by normalized intent (`long_url` + optional `expires_at`).
- Added optional expiry semantics and expired-link handling.
- Added domain error mapping updates (`invalid_expiry`, `conflict`, expired-as-not-found).
- Added service and handler tests for collision, idempotency, expiry, and error envelopes.
- Added race test execution as part of verification.

## Phase 3 changelog

1. 2026-02-28 (scaffold milestone):
- Added Postgres store skeleton (`internal/store/postgres`) implementing `LinkStore`.
- Added initial migration scaffold for links table and required indexes.
- Added `DATABASE_URL`-driven runtime store selection (Postgres or in-memory fallback).
- Added startup DB ping fail-fast and graceful shutdown wiring in `cmd/api/main.go`.
- Refactored router wiring so dependency composition happens in `cmd/api`.
2. 2026-03-01 (completion milestone):
- Added Postgres integration tests for `Save`, `Get`, `FindByIntent`, and unique constraint error mapping.
- Added migration verification integration test flow (`up` from empty schema, `down`, then `up` again).
- Added safe internal handler logging for unexpected errors while keeping generic client `500` responses.
- Phase 3 exit criteria are now covered by tests and documented verification flow.

## Phase 4 changelog

1. 2026-03-01 (config scaffold milestone):
- Added typed Redis env config parsing in `cmd/api/main.go` for:
  - `REDIS_ENABLED`
  - `REDIS_POSITIVE_TTL`
  - `REDIS_MISS_TTL`
  - `REDIS_CONNECT_TIMEOUT`
  - `REDIS_OP_TIMEOUT` (with fallback to legacy `REDIS_TIMEOUT`)
- Added startup validation and fail-fast behavior for invalid cache config.
- Added unit tests for Redis env parsing behavior.
2. 2026-03-01 (startup connectivity milestone):
- Added startup Redis connectivity check (`REDIS_ENABLED=true` -> parse URL + `PING`).
- Added explicit Postgres and Redis connectivity-pass startup logs.
- Added startup tests for disabled Redis path and invalid Redis URL handling.
3. 2026-03-01 (cache-aside core flow milestone):
- Added service-level cache abstraction (`ResolveCache`) with no-op fallback.
- Added Redis cache adapter implementing:
  - `v1:url:short:{code}`
  - `v1:url:miss:{code}`
- Wired cache into service resolve path (`GET short` -> `GET miss` -> DB fallback).
- Added create-path miss-key invalidation and positive-key warming.
- Added service tests for cache-hit, miss-hit, DB fallback, TTL clamping, and create-path cache updates.
4. 2026-03-01 (observability + baseline load milestone):
- Added atomic resolve-path cache counters:
  - short-key hits
  - miss-key hits
  - DB fallbacks/hits/misses
  - cache errors
- Added periodic cache metrics logs controlled by `CACHE_METRICS_LOG_INTERVAL`.
- Added k6 redirect baseline script and setup runbook (`scripts/load/redirect_baseline.js`, `docs/setup/load-testing.md`).
5. 2026-03-01 (completion milestone):
- Executed Redis-off vs Redis-on redirect load tests with identical parameters (`VUS=20`, `DURATION=20s`).
- Recorded measured throughput/latency improvements in `docs/engineering/performance-and-load-engineering.md`.
- Verified rollback path by running full workload with `REDIS_ENABLED=false`.
- Phase 4 exit criteria are now satisfied and documented.

## What is implemented today

1. Gin API with:
- `GET /health`
- `POST /v1/links`
- `GET /:code`
2. Storage implementations:
- in-memory store
- Postgres store (enabled when `DATABASE_URL` is set)
- Redis resolve-cache adapter (enabled when `REDIS_ENABLED=true`)
3. URL validation + normalization in service layer.
4. Stable JSON error envelope and domain error mapping.
5. Random URL-safe short-code generation with collision retries.
6. Idempotent create semantics.
7. Optional expiry (`expires_at`) and expired redirect protection.
8. Runtime graceful shutdown with timeout.

## How to update this file

1. Change phase status as work progresses.
2. Add a short changelog entry under the phase when major milestones complete.
3. Keep references to relevant docs/plan files.
