---
id: REQ-001
type: functional
priority: Must
traces_to: [G-001]
status: draft
---

# REQ-001: Domain Model Alignment

**Priority**: Must

## Description

Align domain models between Go backend and TypeScript frontend to establish a unified ubiquitous language. All core domain entities must have consistent naming, structure, and semantics across both stacks.

## User Story

As a developer, I want domain entities to have the same names and structures in both backend and frontend, so that I can reason about the domain without mental translation.

## Acceptance Criteria

- [ ] Backend domain entities in `internal/domain/*` are documented with their TypeScript equivalents
- [ ] Naming discrepancies are resolved (e.g., `HistorySession` → `SessionRecord`)
- [ ] All core entities have consistent field naming (snake_case in Go, camelCase in TypeScript with proper mapping)
- [ ] Domain glossary is created defining each entity's purpose and relationships
- [ ] Alignment covers: Project, Folder, Visualization, VisualizationVersion, SessionRecord, Asset, Artifact

## Traces

- **Goal**: [G-001](../product-brief.md#goals--success-metrics)
- **Architecture**: [ADR-001](../architecture/ADR-001-naming-conventions.md)
