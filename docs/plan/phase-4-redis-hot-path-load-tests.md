# Phase 4: Redis Caching Hot Path and Load Tests

## Goal

Optimize redirect throughput and latency with Redis cache-aside strategy, then validate with load tests.

## Scope

1. Redis integration for code-to-target resolution.
2. Cache-aside flow for redirect endpoint.
3. TTL strategy and cache invalidation policy.
4. Basic load-test suite and performance baseline.

## Current implementation milestone (2026-03-01)

1. Redis cache adapter added with versioned key format:

- `v1:url:short:{code}`
- `v1:url:miss:{code}`

1. Service resolve flow uses cache-aside:

- `GET short` -> `GET miss` -> DB fallback

1. Create flow clears miss key and warms short key.
2. Runtime toggle + startup fail-fast checks are wired through env config.
3. Resolve-path counters and periodic cache metrics logs are available for local validation.
4. Initial k6 redirect baseline script is available for before/after comparisons.
5. Redis-off vs Redis-on load-test comparison has been executed and documented.

## Implementation approach

1. Redirect read path:

- `GET v1:url:short:{code}` from Redis
- if hit: redirect immediately
- if miss: `GET v1:url:miss:{code}`
- if miss-key hit: return `404` immediately
- if both miss: query Postgres, then:
  - found -> write `v1:url:short:{code}` with TTL, redirect
  - not found/expired -> write `v1:url:miss:{code}` with short TTL, return `404`

1. Use compact serialized cache payload for minimal decode overhead.
2. Add strict timeout and bounded retry policy for Redis calls.
3. Define cache miss behavior:

- optional short-lived negative caching for unknown/expired codes

1. Keep redirect handler lean and avoid synchronous side tasks.
2. Keep Postgres as source of truth; Redis failures must fall back to DB path.
3. Gate Redis path with config (for example `REDIS_ENABLED`) for safe rollback.
4. Parse and validate cache config from env at startup:

- `REDIS_ENABLED`
- `REDIS_URL` (required when enabled)
- `REDIS_POSITIVE_TTL`
- `REDIS_MISS_TTL`
- `REDIS_CONNECT_TIMEOUT`
- `REDIS_OP_TIMEOUT` (fallback to `REDIS_TIMEOUT` for backward compatibility)

1. On `REDIS_ENABLED=true`, perform startup Redis connectivity check (`PING`) and fail fast on connection/auth/TLS errors.

## Runtime and correctness guardrails

1. Connection pooling:

- use one shared Redis client/pool
- never create Redis clients per request

1. Expiry semantics:

- expired links and missing links both map to `404` behavior
- both are eligible for short-lived miss-key caching

1. Correctness-first optimization:

- preserve existing HTTP behavior (`302` redirect, `404` for missing/expired)
- do not let cache logic change API contracts

1. Go implementation constraints:

- propagate `context.Context` through service/store/cache boundaries
- avoid heavy allocations/JSON churn in redirect hot path

## Redis command discipline (Phase 4)

1. Hot path command order:

- `GET v1:url:short:{code}`
- then `GET v1:url:miss:{code}` only when short-key misses

1. Do not use `EXISTS` in redirect hot path:

- `EXISTS + GET` adds an extra round trip with no value in this flow

1. `MGET` is optional and non-hot-path:

- use for cache warm-up, debug, or batch admin tooling only
- do not introduce `MGET` in single-code redirect handler path

## Scale and hotspot safeguards

1. Cache stampede mitigation:

- start with short negative-cache TTL and optional per-code `singleflight`
- add small TTL jitter to reduce synchronized expirations

1. Hot key behavior:

- expect skew where a few codes dominate traffic
- validate Redis/DB pool behavior under skewed load tests

## Data and API notes

1. Use versioned key prefixes:

- `v1:url:short:{code}` for positive cache entries
- `v1:url:miss:{code}` for short-lived negative cache entries

1. Keep DB as source of truth.
2. Ensure create/update/delete flows invalidate or refresh relevant keys:

- delete `v1:url:short:{code}`
- delete `v1:url:miss:{code}`

## Observability requirements

1. Measure cache effectiveness:

- hit ratio (positive key)
- negative-cache hit ratio (miss key)

1. Measure resilience:

- Redis timeout/error count
- fallback-to-DB count

1. Measure user impact:

- redirect latency (`p50`, `p95`, `p99`)
- throughput and error rate

1. Keep metrics labels bounded to avoid high-cardinality issues.

## Load testing plan

1. Choose a simple tool (`k6`, `vegeta`, or equivalent).
2. Define baseline scenarios:

- redirect-heavy read workload
- mixed create + redirect workload

1. Capture:

- throughput
- p50/p95/p99 latency
- error rate
- cache hit ratio
- fallback-to-DB rate

1. Compare before/after Redis integration.

## Rollout and rollback strategy

1. Roll out Redis caching behind config toggle.
2. If regressions occur, disable Redis path and serve via DB-only path.
3. Keep rollback data-safe:

- Redis keys are disposable cache state
- no schema/data migration dependency for Redis disablement

## Risks and mitigations

1. Cache inconsistency after updates:

- explicit invalidation on writes.

1. Redis outage impact:

- graceful fallback to DB with tighter safeguards.

1. Over-caching stale data:

- TTL and invalidation policy documented and tested.

1. Hot path command bloat:

- enforce `GET`-first command discipline and avoid extra lookups (`EXISTS`)

1. Stampede on cache miss bursts:

- short miss TTL, jitter, and optional `singleflight` for same-code concurrent misses

1. Hot-key saturation:

- load-test skewed traffic profile and verify pool/latency behavior

## Exit criteria

1. Redirect cache-aside flow implemented and tested. `done`
2. Cache hit ratio and latency improvements are measurable. `done`
3. System remains correct under cache miss and Redis failure cases. `done`
4. Rollback path (disable Redis) is tested and documented. `done`
5. Load-test results documented for future tuning. `done`

## Completion evidence (2026-03-01)

1. Workload:

- `k6 run --summary-trend-stats="avg,min,med,max,p(90),p(95),p(99)" scripts/load/redirect_baseline.js`
- `VUS=20`, `DURATION=20s`

1. Redis-off baseline:

- `http_reqs`: `26.61 req/s`
- `p50`: `638ms`
- `p95`: `1.15s`
- `p99`: `2.88s`

1. Redis-on:

- `http_reqs`: `65.19 req/s`
- `p50`: `285.5ms`
- `p95`: `320.65ms`
- `p99`: `1.37s`

1. Rollback validation:

- ran the same workload with `REDIS_ENABLED=false`; service remained correct and stable.

1. Source of record:

- `docs/engineering/performance-and-load-engineering.md`

