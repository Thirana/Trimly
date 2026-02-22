# AGENTS.md - `cmd/api` Scope

This file supplements the repo root `AGENTS.md`.

Rules for this folder:

1. Keep `cmd/api/main.go` wiring-only.
2. Allow only composition and lifecycle code here:
- config loading
- dependency construction
- router creation
- HTTP server start/shutdown
3. Do not add business logic, validation logic, or storage/query logic here.
4. Keep startup/shutdown paths explicit and testable.
5. If startup behavior, environment handling, or lifecycle strategy changes, update `docs/engineering/observability-and-operations.md` and any other impacted docs in `docs/engineering`.
6. If bootstrap changes represent phase milestones, update `docs/engineering/phase-progress.md`.
