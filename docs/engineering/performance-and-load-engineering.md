# Performance and Load Engineering

## Purpose

Track how performance claims are validated and which optimizations are accepted.

## Current status

1. No formal benchmark suite yet.
2. No load-test harness committed yet.
3. Basic redirect path is implemented and intentionally simple.

## Principles

1. No "faster" claims without measurement.
2. Optimize highest-traffic path first (`GET /:code`).
3. Prefer simple architecture until measurements justify complexity.
4. Record baseline before and after major changes.

## Measurement roadmap

1. Phase 4:
- introduce load tests for redirect-heavy and mixed workloads
- record latency and throughput baseline with/without Redis
2. Phase 6:
- add observability signals needed for performance diagnosis
3. Phase 7:
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
