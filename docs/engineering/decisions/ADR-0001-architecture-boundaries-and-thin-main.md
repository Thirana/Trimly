# ADR-0001: Architecture Boundaries and Thin `cmd/api/main.go`

## Status

`accepted`

## Date

2026-02-22

## Context

The project needs clear package boundaries to support long-term maintainability and staff-level decision quality.
Without strict boundaries, HTTP transport logic, business logic, and storage concerns tend to mix and become harder to test and evolve.

## Decision

Adopt strict layering:

1. `cmd/api` remains wiring-only.
2. `internal/httpapi` handles transport concerns (routing, binding, response mapping).
3. `internal/shortener` contains business/domain logic.
4. `internal/store` exposes storage interfaces and implementations.

`main.go` will contain only bootstrap and lifecycle orchestration.

## Consequences

1. Better testability and clearer ownership per package.
2. Lower coupling when swapping storage backends (memory -> Postgres/Redis).
3. Slight upfront discipline cost when adding features.

## Alternatives considered

1. Feature-based folders mixing handler/service/store logic per endpoint.
2. Single package for all logic at this project stage.

## Implementation notes

1. Current structure already follows this decision.
2. Enforced by:
- `AGENTS.md`
- `cmd/api/AGENTS.md`
- `internal/httpapi/AGENTS.md`
- `internal/shortener/AGENTS.md`
- `internal/store/AGENTS.md`
