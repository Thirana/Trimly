# ADR-0005: Phase 2 Correctness Contract (Collision, Idempotency, Expiry)

## Status

`accepted`

## Date

2026-02-26

## Context

The project needs deterministic behavior for client retries and reliable lifecycle rules before introducing DB/Redis complexity.
Without explicit contracts, duplicate creates, collisions, and expiry behavior drift quickly across handlers and stores.

## Decision

1. Create-link requests are idempotent by normalized intent:
- normalized `long_url`
- normalized optional `expires_at`
2. Duplicate create intent returns existing link payload with `200 OK`.
3. New create returns `201 Created`.
4. `expires_at` is optional and must be in the future.
5. Expired links are not redirectable and are returned as `404 not_found` (same as missing code).
6. Short-code allocation retries collisions with bounded attempts (`maxAttempts = 5`).
7. Collision exhaustion maps to domain `ErrCollision` and HTTP `409 conflict`.

## Consequences

1. Client retries for the same intent become deterministic.
2. Redirect endpoint avoids revealing whether a code exists but is expired.
3. Collision failure mode is explicit and testable.
4. Store contract expands to support idempotency lookup and uniqueness enforcement.

## Alternatives considered

1. Always create a new code for duplicate requests.
2. Treat expired redirects differently from not-found (for example `410`).
3. Unbounded collision retries.
4. Idempotency keys only at HTTP layer.

## Implementation notes

1. Domain logic: `internal/shortener/service.go`, `internal/shortener/validate.go`
2. Store contract + memory behavior: `internal/store/link_store.go`, `internal/store/memory_store.go`
3. HTTP mapping + DTOs: `internal/httpapi/dto.go`, `internal/httpapi/errors.go`, `internal/httpapi/links_handlers.go`
4. Tests: `internal/shortener/service_test.go`, `internal/httpapi/links_handlers_test.go`
