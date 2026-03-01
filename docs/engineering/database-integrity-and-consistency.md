# Database Integrity and Consistency

## Current implementation status

1. `LinkStore` interface decouples business logic from storage backend.
2. `MemoryStore` remains available for local/no-DB runs.
3. Postgres store implementation exists in `internal/store/postgres`.
4. Versioned SQL migrations exist in `internal/store/postgres/migrations`.
5. Runtime store selection is environment-driven:
- `DATABASE_URL` set -> Postgres path
- `DATABASE_URL` unset -> in-memory path
6. Postgres integration tests now cover:
- `Save`, `Get`, `FindByIntent`
- duplicate code/intent unique constraint mapping to domain errors
- migration verification flow (`up` -> `down` -> `up`)
7. Redis cache adapter exists for resolve-path optimization:
- `internal/store/rediscache/cache.go`
- DB lookup remains authoritative on cache misses/errors

## Why this is a solid progression

1. Interface-first design keeps service logic stable while backends evolve.
2. Memory fallback preserves fast local iteration when DB is not required.
3. Postgres path introduces durable constraints and persistence readiness.
4. Versioned migrations establish reproducible schema evolution.
5. Cache remains optional/replaceable via service-level cache interface.

## Current code references

1. `internal/store/link_store.go`
2. `internal/store/intent.go`
3. `internal/store/memory_store.go`
4. `internal/store/postgres/store.go`
5. `internal/store/postgres/migrations/000001_create_links_table.up.sql`
6. `internal/store/postgres/migrations/000001_create_links_table.down.sql`
7. `internal/store/postgres/store_integration_test.go`
8. `internal/store/rediscache/cache.go`
9. `cmd/api/main.go`

## Integrity guarantees now covered

1. App-layer idempotency semantics shared across store implementations via `BuildIntentKey`.
2. Postgres schema constraints include:
- unique `code`
- unique `intent_key`
3. Store-level duplicate writes map to explicit app errors:
- `ErrCodeExists`
- `ErrIntentExists`
4. Migration files are verified in tests against an empty isolated schema.
5. Migration rollback safety is verified via `down` then `up` test flow.
6. Cache does not redefine persistence truth:
- create path clears stale miss keys and warms short keys
- resolve path falls back to DB when cache does not provide a valid answer
7. Postgres store queries use schema-qualified relations (`public.links`) to avoid `search_path` drift issues when using managed poolers.

## Remaining persistence work beyond Phase 3

1. Decide when to make Postgres the default runtime path for all environments.
2. Add stronger update semantics (`updated_at` maintenance) when write/update flows expand.

## Staff-level practice to keep

1. Put invariants in DB constraints, not only in application code.
2. Keep repository interfaces stable while swapping implementations.
3. Treat migrations as first-class, versioned code.
