---
id: REQ-003
type: functional
priority: Must
traces_to: [G-002]
status: draft
---

# REQ-003: TypeScript Type Generation

**Priority**: Must

## Description

Generate TypeScript type definitions from the OpenAPI specification to ensure frontend types are always synchronized with the backend API contract.

## User Story

As a frontend developer, I want TypeScript types to be auto-generated from the OpenAPI spec, so that I have compile-time type safety that matches the actual API.

## Acceptance Criteria

- [ ] TypeScript types are generated from OpenAPI spec using openapi-typescript
- [ ] Generated types cover all API request/response schemas
- [ ] Generated types are committed to version control or regenerated on build
- [ ] Frontend uses generated types exclusively (no manual type definitions for API entities)
- [ ] Generation script is documented and integrated into dev/build workflow
- [ ] Type generation is idempotent (same output for same input)

## Traces

- **Goal**: [G-002](../product-brief.md#goals--success-metrics)
- **Architecture**: [ADR-003](../architecture/ADR-003-typescript-generation.md)
