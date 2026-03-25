# Codebase Stack Analysis

## Overview

PaperBanana is a full-stack image refinement workspace built with Go backend and React frontend. The stack is modern, minimal, and focused on AI/LLM integration for image processing tasks.

---

## Backend (Go)

### Language & Runtime
- **Go 1.23** (go.mod:3)
- CGO enabled for SQLite driver (Dockerfile:20)

### Web Framework
- **Gin v1.10.0** (go.mod:7) - HTTP web framework
- REST API architecture with JSON responses

### Database & Persistence
- **SQLite via glebarez/sqlite v1.11.0** (go.mod:8) - Pure Go SQLite driver (CGO-free)
- **GORM v1.31.1** (go.mod:21) - ORM for database operations
- Configuration: WAL mode optional, foreign key enforcement, busy timeout (config.go:58-70)
- Default path: `.paperbanana/paperbanana.db` (config.go:117)

### Caching
- **Redis v9.17.0** (go.mod:11) - Optional caching layer
- Disabled by default (config.yaml:31, config.go:111)
- Connection: `redis:7-alpine` in docker-compose (docker-compose.yml:34)

### LLM/AI Integrations
- **Google Generative AI Go SDK v0.20.1** (go.mod:9) - Gemini integration
- **OpenAI Go SDK v1.41.2** (go.mod:12) - via sashabaranov/go-openai
- **Anthropic Claude** - via OpenRouter proxy (config.yaml:23-27)
- **OpenRouter** - Multi-provider gateway (config.yaml:23-27)

### Configuration
- **Viper v1.19.0** (go.mod:14) - Configuration management
- YAML config files with environment variable expansion
- Environment prefix: `PAPERBANANA_` (config.go:85)

### Logging
- **Zap v1.27.0** (go.mod:16) - Structured, high-performance logging

### Utilities
- **google/uuid v1.6.0** (go.mod:10) - UUID generation
- **cenkalti/backoff/v4 v4.3.0** (go.mod:6) - Retry logic with exponential backoff
- **sony/gobreaker v1.0.0** (go.mod:13) - Circuit breaker pattern
- **golang.org/x/crypto v0.31.0** (go.mod:18) - Cryptographic operations
- **golang.org/x/sync v0.10.0** (go.mod:19) - Sync primitives (errgroup, singleflight)

### Testing
- **testify v1.9.0** (go.mod:15) - Testing assertions and mocks

---

## Frontend (React + TypeScript)

### Language & Build
- **TypeScript 5.9.3** (package.json:30)
- **Vite 6.0.0** (package.json:31) - Build tool and dev server
- Target: ES2022 (tsconfig.json:4)
- Module resolution: bundler mode (tsconfig.json:10)

### UI Framework
- **React 19.2.4** (package.json:16-17) - Latest React with concurrent features
- **Tailwind CSS 4.2.1** (package.json:29) - Utility-first CSS
- **@tailwindcss/vite v4.0.0** (package.json:21) - Vite plugin integration

### Internationalization
- **i18next v25.8.18** (package.json:15)
- **react-i18next v16.5.8** (package.json:18)

### Testing
- **Vitest 3.2.4** (package.json:32) - Vite-native test runner
- **@testing-library/react 16.3.2** (package.json:23)
- **@testing-library/jest-dom v6.6.3** (package.json:22)
- **@testing-library/user-event v14.6.1** (package.json:24)
- **jsdom v26.0.0** (package.json:28) - DOM simulation

### Frontend Structure
```
web/src/
  components/   - Reusable UI components
  hooks/        - Custom React hooks
  i18n/         - Internationalization resources
  lib/          - Utility functions
  pages/        - Page-level components
  themes/       - Theme configuration
  types/        - TypeScript type definitions
  test/         - Test setup and utilities
```

---

## Infrastructure

### Containerization
- **Docker** with multi-stage builds (Dockerfile)
- **docker-compose v3.8** for local development
- Base images: `golang:1.23-alpine`, `alpine:3.19`, `nginx:alpine`, `redis:7-alpine`

### Reverse Proxy
- **Nginx** (nginx.conf) for serving static frontend and API proxying

### Health Checks
- HTTP health endpoint at `/health` (Dockerfile:41-42, docker-compose.yml:15-19)

---

## LLM Provider Configuration

| Provider   | Default Model                    | Base URL                        |
|------------|----------------------------------|---------------------------------|
| Gemini     | gemini-2.0-flash-exp             | generativelanguage.googleapis.com |
| OpenAI     | gpt-4o                           | api.openai.com/v1               |
| Anthropic  | claude-3-5-sonnet-20241022       | api.anthropic.com/v1            |
| OpenRouter | anthropic/claude-3.5-sonnet      | openrouter.ai/api/v1            |

All providers have 60s timeout configured (config.yaml:12,17,22,27).

---

## Key Patterns

### Backend Architecture Pattern
- Layered architecture with Clean/Hexagonal influence
- `cmd/` - Entry points
- `internal/` - Private application code
  - `domain/` - Domain entities and interfaces
  - `application/` - Use cases and services
  - `infrastructure/` - External integrations (DB, LLM, assets)
  - `api/` - HTTP handlers and routing

### Configuration Pattern
- YAML config file with environment variable substitution
- Struct-based configuration with mapstructure tags
- Validation at startup

### Dependency Injection
- Manual DI via constructor functions
- Factory patterns for LLM clients (internal/infrastructure/llm/)

---

## Build Commands

### Backend
```bash
go run ./cmd/server --config ./configs/config.yaml
go build -ldflags="-w -s" -o /server ./cmd/server
```

### Frontend
```bash
cd web && npm install && npm run dev
npm run build  # tsc -b && vite build
```

---

## Recommendations

1. **Consider adding ESLint config** - package.json has lint script but no eslint config file visible
2. **Add swagger/OpenAPI** - For API documentation as the project grows
3. **Consider migration tool** - GORM AutoMigrate works, but explicit migrations would help with production deployments
4. **Add structured error types** - For consistent API error responses across handlers

---

## Evidence References

| Finding | Source |
|---------|--------|
| Go 1.23 | go.mod:3 |
| Gin framework | go.mod:7 |
| SQLite via glebarez | go.mod:8 |
| GORM ORM | go.mod:21 |
| Redis optional | go.mod:11, config.go:111 |
| React 19 | package.json:16-17 |
| Vite 6 | package.json:31 |
| Tailwind 4 | package.json:29 |
| TypeScript 5.9 | package.json:30 |
| Multi-stage Docker | Dockerfile:1-46 |
| LLM providers | config.yaml:5-27 |
