---
id: REQ-004
type: functional
priority: Must
traces_to: [G-004]
status: draft
---

# REQ-004: TDD Workflow Definition

**Priority**: Must

## Description

Establish a unified Test-Driven Development workflow for both frontend (TypeScript/React) and backend (Go) that follows the Red-Green-Refactor cycle.

## User Story

As a developer, I want clear TDD guidelines for my stack, so that I can write tests first with confidence and maintain consistent code quality.

## Acceptance Criteria

- [ ] TDD workflow guidelines document is created
- [ ] Backend TDD workflow: table-driven tests, mock interfaces, golden files
- [ ] Frontend TDD workflow: component tests, hook tests, MSW for API mocking
- [ ] Testing directory structure is standardized for both stacks
- [ ] Coverage targets defined: Backend 80%, Frontend 70%
- [ ] Pre-commit hooks run tests before allowing commits

## Traces

- **Goal**: [G-004](../product-brief.md#goals--success-metrics)
- **Architecture**: [ADR-004](../architecture/ADR-004-tdd-workflow.md)
