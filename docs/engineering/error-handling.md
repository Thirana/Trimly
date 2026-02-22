# Error Handling

## Current implementation

1. API errors use a consistent JSON envelope:

```json
{
  "error": {
    "code": "string",
    "message": "string"
  }
}
```

2. Domain errors from service are mapped to HTTP status:
- `ErrInvalidURL` -> `400 invalid_url`
- `ErrNotFound` -> `404 not_found`
- unknown errors -> `500 internal`
3. Common formatting is centralized via `jsonError`.

## Why this matters

1. Consistent error shape keeps client handling simple.
2. Domain-to-HTTP mapping keeps transport concerns out of business logic.
3. Centralized formatting reduces drift across handlers.

## Current code references

1. `internal/httpapi/errors.go`
2. `internal/httpapi/links_handlers.go`
3. `internal/shortener/service.go`

## Gaps and next steps

1. Some bad-request paths still return raw binder errors directly.
2. There is no typed error taxonomy beyond a small set of sentinel errors.
3. Add more explicit domain error categories in Phase 2:
- collision
- expiry
- conflict/idempotency behavior

## Staff-level practice to keep

1. Keep domain errors transport-agnostic.
2. Keep HTTP mapping stable and explicit.
3. Avoid leaking internals in `5xx` responses.
