# API Endpoints

Base URL examples in this file use `http://localhost:8080`.

## Error response contract

All error responses follow this shape:

```json
{
  "error": {
    "code": "string",
    "message": "string"
  }
}
```

## `GET /health`

Purpose: lightweight service liveness endpoint.

Request body: none.

Example:

```bash
curl -i http://localhost:8080/health
```

Success response:

1. Status: `200`
2. Body:

```json
{
  "status": "ok"
}
```

## `POST /v1/links`

Purpose: create or return an idempotent short link for a long URL.

Request body:

```json
{
  "long_url": "https://example.com/some/path",
  "expires_at": "2026-03-01T10:00:00Z"
}
```

Notes:

1. `long_url` is required.
2. `expires_at` is optional and must be a future RFC3339 timestamp.

Example:

```bash
curl -i -X POST http://localhost:8080/v1/links \
  -H 'Content-Type: application/json' \
  -d '{"long_url":"https://example.com/some/path","expires_at":"2026-03-01T10:00:00Z"}'
```

Success response:

1. Status:
- `201` when a new short link is created.
- `200` when request is a duplicate intent and existing link is returned.
2. Body:

```json
{
  "code": "abc123X",
  "short_url": "http://localhost:8080/abc123X",
  "long_url": "https://example.com/some/path",
  "expires_at": "2026-03-01T10:00:00Z"
}
```

Common errors:

1. `400 bad_request` when body is invalid JSON or missing required fields.
2. `400 invalid_url` when `long_url` is not a valid `http/https` URL.
3. `400 invalid_expiry` when `expires_at` is not in the future.
4. `409 conflict` when code allocation collisions are exhausted.
5. `500 internal` for unexpected failures.

## `GET /:code`

Purpose: resolve short code and redirect to original long URL.

Request body: none.

Example:

```bash
curl -i http://localhost:8080/abc123X
```

Success response:

1. Status: `302`
2. Header: `Location: <long_url>`

Common errors:

1. `400 bad_request` when path param binding fails.
2. `404 not_found` when code does not exist or has expired.
3. `500 internal` for unexpected failures.

## Suggested manual test flow

1. Call `POST /v1/links` with a valid `long_url`.
2. Repeat the same call and confirm second response is `200` with same `code`.
3. Call `GET /{code}` and verify `302` redirect target.
4. Create with past `expires_at` and confirm `400 invalid_expiry`.
