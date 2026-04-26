# PaperBanana Documentation

## Development Guidelines

All development guidelines have been consolidated into `.trellis/spec/` -- the single source of truth for AI-assisted development.

| Category | Location | Content |
|----------|----------|---------|
| Backend | [.trellis/spec/backend/](../.trellis/spec/backend/) | Architecture, conventions, directory structure, error handling, logging, database, quality, integrations |
| Frontend | [.trellis/spec/frontend/](../.trellis/spec/frontend/) | Directory structure, conventions, state management, component patterns, testing |
| Testing | [.trellis/spec/testing/](../.trellis/spec/testing/) | Strategy, backend testing, frontend testing, known gaps |
| ADRs | [.trellis/spec/adr/](../.trellis/spec/adr/) | Architecture Decision Records |
| Domain | [.trellis/spec/domain/](../.trellis/spec/domain/) | Domain model and bounded contexts |
| Guides | [.trellis/spec/guides/](../.trellis/spec/guides/) | Thinking guides, pre-Tauri reference, Tauri migration research |

## Product Documentation

| Document | Location | Content |
|----------|----------|---------|
| PRD | [prd/](./prd/) | Frontend workspace rebuild PRD |
| Product Brief | [ddd-tdd-architecture/product-brief.md](./ddd-tdd-architecture/product-brief.md) | DDD+TDD product overview |
| Requirements | [ddd-tdd-architecture/requirements/](./ddd-tdd-architecture/requirements/) | REQ-001 through REQ-005 |
| Epics | [ddd-tdd-architecture/epics/](./ddd-tdd-architecture/epics/) | EPIC-001 through EPIC-004 |

## Quick Start

```bash
# Backend
go run ./cmd/server --config ./configs/config.yaml

# Frontend
cd web && npm install && npm run dev

# Tests
go test ./...
cd web && npm run test:run
```

See [README.md](../README.md) for full setup instructions.

---

**Last Updated:** 2026-04-17
