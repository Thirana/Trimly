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

## Integration tests with conditional execution

`internal/store/postgres/store_integration_test.go` uses environment-aware integration tests:

1. Tests read `DATABASE_URL`.
2. If not set, tests call `t.Skip(...)` instead of failing unit-test runs.
3. This keeps `go test ./...` usable for contributors without a DB.

## Schema-isolated Postgres integration testing

Postgres integration tests create a temporary schema per test and set `search_path` on connections.

Why this pattern:

1. Avoids polluting shared tables.
2. Allows migration verification (`up` -> `down` -> `up`) safely.
3. Keeps tests repeatable even against a managed DB target.

## Fail-fast env parsing with typed defaults

`cmd/api/main.go` parses Redis runtime env values at startup into typed config:

1. booleans via `strconv.ParseBool` (`REDIS_ENABLED`)
2. durations via `time.ParseDuration` (`REDIS_POSITIVE_TTL`, `REDIS_MISS_TTL`, `REDIS_CONNECT_TIMEOUT`, `REDIS_OP_TIMEOUT`)
3. explicit validation (`> 0` durations, `REDIS_URL` required when enabled)

Why this pattern:

1. catches bad env values at startup instead of runtime hot path
2. keeps behavior tunable without code changes
3. preserves safe rollback by disabling cache through config

## Startup connectivity fail-fast checks

`cmd/api/main.go` validates external dependencies before serving traffic:

1. Postgres path: pool creation + startup `Ping`.
2. Redis path (when enabled): URL parse + startup `PING`.

Why this pattern:

1. invalid URLs fail at boot, not on first request
2. misconfigured credentials are detected early
3. startup logs clearly show dependency readiness

## Interface-based cache injection (no-op default)

`internal/shortener/service.go` now depends on a cache interface (`ResolveCache`) with a no-op default.

Why this pattern:

1. service logic can use caching without hard-coding Redis dependency
2. local/in-memory runs work without cache infrastructure
3. tests can inject fake caches to verify hit/miss/fallback behavior deterministically

## Cache-aside in service layer

Resolve flow is implemented in domain/service code instead of HTTP layer:

1. read `short` key
2. read `miss` key
3. fallback to store
4. write back to cache

Why this pattern:

1. keeps HTTP handlers thin
2. keeps caching policy consistent across all callers of resolve logic
3. makes key/TTL semantics testable at service level

## Low-overhead counters with `sync/atomic`

`internal/shortener/service.go` tracks resolve-path counters using atomic integers.

Why this pattern:

1. concurrency-safe increments without mutex contention on hot path
2. cheap enough for redirect-path observability
3. supports periodic metrics logging without adding external telemetry dependency
