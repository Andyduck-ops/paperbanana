# Directory Structure

> How backend code is organized in PaperBanana.

**[PRE-TAURI]** This document describes the current web-app layout (Go HTTP server + React SPA). The directory structure will change significantly when the project migrates to Tauri + Go sidecar.

---

## Directory Layout

```
[project-root]/
├── cmd/
│   └── server/                 # Backend process entrypoint and server wiring
│       └── main.go             # Bootstrap, DI, graceful shutdown
├── internal/
│   ├── api/                    # Gin router, handlers, DTOs, middleware
│   │   ├── handlers/           # HTTP request handlers (generate, workspace, history, assets, batch, refine, provider, cancel)
│   │   ├── dto/                # Request/response data transfer objects
│   │   ├── middleware/         # Gin middleware (logger, auth, CORS, rate-limit, validation, audit, metrics)
│   │   └── router.go           # Route registration and service wiring
│   ├── application/            # Use cases, orchestrator, agents, config services
│   │   ├── orchestrator/       # Pipeline runner, batch runner, session tracking
│   │   ├── agents/             # Canonical pipeline agents
│   │   │   ├── retriever/      # Reference retrieval agent
│   │   │   ├── planner/        # Layout planning agent
│   │   │   ├── stylist/        # Style specification agent
│   │   │   ├── visualizer/     # Code generation / plot execution agent
│   │   │   ├── critic/         # Quality evaluation agent
│   │   │   └── modelselection/ # Model selection helper
│   │   ├── config/             # Provider config service, watcher, validator, startup sync
│   │   └── persistence/        # Application-level persistence services (workspace, history, asset, tx)
│   ├── domain/                 # Core entities, repository interfaces, shared contracts
│   │   ├── agent/              # BaseAgent interface, events, errors, types (pipeline contracts)
│   │   ├── llm/                # LLMClient interface and test doubles
│   │   ├── workspace/          # Workspace entities, repository interfaces
│   │   ├── config/             # Provider presets, API key entities, model defaults, events
│   │   └── crypto/             # Encryption service interface
│   ├── infrastructure/         # SQLite, LLM adapters, crypto, cache, node execution
│   │   ├── persistence/
│   │   │   └── sqlite/         # GORM models, repositories, bootstrap, tx_manager, snapshot store
│   │   ├── llm/                # LLM client factory, OpenAI, Gemini, Anthropic, OpenRouter adapters
│   │   │   ├── openai/         # OpenAI and OpenAI-compatible client
│   │   │   ├── gemini/         # Google Gemini client
│   │   │   ├── anthropic/      # Anthropic Claude client
│   │   │   ├── openrouter/     # OpenRouter client
│   │   │   └── models/         # Provider-specific model lists
│   │   ├── crypto/
│   │   │   ├── aesgcm/         # AES-256-GCM encryption service
│   │   │   └── keyderivation/  # Argon2id key derivation
│   │   ├── cache/redis/        # Optional Redis response cache
│   │   ├── assets/localstore/  # Opaque local asset store (filesystem)
│   │   ├── agentstate/         # Snapshot schema for pipeline state persistence
│   │   ├── nodes/httpnode/     # HTTP adapter for custom node execution
│   │   └── resilience/         # Resilient HTTP client (circuit breaker + retry)
│   └── agents/                 # Non-canonical agents used outside the main pipeline
│       ├── stylist/            # Legacy stylist with style guides
│       └── polish/             # Polish agent (side-path)
├── web/src/                    # React frontend source
│   ├── components/             # UI components
│   ├── hooks/                  # React hooks for data transport and state
│   ├── context/                # Shared React context providers
│   ├── pages/                  # Page-level components
│   ├── lib/                    # API client, utilities
│   └── types/                  # TypeScript type definitions
├── configs/                    # Runtime YAML config and custom-node definitions
│   ├── config.example.yaml
│   └── custom_nodes.example.yaml
├── data/                       # Local runtime database/assets and benchmark dataset
│   └── PaperBananaBench/       # Benchmark reference data (diagrams, plots)
├── test/                       # Backend golden/integration-style fixtures
├── testdata/                   # Prompt fixtures and legacy agent data
├── go.mod                      # Go module definition (github.com/paperbanana/paperbanana)
└── go.sum
```

