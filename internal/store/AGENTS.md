# AGENTS.md - `internal/store` Scope

This file supplements the repo root `AGENTS.md`.

Rules for this folder:

1. Keep storage behind interfaces so implementations remain swappable.
2. Ensure store methods accept `context.Context`.
3. Keep storage concerns separate from HTTP concerns.
4. Enforce deadlines/timeouts at call boundaries where applicable.
5. Preserve deterministic behavior for tests and in-memory implementations.
6. If store contracts, persistence strategy, migrations, indexes, or cache interactions change, update `docs/engineering/database-integrity-and-consistency.md` (and related docs) in the same change.
7. If cache-aside behavior is introduced or changed, update `docs/engineering/redirect-hot-path-and-caching.md`.
