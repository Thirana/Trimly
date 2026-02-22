# ADR-0004: Validation at HTTP Boundary and Service Layer

## Status

`accepted`

## Date

2026-02-22

## Context

Single-layer validation often fails over time as new entry points are added (new handlers, tests, jobs, internal calls).
The project needs both transport-level and domain-level safeguards.

## Decision

Use defense-in-depth validation:

1. Validate request shapes at HTTP boundary using Gin binding and DTOs.
2. Validate critical domain invariants in `shortener.Service` (URL validity and future lifecycle rules).

## Consequences

1. Reduced risk of invalid state entering storage.
2. Better resilience when transport code evolves.
3. Slight duplication cost that is acceptable for correctness.

## Alternatives considered

1. HTTP-only validation.
2. Service-only validation.
3. Validation in store layer only.

## Implementation notes

1. DTO and binding: `internal/httpapi/dto.go`, `internal/httpapi/links_handlers.go`
2. Domain URL validation: `internal/shortener/validate.go`, `internal/shortener/service.go`
3. Future extension target: collision/idempotency/expiry in Phase 2.
