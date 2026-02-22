# ADR-0002: Service Depends on Store Interface, In-Memory First

## Status

`accepted`

## Date

2026-02-22

## Context

The project needs rapid iteration for domain rules while preserving a path to production persistence.
Directly coupling service logic to a concrete database early slows iteration and increases refactor cost.

## Decision

Define a `LinkStore` interface and make `shortener.Service` depend on it.
Use an in-memory implementation first, then introduce Postgres and Redis behind the same boundaries in later phases.

## Consequences

1. Faster early delivery and easier unit testing.
2. Cleaner migration path to Postgres/Redis.
3. Need to carefully preserve interface compatibility as persistence features grow.

## Alternatives considered

1. Build directly against Postgres from day one.
2. Use globals/singletons for storage state.

## Implementation notes

1. Interface: `internal/store/link_store.go`
2. Memory implementation: `internal/store/memory_store.go`
3. Service usage: `internal/shortener/service.go`
4. Planned expansion: `docs/plan/phase-3-persistence-migrations.md` and `docs/plan/phase-4-redis-hot-path-load-tests.md`
