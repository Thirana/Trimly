# Go Concepts in Use

## Package boundaries with `internal/`

The repo uses `internal/` packages to enforce architecture boundaries:

1. `internal/httpapi` for HTTP handlers and DTO binding.
2. `internal/shortener` for business logic.
3. `internal/store` for persistence interfaces and implementations.

This prevents external modules from importing these packages directly.

## Implicit interface implementation

`internal/store/link_store.go` defines `LinkStore`.
Both store implementations satisfy it without an `implements` keyword:

1. `*MemoryStore` in `internal/store/memory_store.go`
2. `*postgres.Store` in `internal/store/postgres/store.go`

Key benefit: service code depends on behavior (`LinkStore`), not concrete backend.

## Composition root in `cmd/api`

Dependency composition now happens in `cmd/api/main.go`:

1. choose store implementation based on `DATABASE_URL`
2. construct `shortener.Service`
3. construct `httpapi.LinksHandler`
4. pass handler into `httpapi.NewRouter(links)`

This keeps HTTP routing separate from backend selection.

## Environment bootstrapping with `godotenv`

`cmd/api/main.go` loads `.env` via `godotenv.Load(".env")` at startup for local runs.

Behavior:

1. Missing `.env` is non-fatal.
2. Existing process environment variables keep precedence.
3. Cloud deployments can rely on provider-managed env vars without `.env` files.

## Pointer receiver methods

`Service`, `MemoryStore`, and `postgres.Store` methods use pointer receivers.

Why it matters:

1. Avoids copying structs.
2. Allows internal mutation/resource management.
3. Ensures method sets align with injected pointer values.

## `context.Context` propagation

Handlers pass request context into the service:

1. `h.svc.Create(c.Request.Context(), req.LongURL, req.ExpiresAt)`
2. `h.svc.Resolve(c.Request.Context(), uri.Code)`

Service/store APIs accept `context.Context` so cancellation and deadlines can be enforced consistently.

## Sentinel errors and HTTP mapping

Service-level sentinel errors:

1. `shortener.ErrInvalidURL`
2. `shortener.ErrNotFound`
3. `shortener.ErrExpired`
4. `shortener.ErrCollision`

Store-level sentinel errors:

1. `store.ErrCodeExists`
2. `store.ErrIntentExists`

Handlers map domain errors to stable HTTP responses.

## URL normalization and shared intent keys

Idempotency relies on normalized URL + normalized expiry.
`internal/store/intent.go` provides shared `BuildIntentKey` so all backends use the same key shape.

## Deterministic retry and injectable dependencies

Service uses bounded collision retries and injectable functions for testability:

1. `newCode` function field for deterministic code-generation tests.
2. `now` function field for deterministic expiry tests.
3. bounded `maxAttempts` for collision exhaustion behavior.

## Concurrency safety across stores

1. In-memory store uses `sync.RWMutex` around maps.
2. Postgres store relies on DB constraints and transactional write behavior.
