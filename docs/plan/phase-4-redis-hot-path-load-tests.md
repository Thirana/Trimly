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
- `GET code` from Redis
- on hit: redirect immediately
- on miss: query Postgres, write cache with TTL, then redirect
2. Use compact serialized cache payload for minimal decode overhead.
3. Add timeout and retry policy appropriate for Redis calls.
4. Define cache miss behavior:
- optional short-lived negative caching for unknown/expired codes
5. Keep redirect handler lean and avoid synchronous side tasks.

## Data and API notes

1. Use stable cache key format such as `shortener:code:<code>`.
2. Keep DB as source of truth.
3. Ensure create/update/delete flows invalidate or refresh relevant cache keys.

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
4. Compare before/after Redis integration.

## Risks and mitigations

1. Cache inconsistency after updates:
- explicit invalidation on writes.
2. Redis outage impact:
- graceful fallback to DB with tighter safeguards.
3. Over-caching stale data:
- TTL and invalidation policy documented and tested.

## Exit criteria

1. Redirect cache-aside flow implemented and tested.
2. Cache hit ratio and latency improvements are measurable.
3. System remains correct under cache miss and Redis failure cases.
4. Load-test results documented for future tuning.
