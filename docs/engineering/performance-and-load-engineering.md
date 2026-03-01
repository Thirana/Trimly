# Performance and Load Engineering

## Purpose

Track how performance claims are validated and which optimizations are accepted.

## Current status

1. Initial load-test harness committed:

- `scripts/load/redirect_baseline.js` (k6 redirect baseline)

1. Phase 4 cache-aside logic is implemented in redirect resolve path:

- short-key and miss-key lookups in Redis
- DB fallback on cache misses/errors

1. Redirect handler remains intentionally minimal.
2. Phase 2 introduced correctness logic mainly on create path:

- idempotency lookup by intent
- bounded collision retries
- expiry checks

1. Phase 4 baseline measurements are now recorded (Redis off vs Redis on, same workload).

## Phase 4 measurement run (2026-03-01)

1. Workload:

- `k6` script: `scripts/load/redirect_baseline.js`
- `VUS=20`, `DURATION=20s`, redirect-only workload (`GET /:code`)

1. Environment:

- local machine
- backend: Go + Gin
- DB: Neon Postgres
- cache: Upstash Redis (enabled only in Redis-on run)

1. Method:

- run A with `REDIS_ENABLED=false` (rollback path)
- run B with `REDIS_ENABLED=true` and same k6 parameters
- collect k6 summary stats with `--summary-trend-stats="avg,min,med,max,p(90),p(95),p(99)"`

1. Results:

- Redis off:
  - `http_reqs`: `26.61 req/s` (549 requests)
  - `p50` (`med`): `638ms`
  - `p95`: `1.15s`
  - `p99`: `2.88s`
  - `http_req_failed`: `0.00%`
  - threshold `p(95)<500ms`: failed
- Redis on:
  - `http_reqs`: `65.19 req/s` (1322 requests)
  - `p50` (`med`): `285.5ms`
  - `p95`: `320.65ms`
  - `p99`: `1.37s`
  - `http_req_failed`: `0.00%`
  - threshold `p(95)<500ms`: passed

1. Observed cache counters during Redis-on run:

- `cache_metrics short_hit=1130 miss_hit=0 db_fallback=0 db_hit=0 db_miss=0 cache_error=0`
- This is expected for warmed links (create path warms positive key).

1. Decision:

- keep Redis cache-aside path enabled by default where `REDIS_ENABLED=true`.
- keep rollback switch (`REDIS_ENABLED=false`) as the fast mitigation path.

## Principles

1. No "faster" claims without measurement.
2. Optimize highest-traffic path first (`GET /:code`).
3. Prefer simple architecture until measurements justify complexity.
4. Record baseline before and after major changes.

## Measurement roadmap

1. Phase 4:

- introduce load tests for redirect-heavy and mixed workloads
- record latency and throughput baseline with/without Redis

1. Phase 6:

- add observability signals needed for performance diagnosis

1. Phase 7:

- profile CPU/memory/allocation hotspots
- tune runtime, connection pools, and deployment settings

## Baseline KPI set

1. Throughput (requests/second)
2. Latency (p50, p95, p99)
3. Error rate
4. Cache hit ratio (once Redis exists)
5. DB query latency (once Postgres exists)

## Reporting format for major changes

1. Workload description
2. Test environment
3. Baseline metrics
4. Post-change metrics
5. Decision: keep/revert/follow-up

## Related plans

1. `docs/plan/phase-4-redis-hot-path-load-tests.md`
2. `docs/plan/phase-7-performance-deployment-hardening.md`

