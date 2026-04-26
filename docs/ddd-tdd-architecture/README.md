# PaperBanana DDD+TDD Full-Stack Architecture

This directory contains the complete architecture specification for unifying PaperBanana's full-stack development with Domain-Driven Design (DDD) and Test-Driven Development (TDD) methodologies.

## Executive Summary

**Goal**: Establish a unified DDD + TDD methodology across frontend (React+TypeScript) and backend (Go) to ensure:
- Consistent domain language across stacks
- API contracts as the single source of truth
- Automated TypeScript type generation from Go backend
- Standardized TDD workflows for both stacks

## Deliverables

### 1. Product Brief
**File**: [product-brief.md](product-brief.md)

Defines the vision, goals, and scope of the DDD+TDD initiative:
- Unified domain language (Goal G-001)
- Automated type synchronization (Goal G-002)
- OpenAPI contract specification (Goal G-003)
- TDD workflow adoption (Goal G-004)
- Improved developer experience (Goal G-005)

### 2. Requirements PRD
**Directory**: [requirements/](requirements/)

Detailed requirements specification:

| ID | Title | Priority |
|----|-------|----------|
| REQ-001 | Domain Model Alignment | Must |
| REQ-002 | OpenAPI Specification Generation | Must |
| REQ-003 | TypeScript Type Generation | Must |
| REQ-004 | TDD Workflow Definition | Must |
| REQ-005 | Domain Language Documentation | Must |

### 3. Architecture Design
**Directory**: [architecture/](architecture/)

Complete architecture specification including:

#### Architecture Decision Records (ADRs)

| ADR | Title | Key Decision |
|-----|-------|--------------|
| ADR-001 | Domain Naming Conventions | Backend naming as source of truth |
| ADR-002 | OpenAPI Generation Strategy | Use swaggo with Gin annotations |
| ADR-003 | TypeScript Generation Strategy | Use openapi-typescript |
| ADR-004 | TDD Workflow Architecture | Red-Green-Refactor for both stacks |

#### Key Architecture Elements

- **Domain Model**: Unified entities (Project, Folder, Visualization, SessionRecord, Asset)
- **API Contracts**: OpenAPI 3.0 specification as single source of truth
- **Type Synchronization**: Go structs → OpenAPI → TypeScript types
- **TDD Pattern**: Table-driven tests (Go), Component/hook tests (React)

### 4. Epics & Implementation Stories
**Directory**: [epics/](epics/)

Implementation roadmap organized into 4 epics:

| Epic | Title | Effort | Dependencies |
|------|-------|--------|--------------|
| EPIC-001 | Domain Model Alignment | 11 pts | None |
| EPIC-002 | OpenAPI Infrastructure | 16 pts | EPIC-001 |
| EPIC-003 | TypeScript Integration | 14 pts | EPIC-002 |
| EPIC-004 | TDD Workflow Implementation | 16 pts | None |

**Total Estimated Effort**: 57 story points (~4 weeks)

## Quick Start

### For Backend Developers

1. Read [ADR-002: OpenAPI Strategy](architecture/ADR-002-openapi-strategy.md)
2. Annotate handlers with swaggo comments
3. Run `make generate-api` to generate OpenAPI spec

### For Frontend Developers

1. Read [ADR-003: TypeScript Generation](architecture/ADR-003-typescript-generation.md)
2. Run `npm run generate-types` to sync types
3. Import from generated `api-generated.ts`

### For Tech Leads

1. Review [architecture/_index.md](architecture/_index.md) for full system design
2. Review [epics/_index.md](epics/_index.md) for implementation planning
3. Set up CI/CD integration per [REQ-002](requirements/REQ-002-openapi-generation.md)

## Implementation Phases

### Phase 1: Foundation (Week 1)
- Complete EPIC-001: Domain alignment documentation
- Start EPIC-002: OpenAPI tooling setup

### Phase 2: Integration (Week 2)
- Complete EPIC-002: Handler annotation coverage
- Start EPIC-003: TypeScript generation pipeline

### Phase 3: Adoption (Week 3-4)
- Complete EPIC-003: Frontend type migration
- Complete EPIC-004: TDD workflow rollout

## Key Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Naming Consistency | 100% | Domain glossary coverage |
| OpenAPI Coverage | 100% | All endpoints documented |
| Type Synchronization | Auto | CI pipeline integration |
| Test Coverage | 80% BE, 70% FE | Coverage reports |
| Onboarding Time | < 30 min | Time to understand domain |

## Tools & Technologies

### Backend (Go)
- **OpenAPI Generation**: [swaggo/swag](https://github.com/swaggo/swag)
- **Testing**: Standard `go test` with table-driven patterns
- **Mocking**: Interface-based mocks

### Frontend (TypeScript)
- **Type Generation**: [openapi-typescript](https://github.com/drwpow/openapi-typescript)
- **Testing**: [Vitest](https://vitest.dev/) with React Testing Library
- **API Mocking**: [MSW](https://mswjs.io/)

## Session Information

- **Session ID**: TLV4-paperbanana-ddd-tdd-20260408
- **Skill**: team-lifecycle-v4
- **Pipeline**: spec-only
- **Generated**: 2026-04-08

## References

- Backend DDD Structure: `internal/domain/{agent,config,crypto,llm,workspace}/`
- Frontend Types: `web/src/types/api.ts`
- Frontend API: `web/src/lib/api.ts`
