# Phase 4 Concepts: Redis Caching and Load Testing

This guide explains the core concepts you need before implementing Phase 4.
Each concept includes:

1. What it is
2. Why we use it
3. What problem it solves

## 1) Redirect hot path

1. What it is:

- The highest-traffic request path in this app: `GET /:code`.

1. Why we use it:

- Most shortener traffic is redirects, not link creation.

1. Problem it solves:

- Keeps optimization work focused where latency and throughput matter most.

## 2) Cache-aside pattern

1. What it is:

- Read from cache first; if missing, read DB, then write result to cache.

1. Why we use it:

- Simple, widely used, and keeps DB as source of truth.

1. Problem it solves:

- Reduces repeated DB reads for popular codes.

## 3) Source of truth (DB-first correctness)

1. What it is:

- Postgres remains authoritative for final state; Redis is only a copy for speed.

1. Why we use it:

- Cache can be stale or unavailable; DB correctness must remain intact.

1. Problem it solves:

- Prevents logic from depending on volatile cache state.

## 4) Cache hit / miss

1. What it is:

- Hit: code found in Redis.
- Miss: code not found in Redis.

1. Why we use it:

- Different handling paths let us minimize work on hits.

1. Problem it solves:

- Speeds common reads while retaining fallback for correctness.

## 5) Cache key design

1. What it is:

- Stable and versioned key naming convention:
- `v1:url:short:{code}`
- `v1:url:miss:{code}`

1. Why we use it:

- Predictable naming avoids collisions and supports operations/debugging.
- `v1` allows key schema evolution later (for example `v2`) without breaking old keys.

1. Problem it solves:

- Prevents accidental key overlap and messy cache management.

## 6) Serialized cache payload

1. What it is:

- Compact value stored in Redis (for example, URL + expiry info).

1. Why we use it:

- Small payloads are faster to read/write and decode.

1. Problem it solves:

- Reduces per-request overhead on hot path.

## 7) TTL (time-to-live)

1. What it is:

- Automatic expiration time for cache entries.

1. Why we use it:

- Limits stale data duration and controls memory usage.

1. Problem it solves:

- Avoids indefinitely stale cache entries.

## 8) Invalidation policy

1. What it is:

- Rules to remove/update cache keys when write-side data changes.

1. Why we use it:

- Writes can make old cache entries incorrect.

1. Problem it solves:

- Prevents stale redirects after link changes.

## 9) Negative caching

1. What it is:

- Caching unresolved codes using a dedicated miss key: `v1:url:miss:{code}` for a short TTL.

1. Why we use it:

- Repeated misses can overload DB.

1. Problem it solves:

- Reduces DB pressure from repeated invalid code lookups.

## 10) Expiry semantics in cache path

1. What it is:

- Cached entries must respect `expires_at` and existing `404` behavior for expired links.
- Expired or missing should both be eligible for miss-key caching with short TTL.

1. Why we use it:

- Current contract treats expired and missing as the same public response.

1. Problem it solves:

- Prevents correctness/security regression while adding cache.

## 11) Redis timeout strategy

1. What it is:

- Strict per-request time limits for Redis operations.

1. Why we use it:

- Redis delays should not block redirect path for too long.

1. Problem it solves:

- Avoids tail-latency spikes and request pileups.

## 12) Retry policy

1. What it is:

- Controlled retry rules for transient Redis errors.

1. Why we use it:

- Some failures are brief; retries can recover quickly.

1. Problem it solves:

- Improves resilience without causing retry storms.

## 13) Graceful fallback behavior

1. What it is:

- If Redis fails, continue with DB path.

1. Why we use it:

- Availability should not depend fully on cache health.

1. Problem it solves:

- Keeps service functional during cache incidents.

## 14) Connection pooling

1. What it is:

- Reusing established Redis connections.

1. Why we use it:

- Creating connections per request is expensive.

1. Problem it solves:

- Reduces CPU/network overhead and latency.

## 15) Cache stampede risk

1. What it is:

- Many requests miss same key and all hit DB at once.

1. Why we use it:

- Popular keys can create synchronized DB bursts.

1. Problem it solves:

- Helps design protections (for example singleflight or jittered TTL).

## 16) Hot key behavior

1. What it is:

- A few very popular codes dominate traffic.

1. Why we use it:

- Shorteners naturally produce skewed access patterns.

1. Problem it solves:

- Drives monitoring and capacity planning decisions.

## 17) Observability for cache layer

1. What it is:

- Metrics/logs for hit ratio, Redis errors, fallback rate, and latency.

1. Why we use it:

- You cannot tune what you cannot measure.

1. Problem it solves:

- Prevents blind optimization and unclear incidents.

## 18) Performance metrics (p50/p95/p99)

1. What it is:

- Percentile latency and throughput metrics.

1. Why we use it:

- Average latency hides user-facing outliers.

1. Problem it solves:

- Makes performance evaluation realistic and actionable.

## 19) Load testing baseline vs after-change

1. What it is:

- Measure current system first, then compare after Redis integration.

1. Why we use it:

- Performance claims need evidence.

1. Problem it solves:

- Avoids shipping complexity without measurable benefit.

## 20) Workload scenarios

1. What it is:

- Redirect-heavy and mixed create+redirect test profiles.

1. Why we use it:

- Real traffic is not one-dimensional.

1. Problem it solves:

- Prevents overfitting to a single synthetic test.

## 21) Correctness-first optimization

1. What it is:

- Keep current API/error contracts while improving speed.

1. Why we use it:

- Phase 2 already established correctness rules.

1. Problem it solves:

- Prevents performance work from breaking behavior guarantees.

## 22) Rollback strategy

1. What it is:

- Ability to disable cache path safely if issues appear.

1. Why we use it:

- New infrastructure can introduce unknown failure modes.

1. Problem it solves:

- Reduces production risk during rollout.

## 23) Go-specific implementation concerns

1. What it is:

- Context propagation, timeout usage, and minimizing allocations in redirect path.

1. Why we use it:

- Go services can degrade under tail latency and allocation pressure.

1. Problem it solves:

- Keeps hot path predictable and resource-efficient.

## 24) `MGET` for batch workflows (optional, practical)

1. What it is:

- Redis command to fetch multiple keys in one round trip.

1. Why we use it:

- Helpful for cache warm-up scripts, batch debugging, and internal tools.

1. Problem it solves:

- Reduces network round trips when checking many codes at once.
- Keeps operational/debug workflows efficient.

## 25) Prefer `GET` over `EXISTS` in hot path logic

1. What it is:

- In runtime read paths, call `GET` directly and treat nil as cache miss.

1. Why we use it:

- `EXISTS` followed by `GET` usually creates an extra network call.

1. Problem it solves:

- Avoids unnecessary latency and load in redirect hot path.
- Keeps command usage minimal and predictable.

Use `EXISTS` mainly for debugging or special branching where value retrieval is not required.

## Suggested learning order

1. Cache-aside, source-of-truth, key design, TTL, invalidation.
2. Failure handling: timeout, fallback, retry, negative caching.
3. Performance: p95/p99, hit ratio, workload design, baseline comparisons.
4. Advanced risks: stampede, hot keys, stale data windows.

## Phase 4 readiness checklist

You are ready to implement Phase 4 when you can explain:

1. How a single redirect request behaves on hit vs miss.
2. Why DB remains source of truth even with Redis.
3. Which keys to invalidate on write operations.
4. Which metrics prove Redis is actually helping.
5. How system behaves when Redis is slow or down.
