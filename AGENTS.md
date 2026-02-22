# AGENTS.md - Enforced Rules (URL Shortener)

This project is a learning-first, production-grade URL shortener.
If user instructions conflict with this file, follow the user.

## Scope and boundaries

1. `cmd/api` is wiring only (config, dependencies, router, server lifecycle).
2. `internal/httpapi` is HTTP edge only (Gin handlers, DTO binding, error mapping, middleware).
3. `internal/shortener` is business/domain logic only (no Gin imports).
4. `internal/store` is storage interfaces and implementations (swappable backends).
5. `docs/engineering` is the canonical engineering-notes set for this repo and must track real implementation decisions.

Local rules also exist in:

1. `cmd/api/AGENTS.md`
2. `internal/httpapi/AGENTS.md`
3. `internal/shortener/AGENTS.md`
4. `internal/store/AGENTS.md`

Engineering notes are in:

1. `docs/engineering/README.md`
2. `docs/engineering/phase-progress.md`
3. `docs/engineering/decisions/README.md`

## Non-negotiable engineering rules

1. Keep `GET /:code` minimal and fast.
2. Use `context.Context` across service/store APIs and outbound calls.
3. Prefer correctness over cleverness, especially with concurrency.
4. Avoid adding dependencies unless clearly justified.
5. Keep diffs focused and avoid unrelated refactors.

## API and behavior rules

1. Validate input at HTTP boundary and re-check critical invariants in service layer.
2. Use safe URL parsing (`net/url`), not regex/prefix shortcuts.
3. Error responses must be consistent:

```json
{
  "error": {
    "code": "string",
    "message": "string"
  }
}
```

4. Status code baseline:
- `400` bad input
- `401` unauthenticated
- `403` unauthorized
- `404` not found
- `409` conflict
- `429` rate limited
- `5xx` unexpected only, no internal leak
5. Redirects default to `302`; do not introduce unsafe open-redirect behavior.

## Security rules

1. Never store plaintext passwords.
2. Never log secrets, tokens, cookies, session IDs, or password hashes.
3. Enforce ownership checks on user-owned resources.
4. Enforce request limits (body size, pagination bounds, timeouts).
5. Use secure cookie settings if cookies are used (`HttpOnly`, `Secure` in prod, explicit `SameSite` and CSRF strategy).

## Performance and observability rules

1. Redirect path should avoid heavy synchronous work and unnecessary allocations.
2. Set deadlines/timeouts for DB/cache calls.
3. Do not claim performance gains without measurement.
4. Never expose `pprof` publicly; keep it protected/disabled by default in production.

## Required workflow

If a task changes behavior or touches more than 3 files, write a short plan first.

Engineering docs maintenance is required:

1. If a change modifies architecture, validation, error flow, persistence, security, auth, authorization, or observability behavior, update the relevant file(s) in `docs/engineering` in the same change.
2. If phase scope or sequence changes, update `docs/plan` in the same change.
3. Do not leave engineering docs stale after implementation changes.
4. When work affects collision handling, idempotency, expiry, redirect hot path, caching, or performance claims, update:
- `docs/engineering/link-correctness-and-lifecycle.md`
- `docs/engineering/redirect-hot-path-and-caching.md`
- `docs/engineering/performance-and-load-engineering.md`
5. Keep `docs/engineering/phase-progress.md` aligned with real phase status.
6. When a key staff-level technical decision is made or changed, add/update an ADR in `docs/engineering/decisions` in the same change.
7. Never silently change a previously accepted major decision; add a new ADR that supersedes the old one.

Before finishing any code change:

1. Run `gofmt -w .`
2. Run `go test ./...`
3. Run `go test -race ./...` when concurrency was touched

Done criteria:

1. Behavior is correct and tested.
2. Code is formatted and idiomatic.
3. Error handling is consistent and safe.
4. Redirect path remains lean.
5. Security/performance impact is considered.
