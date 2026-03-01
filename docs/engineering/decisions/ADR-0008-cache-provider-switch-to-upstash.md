# ADR-0008: Cache Provider Switch to Upstash

## Status

`accepted`

## Date

2026-03-01

## Context

Phase 4 introduces Redis caching for the redirect hot path.
The project is learning-first and should preserve enough free-tier capacity for cache-key experiments and load-test drills without immediate paid limits.

The initial provider baseline (`ADR-0006`) selected Redis Cloud.
The current decision revisits only the Redis provider choice while preserving other platform targets.

## Decision

1. Keep Postgres provider as `Neon`.
2. Keep backend hosting provider as `Render`.
3. Keep frontend hosting provider as `Vercel`.
4. Switch Redis provider baseline from `Redis Cloud` to `Upstash Redis` for Phase 4+.

## Consequences

1. Setup docs now use Upstash as the default source of `REDIS_URL`.
2. Deployment workflow now provisions Upstash Redis in the base path.
3. Runtime cache behavior in code remains provider-agnostic (`REDIS_URL` + Redis protocol).
4. Operational caveats for this baseline:
- account for Upstash quota/command limits during load testing
- prefer TLS connection URL (`rediss://...`)
- keep DB fallback logic as primary correctness guardrail

## Alternatives considered

1. Keep Redis Cloud as initially selected in `ADR-0006`.
2. Self-managed Redis for local and cloud environments.
3. Defer Redis provider selection until after initial Phase 4 coding.

## Implementation notes

1. `docs/setup/deployment-targets.md`
2. `docs/setup/environment-matrix.md`
3. `docs/setup/deployment-workflow.md`
4. `docs/plan/phase-3-persistence-migrations.md`
5. `docs/engineering/decisions/ADR-0006-initial-platform-provider-baseline.md`

## Supersedes / superseded by

Supersedes cache-provider portion of `ADR-0006`.
