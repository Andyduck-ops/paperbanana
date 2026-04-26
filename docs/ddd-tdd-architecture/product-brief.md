---
session_id: TLV4-paperbanana-ddd-tdd-20260408
phase: 2
document_type: product-brief
status: draft
generated_at: 2026-04-08T21:35:00+08:00
stepsCompleted: ["discovery", "analysis"]
version: 1
dependencies:
  - spec-config.json
  - discovery-context.json
---

# Product Brief: PaperBanana DDD+TDD Full-Stack Architecture

Establish a unified Domain-Driven Design (DDD) and Test-Driven Development (TDD) methodology across PaperBanana's Go backend and TypeScript frontend to ensure consistent domain language, API contracts, and development workflows.

## Vision

PaperBanana becomes a reference implementation for full-stack DDD+TDD in AI-powered tools, where domain models flow seamlessly from backend to frontend, API contracts are the single source of truth, and every feature is built test-first with confidence.

## Problem Statement

### Current Situation

- **Disconnected domains**: Frontend and backend maintain separate, manually-synced type definitions
- **No API contract**: No OpenAPI/Swagger specification exists to formalize API boundaries
- **Inconsistent terminology**: Domain entities have different names across stacks (e.g., `HistorySession` vs `SessionRecord`)
- **Ad-hoc testing**: No unified TDD workflow; testing practices vary between frontend (Vitest) and backend (go test)
- **Manual type maintenance**: TypeScript types are manually defined and drift from Go structs over time

### Impact

- **Development friction**: Developers must manually update types in two places, causing errors and inconsistencies
- **API mismatches**: Frontend expects fields that backend doesn't provide, causing runtime failures
- **Onboarding overhead**: New developers lack clear documentation of domain models and API contracts
- **Testing gaps**: Features ship without comprehensive test coverage due to inconsistent TDD practices

## Target Users

### Backend Developer
- **Role**: Implements domain logic in Go
- **Needs**: Clear domain boundaries, automated type generation for API contracts
- **Pain Points**: Frontend breaks when API changes; no clear contract to code against
- **Success Criteria**: Can generate OpenAPI spec from code; types stay synchronized automatically

### Frontend Developer
- **Role**: Builds React+TypeScript UI components
- **Needs**: Type-safe API clients, accurate type definitions
- **Pain Points**: TypeScript types don't match actual API responses; runtime type errors
- **Success Criteria**: Types are auto-generated from OpenAPI; full IntelliSense support

### Architecture Team
- **Role**: Maintains coding standards and reviews architecture
- **Needs**: Unified domain language, consistent TDD workflow across teams
- **Pain Points**: Code reviews catch API mismatches; no standardized testing approach
- **Success Criteria**: Clear DDD boundaries; TDD is mandatory and enforced

## Goals & Success Metrics

| Goal ID | Goal | Success Metric | Target |
|---------|------|----------------|--------|
| G-001 | Unified Domain Language | Naming consistency score | 100% alignment on core entities |
| G-002 | Automated Type Synchronization | Type sync automation | OpenAPI → TypeScript generation pipeline |
| G-003 | API Contract Specification | OpenAPI coverage | 100% of API endpoints documented |
| G-004 | TDD Workflow Adoption | Test coverage | Backend: 80%, Frontend: 70% |
| G-005 | Developer Experience | Onboarding time | < 30 minutes to understand domain model |

## Scope

### In Scope
- Domain model alignment and documentation
- OpenAPI specification generation from Go backend
- TypeScript type generation from OpenAPI
- TDD workflow guidelines for both stacks
- Shared domain language glossary
- API contract testing strategy

### Out of Scope
- Complete codebase refactoring (incremental adoption)
- Migration away from existing tech stack
- Third-party API standardization
- Deployment pipeline changes

### Assumptions
- Team is familiar with DDD and TDD concepts
- Go backend can be extended with OpenAPI annotations
- Frontend can consume generated TypeScript types
- Testing infrastructure (Vitest, go test) is already in place

## Competitive Landscape

| Aspect | Current State | Proposed Solution | Advantage |
|--------|--------------|-------------------|-----------|
| Type Safety | Manual sync, error-prone | OpenAPI-generated types | Zero drift, compile-time safety |
| API Documentation | None/Outdated | Living OpenAPI spec from code | Always up-to-date |
| Domain Understanding | Tribal knowledge | Documented ubiquitous language | Faster onboarding |
| Testing | Inconsistent | Unified TDD workflow | Higher confidence, fewer bugs |

## Constraints & Dependencies

### Technical Constraints
- Must maintain backward compatibility with existing API endpoints
- Type generation must work in CI/CD pipeline
- DDD boundaries must respect existing Clean Architecture structure

### Business Constraints
- No dedicated sprint for refactoring; incremental adoption only
- Must not disrupt active feature development

### Dependencies
- OpenAPI code generation tools (swaggo, openapi-typescript)
- CI/CD pipeline integration
- Team training on DDD/TDD practices

## Multi-Perspective Synthesis

### Product Perspective
The unified DDD+TDD approach will significantly reduce API integration bugs and accelerate feature development by establishing clear contracts between frontend and backend teams.

### Technical Perspective
Go's strong typing and struct tags make it ideal for OpenAPI generation. TypeScript's type system can fully represent Go types with proper generation configuration. Existing Clean Architecture provides good DDD boundaries.

### User Perspective
Developers will experience fewer runtime errors, better IDE support, and clearer documentation. The unified language reduces cognitive load when switching between frontend and backend.

### Convergent Themes
- Type safety is paramount for developer productivity
- Documentation must be living, not static
- DDD boundaries already partially exist in Clean Architecture

### Conflicting Views
- **Annotation overhead**: Some prefer minimal code annotations vs comprehensive OpenAPI generation
- **Test coverage targets**: Frontend team concerned about 70% coverage target vs practical value

## Open Questions

- [ ] Which OpenAPI generation tool to adopt? (swaggo vs manual spec)
- [ ] How to handle breaking API changes during transition?
- [ ] What's the incremental adoption strategy for existing endpoints?

## References

- Derived from: [spec-config.json](spec-config.json), [discovery-context.json](discovery-context.json)
- Next: [Requirements PRD](requirements/_index.md)
