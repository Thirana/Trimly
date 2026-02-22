# Engineering Notes

This folder captures engineering decisions, practices, and current implementation status for this URL shortener.
It is intended for learning, onboarding, and keeping agent context accurate.

## How to read

1. Start with `phase-progress.md` for current status.
2. Read `link-correctness-and-lifecycle.md` and `redirect-hot-path-and-caching.md` for core domain and architecture context.
3. Then read `validation-and-types.md`, `error-handling.md`, and `data-modeling-and-api-shaping.md`.
4. Use `database-integrity-and-consistency.md`, `security-hardening.md`, and `observability-and-operations.md` for production-readiness direction.
5. Read `authentication-and-sessions.md` and `authorization.md` for Phase 5 implementation prep.
6. Use `performance-and-load-engineering.md` for Phase 4 and Phase 7 performance work.
7. Read ADRs in `docs/engineering/decisions/README.md` for key staff-level decisions and rationale.

## Index

1. `docs/engineering/validation-and-types.md`
2. `docs/engineering/error-handling.md`
3. `docs/engineering/data-modeling-and-api-shaping.md`
4. `docs/engineering/database-integrity-and-consistency.md`
5. `docs/engineering/security-hardening.md`
6. `docs/engineering/observability-and-operations.md`
7. `docs/engineering/authentication-and-sessions.md`
8. `docs/engineering/authorization.md`
9. `docs/engineering/link-correctness-and-lifecycle.md`
10. `docs/engineering/redirect-hot-path-and-caching.md`
11. `docs/engineering/performance-and-load-engineering.md`
12. `docs/engineering/phase-progress.md`
13. `docs/engineering/decisions/README.md`

## Maintenance rules

1. If code changes behavior or architecture, update the relevant file in this folder in the same change.
2. Keep each document pragmatic: what is implemented now, why, and what is next.
3. Link forward to `docs/plan` phases when a topic is planned but not implemented yet.
