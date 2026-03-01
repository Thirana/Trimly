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
3. Phase 4 (Redis hot path + load tests): `planned`
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

## What is implemented today

1. Gin API with:
- `GET /health`
- `POST /v1/links`
- `GET /:code`
2. Storage implementations:
- in-memory store
- Postgres store skeleton (enabled when `DATABASE_URL` is set)
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
