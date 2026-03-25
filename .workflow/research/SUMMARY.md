# Codebase Mapping Summary

## Executive Summary

**PaperBanana** is a full-stack scientific visualization workspace that transforms text descriptions into diagrams and plots via a 5-stage multi-agent LLM pipeline.

---

## Tech Stack Overview

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.23, Gin, GORM, SQLite |
| **Frontend** | React 19, TypeScript 5.9, Vite 6, Tailwind 4 |
| **LLM** | Gemini (default), OpenAI, Anthropic, OpenRouter |
| **Caching** | Redis (optional) |
| **Deployment** | Docker multi-stage, Nginx reverse proxy |

---

## Architecture Highlights

- **Style**: Clean Architecture with clear layer separation
- **Pipeline**: Retriever → Planner → Stylist → Visualizer → Critic
- **Key Patterns**: Repository, Factory, Unit of Work, Circuit Breaker
- **State Management**: Session snapshots for resume capability
- **Streaming**: SSE for real-time progress updates

---

## Feature Inventory

| Category | Count |
|----------|-------|
| API Endpoints | 30+ |
| Frontend Components | 15+ |
| React Hooks | 7 |
| Domain Entities | 17+ |
| Test Files | 42 |

**Core Features**:
- Multi-agent pipeline generation (diagram/plot modes)
- Retrieval system with 4 modes (auto/manual/random/none)
- Batch generation with parallel execution
- Image refinement standalone feature
- Provider management with API key encryption
- Workspace hierarchy (Project > Folder > Visualization > Version)

---

## Top 3 Concerns

### 1. Security Gaps
- No authentication/authorization
- No rate limiting
- File uploads lack content validation

### 2. Production Readiness
- No graceful shutdown
- WAL mode disabled by default
- No metrics/monitoring integration

### 3. Test Coverage
- No integration tests for external services
- Limited error path coverage
- Missing request ID correlation for tracing

---

## Recommendations for Next Steps

1. **Immediate**: Add authentication and rate limiting for production deployment
2. **High Priority**: Enable WAL mode, add graceful shutdown, implement metrics
3. **Medium Priority**: Add integration tests, request ID tracing, error path tests
4. **Future**: Consider WebSocket for bidirectional control, result caching

---

## Documents Generated

| File | Content |
|------|---------|
| `STACK.md` | Tech stack, dependencies, versions |
| `ARCHITECTURE.md` | Layer structure, data flow, patterns |
| `FEATURES.md` | Feature inventory, API endpoints, frontend |
| `PITFALLS.md` | Security, performance, test coverage gaps |
