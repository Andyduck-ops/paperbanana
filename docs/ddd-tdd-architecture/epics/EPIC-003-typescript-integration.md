---
id: EPIC-003
title: TypeScript Generation Integration
priority: P0
status: Planned
requirements: [REQ-003]
---

# EPIC-003: TypeScript Generation Integration

## Objective

Integrate automated TypeScript type generation from OpenAPI specification.

## Stories

### Story 3.1: Generation Pipeline
**As a** frontend developer, **I want** types auto-generated, **so that** they stay synchronized.

**Acceptance Criteria**:
- [ ] Install openapi-typescript
- [ ] Add `npm run generate-types` script
- [ ] Configure generation options

**Points**: 3

### Story 3.2: Type Migration
**As a** frontend developer, **I want** to use generated types, **so that** I have type safety.

**Acceptance Criteria**:
- [ ] Generate initial `api-generated.ts`
- [ ] Create `api.ts` that re-exports generated types
- [ ] Migrate existing manual types to use generated

**Points**: 8

### Story 3.3: Build Integration
**As a** developer, **I want** types regenerated on build, **so that** they're always fresh.

**Acceptance Criteria**:
- [ ] Add type generation to build script
- [ ] Handle generation failures gracefully
- [ ] Document manual regeneration process

**Points**: 3

## Definition of Done

- [ ] Types generated automatically
- [ ] Frontend uses generated types exclusively
- [ ] Build fails on type mismatches

## Dependencies

- EPIC-002 (OpenAPI spec)

## Estimated Effort

14 points (~1.5 weeks)
