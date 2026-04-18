---
id: ADR-003
status: Accepted
traces_to: [REQ-003]
date: 2026-04-08
---

# ADR-003: TypeScript Generation Strategy

## Context

TypeScript types must be generated from OpenAPI to ensure synchronization. We need a tool that:
- Generates clean, readable TypeScript
- Supports OpenAPI 3.0
- Integrates into build pipeline

## Decision

Use **openapi-typescript** for type generation:

1. **Command**: `npx openapi-typescript openapi.yaml -o src/types/api-generated.ts`
2. **Output**: Single file with all API types
3. **Integration**: Pre-build step in package.json scripts
4. **Re-export**: Create `src/types/api.ts` that re-exports and extends generated types

Generation workflow:
```bash
# In CI or local development
cd web
npx openapi-typescript ../docs/openapi.yaml -o src/types/api-generated.ts
```

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| openapi-typescript | Clean output, active maintenance | Requires Node runtime |
| openapi-generator | Multiple language support | Heavier, more complex |
| orval | Generates fetch clients too | More opinionated |

## Consequences

- **Positive**: Clean TypeScript, flexible integration
- **Negative**: Additional build step
- **Risks**: Generation failures block builds

> **Pre-Tauri Note**: This ADR was written for the current web-app architecture. Naming conventions remain valid in the Tauri migration.