---

## Ownership Boundaries

### HTTP vs Use Case

Request parsing, validation, and response formatting belong in `internal/api/handlers/`. Reusable business workflows belong in `internal/application/`. Handlers should not contain business logic; they translate between HTTP and domain types.

```go
// internal/api/handlers/generate.go - Handler parses HTTP, delegates to Runner
func (h *Handler) Generate(c *gin.Context) {
    var req GenerateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    handle, _, err := h.startRun(c.Request.Context(), req)
    // ...
}

// internal/application/orchestrator/runner.go - Runner contains pipeline orchestration
func (r *Runner) Start(ctx context.Context, input domainagent.AgentInput) (*RunHandle, error) {
    // Pipeline execution logic here
}
```

### Domain vs Infrastructure

Entities and repository interfaces are defined in `internal/domain/`. Concrete implementations live in `internal/infrastructure/`. The domain layer has zero imports from infrastructure.

```go
// internal/domain/workspace/repositories.go - Interface definition
type ProjectRepository interface {
    Create(ctx context.Context, project *Project) error
    GetByID(ctx context.Context, id string) (*Project, error)
    // ...
}

// internal/infrastructure/persistence/sqlite/workspace_repository.go - Implementation
type ProjectRepository struct { db *gorm.DB }
```

### Pipeline Agents

Canonical pipeline agents live in `internal/application/agents/<stage>/`. They implement the `domainagent.BaseAgent` interface and are composed by the orchestrator. Side-path agents (polish, legacy stylist) live in `internal/agents/`.

### Frontend State

- **Shared state**: `web/src/context/` -- React context providers
- **Data transport**: `web/src/hooks/` -- API calls and SSE handling
- **Presentation**: `web/src/components/` -- UI components only

---

## Where to Add New Code

| Task | Where | Example |
|------|-------|---------|
| New API endpoint | Handler in `internal/api/handlers/`, route in `internal/api/router.go`, use-case in `internal/application/` | `handlers/batch.go` + `router.go` batch routes |
| New pipeline stage | Agent in `internal/application/agents/<stage>/`, types in `internal/domain/agent/` | `agents/retriever/agent.go` |
| New persistence feature | Contract in `internal/domain/workspace/repositories.go`, implementation in `internal/infrastructure/persistence/sqlite/` | `SessionRepository` interface + `session_repository.go` |
| New LLM provider | Adapter in `internal/infrastructure/llm/<provider>/`, factory case in `internal/infrastructure/llm/factory.go` | `llm/openai/client.go` |
| New middleware | Implementation in `internal/api/middleware/`, mount in `internal/api/router.go` | `middleware/auth.go` |
| New config field | Add to `internal/config/config.go`, set default in `setDefaults()`, validate in `validate()` | `PersistenceConfig` |

---

## Key Files Quick Reference

| File | Purpose |
|------|---------|
| `cmd/server/main.go` | Server entrypoint, full DI wiring, graceful shutdown |
| `internal/api/router.go` | All HTTP routes, service composition |
| `internal/application/orchestrator/runner.go` | Pipeline execution engine |
| `internal/domain/agent/agent.go` | `BaseAgent` interface |
| `internal/domain/agent/types.go` | Pipeline types, stages, input/output |
| `internal/domain/workspace/repositories.go` | All repository interfaces |
| `internal/infrastructure/persistence/sqlite/bootstrap.go` | DB bootstrap and AutoMigrate |
| `internal/infrastructure/persistence/sqlite/models.go` | All GORM models |
| `internal/infrastructure/llm/factory.go` | LLM client creation per provider |
| `internal/config/config.go` | Configuration struct, Viper loading, validation |
