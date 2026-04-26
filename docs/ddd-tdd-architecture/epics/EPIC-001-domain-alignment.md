---
id: EPIC-001
title: Domain Model Alignment
priority: P0
status: Planned
requirements: [REQ-001, REQ-005]
---

# EPIC-001: Domain Model Alignment

## Objective

Establish unified domain language and document alignment between Go backend and TypeScript frontend.

## Stories

### Story 1.1: Domain Entity Mapping
**As a** developer, **I want** a documented mapping of all domain entities, **so that** I understand the relationship between backend and frontend types.

**Acceptance Criteria**:
- [ ] Create `docs/domain/entity-mapping.md` with table of all entities
- [ ] Document field-level mapping for each entity
- [ ] Identify and resolve naming conflicts

**Points**: 3

### Story 1.2: Naming Standardization
**As a** developer, **I want** consistent naming across stacks, **so that** domain concepts are unambiguous.

**Acceptance Criteria**:
- [ ] Rename `HistorySession` to `SessionRecord` in frontend
- [ ] Update all references in hooks and components
- [ ] Ensure no breaking changes in API contracts

**Points**: 5

### Story 1.3: Domain Glossary
**As a** new developer, **I want** a domain glossary, **so that** I can onboard quickly.

**Acceptance Criteria**:
- [ ] Create `docs/domain/glossary.md` with all domain terms
- [ ] Include purpose, relationships, and code references
- [ ] Add to main README

**Points**: 3

## Definition of Done

- [ ] All stories complete
- [ ] Documentation reviewed
- [ ] Team trained on new naming conventions

## Dependencies

- None (foundational)

## Estimated Effort

11 points (~1 week)
