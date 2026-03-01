# Redirect Hot Path and Caching

## Purpose

Define how redirect performance is engineered while preserving correctness.

## Current implementation

1. Redirect route is `GET /:code`.
2. Resolve path now supports Redis cache-aside flow in service layer.
3. Cache command order is:
- `GET v1:url:short:{code}`
- on short miss, `GET v1:url:miss:{code}`
4. Cache behavior:
- short-key hit -> return cached link target
- miss-key hit -> return `404` without DB lookup
- both miss -> query DB and refresh cache
5. Service performs expiry checks for both cached and DB-fetched links.
6. On active link hit, handler issues `302` redirect.
7. On miss or expired link, handler returns `404` JSON error.

## Current code references

1. `internal/httpapi/router.go`
2. `internal/httpapi/links_handlers.go`
3. `internal/shortener/service.go`
4. `internal/shortener/cache.go`
5. `internal/store/rediscache/cache.go`
6. `internal/store/memory_store.go`

## Hot path correctness notes

1. Expired links intentionally map to the same `404 not_found` response as missing links.
2. Redirect path avoids create-path concerns (collision retries/idempotency logic is create-only).
3. Cache errors degrade gracefully to DB path; cache is an optimization layer only.
4. Create path clears miss key and warms short key to avoid stale negative-cache responses.

## Implemented cache-aside architecture

1. Cache-aside flow:
- `Redis GET v1:url:short:{code}` on redirect
- short-key hit -> redirect immediately
- short-key miss -> `Redis GET v1:url:miss:{code}`
- miss-key hit -> return `404` immediately
- both miss -> DB lookup -> cache set -> redirect/404
2. DB remains source of truth.
3. Redirect handler remains minimal (no blocking analytics or heavy logic).

## Key format contract (Phase 4)

1. Positive key:
- `v1:url:short:{code}` -> payload needed for redirect.
2. Negative key:
- `v1:url:miss:{code}` -> marker that code is currently unresolved.
3. `v1` prefix enables future key-schema migration without key collisions.

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

## Runtime toggles and timeouts

1. Cache is toggleable via `REDIS_ENABLED`.
2. Startup Redis connectivity check uses `REDIS_CONNECT_TIMEOUT`.
3. Runtime Redis operations use `REDIS_OP_TIMEOUT` (or legacy `REDIS_TIMEOUT` fallback).

## Metrics to track

1. Redirect QPS
2. p95/p99 latency
3. Cache short-key hit count
4. Cache miss-key hit count
5. DB fallback/hit/miss counts
6. Cache error count
7. Error rate by layer (handler/service/cache/store)

## Related plans

1. `docs/plan/phase-4-redis-hot-path-load-tests.md`
2. `docs/plan/phase-6-abuse-protection-observability-analytics.md`
