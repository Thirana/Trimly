# Database Integrity and Consistency

## Current implementation status

1. Persistence is currently in-memory only.
2. `LinkStore` interface decouples business logic from storage backend.
3. `MemoryStore` uses `sync.RWMutex` and map for thread-safe access.
4. Writes are simple overwrite-by-code semantics.

## Why this is a solid starting point

1. Interface-first design enables backend replacement with minimal service changes.
2. Mutex-protected map provides predictable behavior for local development.
3. In-memory store keeps iteration fast while domain rules are still being shaped.

## Current code references

1. `internal/store/link_store.go`
2. `internal/store/memory_store.go`
3. `internal/shortener/service.go`

## Integrity gaps to close (Phase 2 and Phase 3)

1. No collision-safe uniqueness guarantee beyond current map overwrite behavior.
2. No durable storage, constraints, or migrations.
3. No transaction boundaries because there is no database yet.

## Planned direction

1. Phase 2:
- add bounded collision handling and conflict semantics
- define idempotency behavior
2. Phase 3:
- Postgres-backed store
- migrations
- unique index on `code`
- indexes for planned query paths

## Staff-level practice to keep

1. Put invariants in storage constraints, not only in application code.
2. Keep repository interfaces stable while swapping implementations.
3. Treat migrations as first-class, versioned code.
