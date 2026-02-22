# Phase 3: Persistence and Migrations (Postgres)

## Goal

Move from in-memory storage to durable Postgres persistence with clean migration workflow.

## Scope

1. Postgres-backed link store implementation.
2. Schema migrations and local developer workflow.
3. Config and startup wiring for DB connectivity.
4. Repository-level tests covering store behavior.

## Implementation approach

1. Define schema for links:
- `id` (UUID or bigserial)
- `code` (unique)
- `original_url`
- `created_at`
- `updated_at`
- `expires_at` (nullable)
- ownership fields for future auth integration (`user_id`, nullable initially)
2. Introduce migration tooling:
- versioned `up/down` SQL files
- migration command integrated into dev workflow
3. Implement `internal/store` Postgres adapter behind existing interfaces.
4. Keep service layer unchanged where possible by preserving interface contracts.
5. Add connection handling:
- DSN from `DATABASE_URL`
- startup ping and fail-fast behavior
- graceful shutdown with context deadlines

## Data and API notes

1. Add unique indexes on `code`.
2. Add indexes needed for planned queries:
- `created_at`
- `user_id` (for future "my links" endpoints)
3. Keep API contract stable while switching storage backend.

## Testing plan

1. Store integration tests against disposable Postgres instance.
2. Migration test:
- apply all migrations from empty DB
- verify schema shape and constraints
3. Regression tests for existing handler/service behavior.

## Risks and mitigations

1. Migration drift across environments:
- enforce migration order and checksum validation where supported.
2. DB-specific behavior differences:
- test for unique violation and map to domain conflict error.
3. Local setup friction:
- provide `docker compose` or one-command setup for Postgres.

## Exit criteria

1. DB schema versioned and reproducible.
2. Postgres store is the primary implementation path.
3. Duplicate code conflicts handled correctly.
4. CI/local tests run reliably with persistence enabled.
