# Gin Concepts in Use

## Router bootstrap with selected middleware

The project uses `gin.New()` in `internal/httpapi/router.go` and explicitly adds:

1. `gin.Logger()`
2. `gin.Recovery()`

This keeps middleware explicit and easy to reason about.

## Route grouping

Versioned write endpoints are grouped under `/v1`:

1. `POST /v1/links`

Public redirect remains a top-level route:

1. `GET /:code`

## JSON request binding

`Create` handler uses `c.ShouldBindJSON(&req)` with DTO tags:

1. `LongURL string \`json:"long_url" binding:"required"\``
2. `ExpiresAt *time.Time \`json:"expires_at,omitempty"\``

Result: missing `long_url` or malformed JSON becomes `400 bad_request` before domain logic.

## URI binding

Redirect handler uses `c.ShouldBindUri(&uri)` with DTO tag:

`Code string \`uri:"code" binding:"required"\``

This binds `/:code` route params into a typed struct.

## Standardized JSON error payload

`internal/httpapi/errors.go` centralizes error shape:

```json
{
  "error": {
    "code": "string",
    "message": "string"
  }
}
```

All handler errors should use this helper for consistency.

## Status mapping and response patterns

Current response patterns:

1. `c.JSON(...)` for `POST /v1/links` and `GET /health`.
2. `c.Redirect(http.StatusFound, link.LongURL)` for `GET /:code`.

Current create status behavior:

1. `201` when a new link is created.
2. `200` when duplicate create intent returns existing link.

Redirect defaults to `302` (`http.StatusFound`) as project baseline.

## Gin mode decision

`cmd/api/main.go` sets `gin.SetMode(gin.ReleaseMode)` at startup.

This forces release mode unless code is changed.
