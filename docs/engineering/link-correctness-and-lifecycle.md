# Link Correctness and Lifecycle

## Purpose

Capture the domain rules that make short-link behavior correct, predictable, and safe over time.

## Current implementation

1. Create validates and normalizes `long_url` as `http/https` with host required.
2. Optional `expires_at` is accepted on create.
3. Create rejects non-future `expires_at`.
4. Create is idempotent by normalized intent key:
- normalized `long_url`
- normalized `expires_at` (or no-expiry)
5. Short code allocation retries collisions with bounded attempts (`5`).
6. Resolve returns:
- redirect target on active link
- not found for missing code
- not found for expired code (via domain `ErrExpired`)
7. Redis cache-aside is active when enabled:
- short-key hit resolves without DB lookup
- miss-key hit resolves as not found without DB lookup
- DB remains source of truth on cache misses/errors
8. Redirect behavior uses `302 Found`.

## Current code references

1. `internal/shortener/service.go`
2. `internal/shortener/validate.go`
3. `internal/shortener/code.go`
4. `internal/store/link_store.go`
5. `internal/store/memory_store.go`
6. `internal/httpapi/links_handlers.go`

## Domain error taxonomy

1. `ErrInvalidURL`
2. `ErrNotFound`
3. `ErrExpired`
4. `ErrCollision`

## Non-negotiable invariants

1. Invalid URLs never reach storage.
2. A create request with the same normalized intent returns the same logical link.
3. A code is not overwritten on collision; collisions retry with bounded attempts.
4. Expired links are not redirectable.
5. Missing and expired redirects are indistinguishable at HTTP level (`404 not_found`).
6. Cache must never violate lifecycle rules:
- expired links are never redirectable (even on cache hit)
- stale miss keys are cleared on create for active codes

## Testing coverage

1. Service tests cover URL normalization, idempotency, collision retries/exhaustion, and expiry handling.
2. Handler tests cover create success, duplicate create behavior, invalid input envelope, and expired redirect behavior.
3. Race tests run for updated concurrent store behavior.

## Related plans

1. `docs/plan/phase-2-correctness-reliability.md`
2. `docs/plan/phase-3-persistence-migrations.md`
