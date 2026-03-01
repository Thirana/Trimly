# ADR-0006: Initial Platform Provider Baseline

## Status

`superseded` by `ADR-0008`

## Date

2026-02-28

## Context

The project is learning-first but should follow production-style engineering practices.
A default provider baseline is needed before Phase 3 and Phase 4 implementation so infra choices do not block development.

## Decision

1. Postgres provider: `Neon`.
2. Redis provider: `Redis Cloud`.
3. Backend hosting provider: `Render`.
4. Frontend hosting provider (Next.js): `Vercel`.

## Rationale

1. Minimize operational overhead while keeping architecture realistic.
2. Keep service contracts aligned with mainstream managed offerings.
3. Preserve flexibility by keeping store/cache behind interfaces in code.
4. Avoid premature platform complexity for expected low traffic.

## Consequences

1. Phase 3 DB integration targets Neon-compatible Postgres connection and migration workflow.
2. Phase 4 cache integration targets Redis Cloud-compatible connection semantics.
3. Setup docs must include cloud environment variables and deployment workflow.
4. Provider-specific operational constraints are accepted for now.

## Alternatives considered

1. Supabase for Postgres (and broader platform services).
2. Self-managed database/cache.
3. Single provider for all components.

## Follow-up rules

1. If provider baseline changes, add a new ADR that supersedes this one.
2. Keep these docs synchronized with any provider change:
- `docs/setup/deployment-targets.md`
- `docs/setup/environment-matrix.md`
- `docs/setup/deployment-workflow.md`
- affected phase docs in `docs/plan`
