# Database Integrity and Consistency

## Current implementation status

1. `LinkStore` interface decouples business logic from storage backend.
2. `MemoryStore` remains available for local/no-DB runs.
3. Postgres store skeleton now exists in `internal/store/postgres`.
4. Initial SQL migration scaffold exists in `internal/store/postgres/migrations`.
5. Runtime store selection is environment-driven:
- `DATABASE_URL` set -> Postgres path
- `DATABASE_URL` unset -> in-memory path

## Why this is a solid progression

1. Interface-first design keeps service logic stable while backends evolve.
2. Memory fallback preserves fast local iteration when DB is not required.
3. Postgres path introduces durable constraints and persistence readiness.
4. Versioned migrations establish reproducible schema evolution.

## Current code references

1. `internal/store/link_store.go`
2. `internal/store/intent.go`
3. `internal/store/memory_store.go`
4. `internal/store/postgres/store.go`
5. `internal/store/postgres/migrations/000001_create_links_table.up.sql`
6. `internal/store/postgres/migrations/000001_create_links_table.down.sql`
7. `cmd/api/main.go`

## Integrity guarantees now covered

1. App-layer idempotency semantics shared across store implementations via `BuildIntentKey`.
2. Postgres schema constraints include:
- unique `code`
- unique `intent_key`
3. Store-level duplicate writes map to explicit app errors:
- `ErrCodeExists`
- `ErrIntentExists`

## Remaining Phase 3 work

1. Add integration tests against disposable Postgres.
2. Add migration verification tests (`up` from empty DB and rollback safety).
3. Decide when to make Postgres the default runtime path for all environments.
4. Add stronger update semantics (`updated_at` maintenance) when write/update flows expand.

## Staff-level practice to keep

1. Put invariants in DB constraints, not only in application code.
2. Keep repository interfaces stable while swapping implementations.
3. Treat migrations as first-class, versioned code.
