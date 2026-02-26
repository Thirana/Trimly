# Architecture Decision Records (ADRs)

This folder stores key staff-level engineering decisions for the URL shortener.
Each ADR explains context, decision, consequences, and implementation status.

## Status values

1. `proposed`
2. `accepted`
3. `deprecated`
4. `superseded`

## Index

1. `ADR-0001-architecture-boundaries-and-thin-main.md` (`accepted`)
2. `ADR-0002-service-store-interface-and-in-memory-first.md` (`accepted`)
3. `ADR-0003-redirect-contract-and-302-default.md` (`accepted`)
4. `ADR-0004-validation-at-http-boundary-and-service-layer.md` (`accepted`)
5. `ADR-0005-phase-2-correctness-contract.md` (`accepted`)

## Rules

1. Use increasing ADR numbers (`ADR-0006-*`, `ADR-0007-*`, ...).
2. Do not rewrite accepted ADR history; add a new ADR to supersede old decisions.
3. If a key technical decision changes, add/update ADRs in the same change.
