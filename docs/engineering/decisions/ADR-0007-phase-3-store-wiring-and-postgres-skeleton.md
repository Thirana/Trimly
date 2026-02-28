# ADR-0007: Phase 3 Store Wiring and Postgres Skeleton

## Status

`accepted`

## Date

2026-02-28

## Context

Phase 3 needs durable persistence without breaking Phase 2 behavior.
The project must keep clean architecture boundaries while introducing DB connectivity and migration scaffolding.

## Decision

1. Runtime store backend is selected in `cmd/api` using `DATABASE_URL`.
2. `internal/httpapi` router no longer constructs stores/services; it only maps routes and handlers.
3. Introduce Postgres store skeleton at `internal/store/postgres` implementing `LinkStore`.
4. Use `pgx/v5` (`pgxpool`) for Postgres connectivity.
5. Add migration scaffold under `internal/store/postgres/migrations` with links table and required unique/index constraints.
6. Enable startup DB ping fail-fast and graceful HTTP shutdown with timeout.

## Consequences

1. Composition root is now explicit in `cmd/api/main.go`.
2. Local development still works without DB via in-memory fallback.
3. DB-backed path is available when `DATABASE_URL` is configured.
4. Additional integration tests/migration verification are still needed to complete Phase 3.

## Alternatives considered

1. Keep dependency wiring in router package.
2. Make Postgres mandatory immediately with no fallback.
3. Use `database/sql` with alternate driver package.

## Implementation notes

1. `cmd/api/main.go`
2. `internal/httpapi/router.go`
3. `internal/store/postgres/store.go`
4. `internal/store/postgres/migrations/000001_create_links_table.up.sql`
5. `internal/store/postgres/migrations/000001_create_links_table.down.sql`
6. `internal/store/intent.go`
