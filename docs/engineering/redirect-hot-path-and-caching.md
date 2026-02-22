# Redirect Hot Path and Caching

## Purpose

Define how redirect performance is engineered while preserving correctness.

## Current implementation

1. Redirect route is `GET /:code`.
2. Resolve path uses in-memory store lookup.
3. On hit, handler issues `302` redirect.
4. On miss, handler returns `404` JSON error.

## Current code references

1. `internal/httpapi/router.go`
2. `internal/httpapi/links_handlers.go`
3. `internal/shortener/service.go`
4. `internal/store/memory_store.go`

## Phase 4 target architecture

1. Cache-aside flow:
- `Redis GET(code)` on redirect
- hit -> redirect immediately
- miss -> DB lookup -> cache set -> redirect
2. DB remains source of truth.
3. Redirect handler remains minimal (no blocking analytics or heavy logic).

## Design rules for hot path

1. Keep allocation and per-request work low.
2. Set strict timeouts on Redis/DB calls.
3. Define cache key format and TTL policy explicitly.
4. Define write-path invalidation strategy before rollout.
5. Decide negative cache semantics carefully and keep short TTL if enabled.

## Failure behavior

1. Redis unavailable -> controlled fallback to DB path.
2. DB unavailable -> return consistent internal error behavior.
3. Never block redirects on non-essential work.

## Metrics to track

1. Redirect QPS
2. p95/p99 latency
3. Cache hit ratio
4. Error rate by layer (handler/service/cache/store)

## Related plans

1. `docs/plan/phase-4-redis-hot-path-load-tests.md`
2. `docs/plan/phase-6-abuse-protection-observability-analytics.md`
