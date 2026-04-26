---
id: EPIC-004
title: TDD Workflow Implementation
priority: P1
status: Planned
requirements: [REQ-004]
---

# EPIC-004: TDD Workflow Implementation

## Objective

Establish consistent TDD practices across backend and frontend with clear guidelines and tooling.

## Stories

### Story 4.1: Backend TDD Guidelines
**As a** backend developer, **I want** TDD guidelines, **so that** I can write tests consistently.

**Acceptance Criteria**:
- [ ] Create `docs/testing/backend-tdd.md`
- [ ] Document table-driven test patterns
- [ ] Include mock and golden file examples

**Points**: 3

### Story 4.2: Frontend TDD Guidelines
**As a** frontend developer, **I want** TDD guidelines, **so that** I can test components effectively.

**Acceptance Criteria**:
- [ ] Create `docs/testing/frontend-tdd.md`
- [ ] Document component and hook testing
- [ ] Include MSW setup for API mocking

**Points**: 3

### Story 4.3: Coverage Configuration
**As a** tech lead, **I want** coverage targets enforced, **so that** quality is maintained.

**Acceptance Criteria**:
- [ ] Configure coverage thresholds: backend 80%, frontend 70%
- [ ] Add coverage checks to CI
- [ ] Generate coverage reports

**Points**: 3

### Story 4.4: Pre-commit Hooks
**As a** developer, **I want** tests to run before commit, **so that** I catch issues early.

**Acceptance Criteria**:
- [ ] Configure husky + lint-staged
- [ ] Run relevant tests on commit
- [ ] Allow bypass with `--no-verify` for WIP

**Points**: 2

### Story 4.5: Example Test Suite
**As a** developer, **I want** example tests, **so that** I can learn by example.

**Acceptance Criteria**:
- [ ] Create example backend tests with TDD pattern
- [ ] Create example frontend tests with TDD pattern
- [ ] Add to documentation

**Points**: 5

## Definition of Done

- [ ] Guidelines documented and reviewed
- [ ] Coverage targets enforced in CI
- [ ] Team trained on TDD workflow

## Dependencies

- None (can run in parallel)

## Estimated Effort

16 points (~1.5 weeks)
