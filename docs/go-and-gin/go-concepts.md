# Go Concepts in Use

## Package boundaries with `internal/`

The repo uses `internal/` packages to enforce architecture boundaries:

1. `internal/httpapi` for HTTP handlers and DTO binding.
2. `internal/shortener` for business logic.
3. `internal/store` for persistence interfaces and implementations.

This prevents external modules from importing these packages directly.

## Implicit interface implementation

`internal/store/link_store.go` defines `LinkStore`.
`*MemoryStore` in `internal/store/memory_store.go` implements it without an `implements` keyword.

Key benefit: the service depends on behavior (`LinkStore`), not storage concrete type.

## Constructor-style dependency injection

`internal/httpapi/router.go` wires dependencies manually:

1. `mem := store.NewMemoryStore()`
2. `svc := shortener.NewService(mem)`
3. `links := NewLinksHandler(svc)`

This keeps router wiring explicit and testable.

## Pointer receiver methods

`Service` and `MemoryStore` methods use pointer receivers (`func (s *Service) ...`, `func (m *MemoryStore) ...`).

Why it matters:

1. Avoids copying structs.
2. Allows internal mutation (`MemoryStore` map + mutex).
3. Ensures method sets align with injected pointer values.

## `context.Context` propagation

Handlers pass request context into the service:

1. `h.svc.Create(c.Request.Context(), req.LongURL, req.ExpiresAt)`
2. `h.svc.Resolve(c.Request.Context(), uri.Code)`

Service/store APIs accept `context.Context` so cancellation and deadlines can be enforced later without redesign.

## Sentinel errors and HTTP mapping

Service-level sentinel errors:

1. `shortener.ErrInvalidURL`
2. `shortener.ErrNotFound`
3. `shortener.ErrExpired`
4. `shortener.ErrCollision`

Handlers map these to stable HTTP responses.

## URL normalization for idempotency

Service normalizes URLs using `net/url` before persistence and idempotency checks.

Current normalization includes:

1. trim spaces
2. enforce `http/https` + host
3. lowercase scheme and host
4. remove default ports (`80`/`443`)
5. normalize empty path to `/`

## Deterministic retry and injectable dependencies

Service uses bounded collision retries and injectable functions for testability:

1. `newCode` function field for deterministic code-generation tests.
2. `now` function field for deterministic expiry tests.
3. bounded `maxAttempts` for collision exhaustion behavior.

## Concurrency-safe in-memory store

`MemoryStore` protects in-memory maps with `sync.RWMutex` and tracks:

1. `code -> link`
2. `intentKey -> code`

This enforces code uniqueness and create-intent idempotency safely under concurrent access.
