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

2. Domain errors from service are mapped to stable HTTP responses.

3. Create (`POST /v1/links`) mapping:
- `ErrInvalidURL` -> `400 invalid_url`
- `ErrExpired` -> `400 invalid_expiry`
- `ErrCollision` -> `409 conflict`
- unknown errors -> `500 internal`

4. Resolve (`GET /:code`) mapping:
- `ErrNotFound` -> `404 not_found`
- `ErrExpired` -> `404 not_found` (intentional non-disclosure)
- unknown errors -> `500 internal`

5. Binder failures still return `400 bad_request` with binder message.
6. Unexpected handler/store errors are logged internally with operation + request metadata while client responses remain generic.

6. Common JSON formatting is centralized via `jsonError`.

## Why this matters

1. Consistent error shape keeps client handling simple.
2. Domain-to-HTTP mapping keeps transport concerns out of business logic.
3. Centralized formatting reduces drift across handlers.
4. Expired links remain non-enumerable from HTTP responses.

## Current code references

1. `internal/httpapi/errors.go`
2. `internal/httpapi/links_handlers.go`
3. `internal/shortener/service.go`

## Staff-level practice to keep

1. Keep domain errors transport-agnostic.
2. Keep HTTP mapping stable and explicit.
3. Avoid leaking internals in `5xx` responses.
4. Log unexpected failures with enough context for debugging, without exposing sensitive data to clients.
