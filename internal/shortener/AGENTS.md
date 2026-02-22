# AGENTS.md - `internal/shortener` Scope

This file supplements the repo root `AGENTS.md`.

Rules for this folder:

1. Keep domain and business logic here (create/resolve/validation/code generation).
2. Accept and propagate `context.Context` in service-level APIs.
3. Do not import Gin or HTTP transport types in this package.
4. Prefer table-driven unit tests for rules and invariants.
5. Use domain-oriented errors and wrap unexpected errors with context.
6. If domain rules change (validation, idempotency, expiry, collision, redirect semantics), update the corresponding `docs/engineering` file in the same change.
7. Keep `docs/engineering/link-correctness-and-lifecycle.md` synchronized with current service rules.
