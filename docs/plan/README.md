# URL Shortener Implementation Plan

This folder contains phase-by-phase implementation notes for the URL shortener roadmap.
The goal is shared context for humans and agents, not exhaustive specs.

## Phases

1. `docs/plan/phase-2-correctness-reliability.md`
2. `docs/plan/phase-3-persistence-migrations.md`
3. `docs/plan/phase-4-redis-hot-path-load-tests.md`
4. `docs/plan/phase-5-auth-profiles-link-management.md`
5. `docs/plan/phase-6-abuse-protection-observability-analytics.md`
6. `docs/plan/phase-7-performance-deployment-hardening.md`

## How to use these docs

1. Read phases in order.
2. Treat each file as implementation intent and acceptance baseline.
3. Keep changes incremental and verify with tests at each milestone.
4. Update phase docs when scope or sequencing changes.
5. Before implementing Phase 4, review `docs/go-and-gin/phase-4-concepts.md` to align on Redis command/key and hot-path correctness rules.
6. Keep Phase 4 concept alignment across later phases:
- Phase 4 owns implementation details (keys, command discipline, rollback/fallback)
- Phase 6 owns production observability signals
- Phase 7 owns release-hardening validation under failure/skew
