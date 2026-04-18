---
id: ADR-001
status: Accepted
traces_to: [REQ-001, REQ-005]
date: 2026-04-08
---

# ADR-001: Domain Naming Conventions

## Context

The frontend and backend currently use different naming for the same domain concepts:
- Frontend: `HistorySession` vs Backend: `SessionRecord`
- Frontend: `Visualization` vs Backend: Same (aligned)
- Frontend: camelCase fields vs Backend: snake_case JSON

This inconsistency causes confusion and bugs when developers switch between stacks.

## Decision

Standardize on **backend naming as the source of truth** with automatic transformation:

1. **Entity Names**: Use backend domain entity names in TypeScript
   - `SessionRecord` (not `HistorySession`)
   - `VisualizationVersion` (not abbreviated)

2. **Field Names**: Backend uses snake_case in JSON; frontend transforms to camelCase
   - Backend: `created_at` → Frontend: `createdAt`
   - Transformation handled at API layer, not in domain types

3. **Generated Types**: Preserve backend names in generated code

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Backend as source | Single source of truth, Go naming idiomatic | Requires frontend adaptation |
| Frontend as source | Matches React conventions | Unnatural for Go, backwards |
| Independent naming | Each stack natural | Maximum confusion, no alignment |

## Consequences

- **Positive**: Clear single source of truth, reduced cognitive load
- **Negative**: Frontend needs to update some existing type references
- **Risks**: Breaking changes in frontend type imports

> **Pre-Tauri Note**: This ADR was written for the current web-app architecture. Naming conventions remain valid in the Tauri migration.
