# Validation and Types

## Current implementation

1. HTTP boundary validation uses Gin binding on DTO structs.
2. `CreateLinkRequest` requires `long_url`.
3. URI binding is used for redirect code extraction.
4. Service-level URL validation is enforced with `IsValidHTTPURL`.

## Why this is a good baseline

1. Boundary validation prevents malformed request payloads from entering business logic.
2. Service validation gives defense-in-depth if handlers change or other entry points are added later.
3. DTO structs make API contract explicit and easy to evolve.

## Current code references

1. `internal/httpapi/dto.go`
2. `internal/httpapi/links_handlers.go`
3. `internal/shortener/validate.go`

## Gaps and next steps

1. Validation errors currently surface raw binder messages in some paths.
2. There are no explicit field-level error details yet.
3. Add stronger URL policy decisions in Phase 2:
- optional normalization
- max URL length limits
- expiry field validation when introduced

## Staff-level practice to keep

1. Validate early at transport boundary.
2. Re-check critical domain invariants in service layer.
3. Keep validation rules centralized and deterministic.
