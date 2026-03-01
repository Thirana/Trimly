# Phase 4: Redis Caching Hot Path and Load Tests

## Goal

Optimize redirect throughput and latency with Redis cache-aside strategy, then validate with load tests.

## Scope

1. Redis integration for code-to-target resolution.
2. Cache-aside flow for redirect endpoint.
3. TTL strategy and cache invalidation policy.
4. Basic load-test suite and performance baseline.

## Implementation approach

1. Redirect read path:
- `GET v1:url:short:{code}` from Redis
- if hit: redirect immediately
- if miss: `GET v1:url:miss:{code}`
- if miss-key hit: return `404` immediately
- if both miss: query Postgres, then:
  - found -> write `v1:url:short:{code}` with TTL, redirect
  - not found/expired -> write `v1:url:miss:{code}` with short TTL, return `404`
2. Use compact serialized cache payload for minimal decode overhead.
3. Add strict timeout and bounded retry policy for Redis calls.
4. Define cache miss behavior:
- optional short-lived negative caching for unknown/expired codes
5. Keep redirect handler lean and avoid synchronous side tasks.
6. Keep Postgres as source of truth; Redis failures must fall back to DB path.
7. Gate Redis path with config (for example `REDIS_ENABLED`) for safe rollback.
8. Parse and validate cache config from env at startup:
- `REDIS_ENABLED`
- `REDIS_URL` (required when enabled)
- `REDIS_POSITIVE_TTL`
- `REDIS_MISS_TTL`
- `REDIS_CONNECT_TIMEOUT`
- `REDIS_OP_TIMEOUT` (fallback to `REDIS_TIMEOUT` for backward compatibility)
9. On `REDIS_ENABLED=true`, perform startup Redis connectivity check (`PING`) and fail fast on connection/auth/TLS errors.

## Runtime and correctness guardrails

1. Connection pooling:
- use one shared Redis client/pool
- never create Redis clients per request
2. Expiry semantics:
- expired links and missing links both map to `404` behavior
- both are eligible for short-lived miss-key caching
3. Correctness-first optimization:
- preserve existing HTTP behavior (`302` redirect, `404` for missing/expired)
- do not let cache logic change API contracts
4. Go implementation constraints:
- propagate `context.Context` through service/store/cache boundaries
- avoid heavy allocations/JSON churn in redirect hot path

## Redis command discipline (Phase 4)

1. Hot path command order:
- `GET v1:url:short:{code}`
- then `GET v1:url:miss:{code}` only when short-key misses
2. Do not use `EXISTS` in redirect hot path:
- `EXISTS + GET` adds an extra round trip with no value in this flow
3. `MGET` is optional and non-hot-path:
- use for cache warm-up, debug, or batch admin tooling only
- do not introduce `MGET` in single-code redirect handler path

## Scale and hotspot safeguards

1. Cache stampede mitigation:
- start with short negative-cache TTL and optional per-code `singleflight`
- add small TTL jitter to reduce synchronized expirations
2. Hot key behavior:
- expect skew where a few codes dominate traffic
- validate Redis/DB pool behavior under skewed load tests

## Data and API notes

1. Use versioned key prefixes:
- `v1:url:short:{code}` for positive cache entries
- `v1:url:miss:{code}` for short-lived negative cache entries
2. Keep DB as source of truth.
3. Ensure create/update/delete flows invalidate or refresh relevant keys:
- delete `v1:url:short:{code}`
- delete `v1:url:miss:{code}`

## Observability requirements

1. Measure cache effectiveness:
- hit ratio (positive key)
- negative-cache hit ratio (miss key)
2. Measure resilience:
- Redis timeout/error count
- fallback-to-DB count
3. Measure user impact:
- redirect latency (`p50`, `p95`, `p99`)
- throughput and error rate
4. Keep metrics labels bounded to avoid high-cardinality issues.

## Load testing plan

1. Choose a simple tool (`k6`, `vegeta`, or equivalent).
2. Define baseline scenarios:
- redirect-heavy read workload
- mixed create + redirect workload
3. Capture:
- throughput
- p50/p95/p99 latency
- error rate
- cache hit ratio
- fallback-to-DB rate
4. Compare before/after Redis integration.

## Rollout and rollback strategy

1. Roll out Redis caching behind config toggle.
2. If regressions occur, disable Redis path and serve via DB-only path.
3. Keep rollback data-safe:
- Redis keys are disposable cache state
- no schema/data migration dependency for Redis disablement

## Risks and mitigations

1. Cache inconsistency after updates:
- explicit invalidation on writes.
2. Redis outage impact:
- graceful fallback to DB with tighter safeguards.
3. Over-caching stale data:
- TTL and invalidation policy documented and tested.
4. Hot path command bloat:
- enforce `GET`-first command discipline and avoid extra lookups (`EXISTS`)
5. Stampede on cache miss bursts:
- short miss TTL, jitter, and optional `singleflight` for same-code concurrent misses
6. Hot-key saturation:
- load-test skewed traffic profile and verify pool/latency behavior

## Exit criteria

1. Redirect cache-aside flow implemented and tested.
2. Cache hit ratio and latency improvements are measurable.
3. System remains correct under cache miss and Redis failure cases.
4. Rollback path (disable Redis) is tested and documented.
5. Load-test results documented for future tuning.
