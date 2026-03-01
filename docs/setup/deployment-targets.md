# Deployment Targets

This file defines the default managed services for this learning-first URL shortener.
The goal is low operational overhead, clean developer experience, and production-like patterns.

## Selected services

1. Postgres: `Neon`
2. Redis: `Upstash Redis`
3. Go backend hosting: `Render`
4. Next.js frontend hosting: `Vercel`

## Why this stack

1. Neon

- Managed Postgres with simple connection workflow.
- Good fit for Phase 3 (durable persistence + migrations).

1. Upstash Redis

- Larger free-tier memory footprint for learning-phase cache experiments.
- Good fit for Phase 4 redirect hot-path cache with standard Redis protocol support.
- Acceptable tradeoff for this project: eventual-consistency model and command-based quotas.

1. Render

- Straightforward Go service deployment and environment variable management.
- Good for API-first projects that do not need custom platform engineering.

1. Vercel

- Strong default platform experience for Next.js.
- Simple frontend preview and production deployment model.

## Deployment topology

1. `Local` (developer machine)

- Go API runs locally.
- Postgres and Redis can be local containers or managed dev instances.

1. `Cloud` (single production-like environment)

- Backend on Render.
- Postgres on Neon.
- Redis on Upstash.
- Frontend on Vercel.

1. `Optional staging` (later)

- Add once Phase 4/5 introduces higher risk changes.

## Region strategy

1. Keep backend, Postgres, and Redis in the same region when possible.
2. Start with one primary region for simplicity.
3. Add multi-region only after real latency/availability requirements justify it.

## Practical constraints

1. This project prioritizes learning and correctness over aggressive scaling.
2. Managed free/entry tiers are acceptable for initial phases.
3. Provider limits can change; always verify current pricing/limits before launch.

## Related docs

1. `docs/setup/environment-matrix.md`
2. `docs/setup/deployment-workflow.md`
3. `docs/plan/phase-3-persistence-migrations.md`
4. `docs/plan/phase-4-redis-hot-path-load-tests.md`
