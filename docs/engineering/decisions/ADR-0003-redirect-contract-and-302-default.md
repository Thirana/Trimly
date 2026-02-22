# ADR-0003: Redirect Contract and `302` as Initial Default

## Status

`accepted`

## Date

2026-02-22

## Context

Redirect behavior is the hottest and most externally visible path in a URL shortener.
Permanent redirect codes (`301`/`308`) can create sticky caching behavior in browsers/proxies and are hard to reverse safely during early product evolution.

## Decision

Use `302 Found` as the default redirect status in initial phases.
Revisit `301`/`308` only with explicit product and caching strategy decisions.

## Consequences

1. Safer iterative behavior while product semantics are still evolving.
2. Lower risk of accidental long-lived client/proxy caching mistakes.
3. Slightly weaker cacheability compared to permanent redirects.

## Alternatives considered

1. `301` by default.
2. `308` by default.
3. Per-link configurable redirect status in early phase.

## Implementation notes

1. Current redirect implementation: `internal/httpapi/links_handlers.go`
2. Related docs:
- `docs/engineering/link-correctness-and-lifecycle.md`
- `docs/engineering/redirect-hot-path-and-caching.md`
