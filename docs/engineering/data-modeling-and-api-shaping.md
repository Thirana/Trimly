# Data Modeling and API Shaping

## Current model and API shape

1. Domain link model contains `Code` and `LongURL`.
2. Create response includes:
- `code`
- `short_url`
- `long_url`
3. Redirect path is public and code-based: `GET /:code`.
4. Versioned API namespace is used for create: `POST /v1/links`.

## Why these choices are useful now

1. Minimal model keeps early complexity low and iteration fast.
2. Response includes both canonical short and long values, which is useful for clients.
3. Versioned creation API allows future evolution without breaking clients.
4. Public redirect path stays simple and cache-friendly.

## Current code references

1. `internal/store/link_store.go`
2. `internal/httpapi/dto.go`
3. `internal/httpapi/router.go`
4. `internal/httpapi/links_handlers.go`

## Planned model evolution

1. Add persistent fields in Phase 3:
- `id`
- `created_at`
- `updated_at`
- optional `expires_at`
- optional ownership fields (`user_id`)
2. Add idempotency and collision semantics in Phase 2.
3. Add owner-facing management APIs in Phase 5.

## Staff-level practice to keep

1. Keep API contracts explicit through DTOs.
2. Design for backward-compatible API evolution.
3. Keep hot path payloads and data model lean.
