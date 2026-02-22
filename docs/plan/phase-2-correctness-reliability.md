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

## Implementation approach

1. Define domain errors in `internal/shortener`:
- `ErrInvalidURL`
- `ErrNotFound`
- `ErrExpired`
- `ErrCollision`
2. Update short-code generation flow:
- generate candidate code
- check uniqueness
- retry with bounded attempts
- return explicit collision error when exhausted
3. Add idempotency strategy for create endpoint:
- use normalized input (`url`, optional custom alias, owner context)
- for duplicate intent, return existing link instead of creating another
4. Add expiry validation:
- reject already-expired create requests
- treat expired links as not redirectable and return consistent status
5. Keep handler logic thin and centralize mapping from domain errors to HTTP errors.

## Data and API notes

1. Add optional `expires_at` field in create request/response if not already present.
2. Decide and document response for expired links (commonly `404` to avoid information leaks).
3. Keep redirect status at `302` for now.

## Testing plan

1. Unit tests:
- URL validation edge cases
- collision retry behavior
- idempotency cases
- expiry decision logic
2. Handler tests (`httptest`):
- create success
- duplicate create idempotent response
- invalid input and error shapes
- expired redirect handling
3. Add race test only if new concurrent paths are introduced.

## Risks and mitigations

1. Ambiguous idempotency semantics:
- write explicit rules in code comments and docs.
2. Time-based flaky tests:
- inject clock/time provider instead of using direct `time.Now()` in logic.
3. Collision loops under stress:
- use bounded retries and clear metrics/log fields for failures.

## Exit criteria

1. Deterministic collision behavior with tests.
2. Idempotent create behavior defined and tested.
3. Expiry behavior consistent across service and handlers.
4. Error mapping stable and documented.
5. `gofmt -w .` and `go test ./...` pass.
