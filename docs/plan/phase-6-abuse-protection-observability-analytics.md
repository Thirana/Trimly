# Phase 6: Abuse Protection, Observability, and Async Analytics

## Goal

Improve safety and operability with rate limits, observability foundations, and non-blocking analytics.

## Scope

1. Abuse protection and request limiting.
2. Structured logging and metrics.
3. Async click analytics pipeline that never blocks redirect path.
4. Baseline operational dashboards and alerts.

## Implementation approach

1. Rate limiting:
- apply strict limits on auth and link-create endpoints
- consider per-IP and per-user policies
2. Input hardening:
- request body size limits
- pagination caps
- timeout defaults for request handling
3. Logging:
- structured logs with `request_id`, `method`, `path`, `status`, `latency_ms`
- include `user_id` when authenticated
4. Metrics:
- redirect throughput
- p95/p99 latency
- cache hit ratio
- DB latency
- error rate
5. Async analytics:
- bounded channel for events
- worker pool for batched writes
- drop/sample strategy when queue is full

## Operational notes

1. Add request ID middleware if not present.
2. Protect debug/profiling endpoints and disable by default in production.
3. Document SLO-like targets for latency and error rate.

## Testing plan

1. Handler tests for rate-limit responses and headers.
2. Unit tests for analytics queue behavior when full.
3. Concurrency tests for worker shutdown and context cancellation.
4. Race detector run for analytics pipeline changes.

## Risks and mitigations

1. Redirect slowdown from observability:
- keep instrumentation light on hot path.
2. Unbounded memory from analytics buffering:
- enforce bounded queue and backpressure strategy.
3. Noisy logs/metrics:
- standardize fields and cardinality limits.

## Exit criteria

1. Abuse controls are active on critical endpoints.
2. Logging and metrics are consistent and useful.
3. Async analytics is non-blocking and race-safe.
4. Production-safe defaults for debug/profiling endpoints are enforced.
