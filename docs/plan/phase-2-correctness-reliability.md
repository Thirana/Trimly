# Phase 2: Correctness and Reliability

## Goal

Improve correctness guarantees before adding infrastructure complexity.
Focus on collision handling, idempotency behavior, expiry rules, and stronger tests.

## Scope

1. Short-code collision handling with deterministic retries.
2. Idempotent create-link behavior for duplicate requests.
3. Expiry semantics (`expires_at`) and redirect behavior for expired links.
4. Clear domain errors mapped to stable HTTP responses.
5. Broader unit and handler test coverage.

## Locked decisions (2026-02-26)

1. `expires_at` is optional RFC3339 in create request/response.
2. Expired redirects return `404 not_found` (same as missing code).
3. Idempotency key is normalized `long_url` + optional normalized `expires_at`.
4. Duplicate create returns existing link payload with `200 OK`.
5. Collision retries are bounded to `5`; exhaustion maps to `409 conflict`.

## Implementation approach

1. Define domain errors in `internal/shortener`:
- `ErrInvalidURL`
- `ErrNotFound`
- `ErrExpired`
- `ErrCollision`
2. Update short-code generation flow:
- generate candidate code
- rely on store uniqueness checks
- retry with bounded attempts
- return explicit collision error when exhausted
3. Add idempotency strategy for create endpoint:
- normalize input URL
- normalize optional expiry
- return existing link for duplicate intent
4. Add expiry validation:
- reject already-expired create requests
- treat expired links as not redirectable
5. Keep handler logic thin and centralize mapping from domain errors to HTTP errors.

## Data and API notes

1. Create request/response includes optional `expires_at`.
2. Redirect status remains `302`.
3. Error envelope remains stable across new mappings.

## Testing plan

1. Unit tests:
- URL normalization and validation edge cases
- collision retry behavior
- collision exhaustion behavior
- idempotency behavior
- expiry decision logic
2. Handler tests (`httptest`):
- create success (`201`)
- duplicate create idempotent response (`200`)
- invalid input and error shapes
- expired redirect handling (`404`)
3. Run race tests because store concurrency behavior was touched.

## Risks and mitigations

1. Ambiguous idempotency semantics:
- keep rules explicit in code comments and docs.
2. Time-based flaky tests:
- service uses injectable clock for deterministic tests.
3. Collision loops under stress:
- use bounded retries and explicit failure error.

## Exit criteria

1. Deterministic collision behavior with tests.
2. Idempotent create behavior defined and tested.
3. Expiry behavior consistent across service and handlers.
4. Error mapping stable and documented.
5. `gofmt -w .`, `go test ./...`, and `go test -race ./...` pass.
