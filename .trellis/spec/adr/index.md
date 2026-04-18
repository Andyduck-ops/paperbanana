# Architecture Decision Records

Index of all ADRs for the PaperBanana project.

## Records

| ADR | Title | Status | Traces To |
|-----|-------|--------|-----------|
| [ADR-001](ADR-001-naming-conventions.md) | Domain Naming Conventions | Accepted | REQ-001, REQ-005 |
| [ADR-002](ADR-002-openapi-strategy.md) | OpenAPI Generation Strategy | Accepted | REQ-002 |
| [ADR-003](ADR-003-typescript-generation.md) | TypeScript Generation Strategy | Accepted | REQ-003 |
| [ADR-004](ADR-004-tdd-workflow.md) | TDD Workflow Architecture | Accepted | REQ-004 |

## Related Specs

- [Testing Strategy](../testing/strategy.md) -- implements ADR-004 patterns
- [Domain Model](../domain/model.md) -- relies on ADR-001 naming conventions
- [Frontend Testing](../testing/frontend-testing.md) -- applies ADR-004 TDD workflow
- [Backend Testing](../testing/backend-testing.md) -- applies ADR-004 TDD workflow
