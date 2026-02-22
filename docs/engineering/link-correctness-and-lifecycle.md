# Link Correctness and Lifecycle

## Purpose

Capture the domain rules that make short-link behavior correct, predictable, and safe over time.

## Current implementation

1. Create flow validates `long_url` as `http/https` with host required.
2. Short code is generated with `crypto/rand` and URL-safe encoding.
3. Resolve flow returns:
- success with redirect target when found
- not found when code is missing
4. Redirect behavior uses `302 Found`.

## Current code references

1. `internal/shortener/service.go`
2. `internal/shortener/validate.go`
3. `internal/shortener/code.go`
4. `internal/httpapi/links_handlers.go`

## Phase 2 target decisions

1. Collision handling:
- bounded retries for generated code collisions
- explicit failure mode when retries are exhausted
2. Idempotency:
- define duplicate-create semantics (same input -> same logical result)
- ensure deterministic behavior for client retries
3. Expiry rules:
- add optional `expires_at`
- define behavior for expired link resolution
4. Error taxonomy:
- add domain errors for collision, expiry, and conflict where needed

## Non-negotiable invariants

1. A redirect code maps to at most one active target at a time.
2. Invalid URLs never reach storage.
3. Not-found and expired behavior must be explicit and consistently tested.
4. API behavior for retries must be deterministic.

## Testing direction

1. Table-driven service tests for validation, collision, idempotency, and expiry.
2. Handler tests for HTTP status and error envelope consistency.
3. Add race tests when concurrency is introduced in lifecycle flows.

## Related plans

1. `docs/plan/phase-2-correctness-reliability.md`
2. `docs/plan/phase-3-persistence-migrations.md`
