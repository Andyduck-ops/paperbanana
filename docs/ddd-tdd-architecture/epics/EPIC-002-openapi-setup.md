---
id: EPIC-002
title: OpenAPI Infrastructure
priority: P0
status: Planned
requirements: [REQ-002]
---

# EPIC-002: OpenAPI Infrastructure

## Objective

Set up automated OpenAPI specification generation from Go backend code.

## Stories

### Story 2.1: Tooling Setup
**As a** backend developer, **I want** swaggo integrated, **so that** I can generate OpenAPI specs.

**Acceptance Criteria**:
- [ ] Install swaggo CLI tool
- [ ] Add `make generate-api` command
- [ ] Create initial `docs/openapi.yaml` generation

**Points**: 3

### Story 2.2: Handler Annotations
**As a** backend developer, **I want** all handlers annotated, **so that** the OpenAPI spec is complete.

**Acceptance Criteria**:
- [ ] Annotate all handlers in `internal/api/handlers/`
- [ ] Define DTO schemas with examples
- [ ] Document error responses

**Points**: 8

### Story 2.3: CI Integration
**As a** DevOps engineer, **I want** OpenAPI validation in CI, **so that** specs are always valid.

**Acceptance Criteria**:
- [ ] Add spec generation to CI pipeline
- [ ] Add spec validation step
- [ ] Fail build on validation errors

**Points**: 3

### Story 2.4: Documentation Hosting
**As a** developer, **I want** the OpenAPI spec hosted, **so that** I can reference it easily.

**Acceptance Criteria**:
- [ ] Set up Swagger UI or ReDoc
- [ ] Make accessible at `/docs/api`
- [ ] Link from developer docs

**Points**: 2

## Definition of Done

- [ ] All handlers annotated
- [ ] CI validates spec on every PR
- [ ] Documentation hosted and accessible

## Dependencies

- EPIC-001 (naming alignment)

## Estimated Effort

16 points (~1.5 weeks)
