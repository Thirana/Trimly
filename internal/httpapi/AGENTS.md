# AGENTS.md - `internal/httpapi` Scope

This file supplements the repo root `AGENTS.md`.

Rules for this folder:

1. Treat this package as the HTTP edge only.
2. Perform request binding and input validation here.
3. Return consistent JSON error shapes and status codes.
4. Keep redirect and handlers lean and avoid heavy per-request work.
5. Do not place SQL/Redis/storage details directly in this package.
6. Use `net/http/httptest` based tests for handler behavior and error mapping.
7. If request/response contracts, validation behavior, or error mapping changes, update the relevant files in `docs/engineering` in the same change.
8. If redirect endpoint behavior or hot-path performance characteristics change, also update:
- `docs/engineering/redirect-hot-path-and-caching.md`
- `docs/engineering/link-correctness-and-lifecycle.md`
