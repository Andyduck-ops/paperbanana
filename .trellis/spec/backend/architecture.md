# Architecture

> System architecture overview for PaperBanana.

**[PRE-TAURI]** This document describes the current web-app architecture (Go HTTP server + React SPA). The architecture will change significantly when the project migrates to Tauri + Go sidecar.

---

## Pattern Overview

PaperBanana is a **layered monolith** with a clear separation between domain logic and infrastructure concerns. It follows a modified hexagonal/ports-and-adapters pattern where the domain layer defines interfaces (ports) and the infrastructure layer provides implementations (adapters).

```
┌─────────────────────────────────────────────────────────┐
│  Frontend (React SPA)                                   │
│  web/src/                                               │
├─────────────────────────────────────────────────────────┤
│  HTTP/API Layer (Gin)                                   │
│  internal/api/                                          │
├─────────────────────────────────────────────────────────┤
│  Application Layer (Use Cases, Orchestrator, Agents)    │
│  internal/application/                                  │
├─────────────────────────────────────────────────────────┤
│  Domain Layer (Entities, Interfaces, Contracts)         │
│  internal/domain/                                       │
├─────────────────────────────────────────────────────────┤
│  Infrastructure Layer (SQLite, LLM, Crypto, Cache)      │
│  internal/infrastructure/                               │
└─────────────────────────────────────────────────────────┘
```

---

## Layer Descriptions

### Frontend UI Layer

- **Location**: `web/src/`
- **Framework**: React with TypeScript, Vite build tooling
- **State**: React context providers + custom hooks for data transport
- **Transport**: REST API calls (`web/src/lib/api.ts`) + SSE for streaming events
- **Key pattern**: Hooks encapsulate all API/SSE logic; components are presentation-only

### HTTP/API Layer

- **Location**: `internal/api/`
- **Framework**: Gin (`github.com/gin-gonic/gin`)
- **Responsibilities**: Request parsing, input validation, DTO mapping, response formatting, error-to-HTTP mapping
- **Does NOT contain**: Business logic, direct database access, direct LLM calls
- **Middleware stack**: Recovery, Logger, CORS, Metrics (optional: Auth, RateLimit -- defined but not mounted)

### Application Layer

- **Location**: `internal/application/`
- **Components**:
  - `orchestrator/` -- Pipeline execution engine (`Runner`, `BatchRunner`)
  - `agents/` -- Canonical pipeline stage implementations (Retriever, Planner, Stylist, Visualizer, Critic)
  - `config/` -- Provider configuration service, hot-reload watcher, validation, startup sync
  - `persistence/` -- Application-level service facades (WorkspaceService, HistoryService, AssetService)
- **Key pattern**: This is where business workflows are composed. The `Runner` orchestrates agents; the services compose repository calls with business rules.

### Domain Layer

- **Location**: `internal/domain/`
- **Components**:
  - `agent/` -- `BaseAgent` interface, event types, error codes, pipeline types
  - `llm/` -- `LLMClient` interface
  - `workspace/` -- Entities (Project, Folder, Visualization, Version, Session, Asset) and repository interfaces
  - `config/` -- Provider presets, API key entities, model defaults, event types
  - `crypto/` -- Encryption service interface
- **Key rule**: Zero imports from `internal/infrastructure/` or `internal/api/`. The domain layer is pure Go with no framework dependencies.

### Infrastructure Layer

- **Location**: `internal/infrastructure/`
- **Components**:
  - `persistence/sqlite/` -- GORM models, all repository implementations, bootstrap, tx_manager
  - `llm/` -- LLM client factory, provider-specific adapters (OpenAI, Gemini, Anthropic, OpenRouter), caching, resilience
  - `crypto/` -- AES-256-GCM encryption, Argon2id key derivation
  - `cache/redis/` -- Optional Redis response cache
  - `assets/localstore/` -- Filesystem asset store with path traversal protection
  - `agentstate/` -- Snapshot schema for pipeline state
  - `nodes/httpnode/` -- HTTP adapter for custom node execution
  - `resilience/` -- Circuit breaker + retry HTTP client
- **Key rule**: Implements interfaces defined in `internal/domain/`.

---

## Data Flow

### Single Generate Flow

```
User Prompt
    │
    ▼
POST /api/v1/generate (or /generate/stream)
    │
    ▼
handlers.Generate (parse, validate, build AgentInput)
    │
    ▼
Runner.Start(input) ──► RunHandle with Event channel
    │
    ▼
Execute pipeline stages sequentially:
  Retriever → Planner → Stylist → Visualizer → Critic
    │                         │
    │                         ▼
    │                   Each stage: Initialize → Execute → Cleanup
    │                   On failure: emit stage_failed event
    │                   On success: persist snapshot, emit stage_completed
    │
    ▼
Persist artifacts via AssetServiceAdapter
    │
    ▼
Return GenerateResponse (sync) or SSE events (stream)
```

### Batch Generate Flow

```
POST /api/v1/generate/batch
    │
    ▼
BatchHandler.StreamBatchGenerate
    │
    ▼
BatchRunner creates per-candidate agents via AgentFactory
    │
    ▼
Each candidate runs the full pipeline concurrently
    │
    ▼
SSE stream emits: batch_start → candidate_start → stage events → candidate_complete → batch_complete
```

### Refine Flow

```
POST /api/v1/refine (image + prompt)
    │
    ▼
RefineHandler.Refine
    │
    ▼
LLMClient.Generate (single LLM call)
    │
    ▼
SessionSaver persists the result
    │
    ▼
Return refined image/artifacts
```

### Workspace/History Flow

```
CRUD endpoints → WorkspaceService → TxManager.RunInTx → Repository implementations
```

All workspace operations are wrapped in transactions via `TxManager`. The service layer enforces business rules (e.g., project scoping, soft deletes).

### Config Flow

```
Frontend Settings UI
    │
    ▼
Provider CRUD endpoints (POST/PUT/DELETE /api/v1/providers)
    │
    ▼
configservice.Service → ProviderRepository + APIKeyRepository
    │
    ▼
Config SSE stream notifies frontend of changes
    │
    ▼
Watcher fires events → connected clients receive updates
```

---

## Key Abstractions

### Runner (`internal/application/orchestrator/runner.go`)

The `Runner` is the pipeline execution engine. It:
- Holds a map of `StageName → BaseAgent`
- Executes stages sequentially in canonical pipeline order
- Emits events through a buffered channel
- Supports resume from snapshots via `SnapshotStore`
- Supports graceful degradation for non-critical stages
- Supports configurable stage timeouts

```go
type Runner struct {
    agents          map[domainagent.StageName]domainagent.BaseAgent
    pipeline        []domainagent.StageName
    snapshotStore   SnapshotStore
    stageTimeouts   StageTimeouts
    gracefulDegrade bool
}
```

### BaseAgent (`internal/domain/agent/agent.go`)

Every pipeline stage implements this interface:

```go
type BaseAgent interface {
    Initialize(ctx context.Context) error
    Execute(ctx context.Context, input AgentInput) (AgentOutput, error)
    Cleanup(ctx context.Context) error
    GetState() AgentState
    RestoreState(state AgentState) error
}
```

### RuntimeClient (`internal/infrastructure/llm/client_manager.go`)

Manages LLM client instances with different purposes (query vs. generation) and provider resolution. Wraps the raw LLM client with optional caching and provider alias support.

### Workspace/History Services (`internal/application/persistence/`)

Application-level facades that compose repository calls with business rules:
- `WorkspaceService` -- project, folder, visualization CRUD with hierarchy validation
- `HistoryService` -- session listing, version history
- `AssetService` -- asset registration, retrieval, filesystem storage coordination

### PersistentSnapshotStore (`internal/infrastructure/persistence/sqlite/snapshot_store.go`)

Bridges the orchestrator's `SnapshotStore` interface with the `SessionRepository`. Enables pipeline resume after server restart by persisting full session state as JSON in the sessions table.

### Frontend Hooks (`web/src/hooks/`)

React hooks encapsulate all API and SSE logic. Components call hooks; hooks manage transport and state. Key hooks:
- `useGenerate` -- Generation flow with SSE
- `useRefine` -- Image refinement
- `useHistory` -- Session history queries
- `useKeyboardShortcuts` -- Global keyboard shortcuts

---

## Entry Points

### Backend

| Entry Point | File | Purpose |
|-------------|------|---------|
| Server main | `cmd/server/main.go` | Bootstrap, DI, graceful shutdown |
| Router setup | `internal/api/router.go` | Route registration, middleware, service composition |
| Config load | `internal/config/config.go` | Viper config loading, env expansion, validation |

### Frontend

| Entry Point | File | Purpose |
|-------------|------|---------|
| App root | `web/src/main.tsx` | React mount, context providers |
| API client | `web/src/lib/api.ts` | HTTP + SSE transport layer |

---

## Error Handling Patterns

See [Error Handling](./error-handling.md) for full details. Key patterns:

- **Wrapping**: `fmt.Errorf("context: %w", err)` throughout
- **Sentinel errors**: Package-level `var ErrXxx = errors.New(...)` for cross-package comparison
- **Classification**: `domainagent.ClassifyError()` maps raw errors to `ErrorCode` for pipeline events
- **HTTP mapping**: Handlers use `errors.Is` against sentinels to choose HTTP status codes
- **Cleanup**: `errors.Join` for combining multiple failures

---

## Cross-Cutting Concerns

### Logging

Zap structured logging at all boundaries. See [Logging Guidelines](./logging-guidelines.md).

### Validation

Input validation is performed at two levels:
1. **Handler level**: Request DTO validation (field lengths, required fields, format checks) in `validateGenerateRequest` etc.
2. **Config level**: Configuration validation in `internal/config/validation.go` and `internal/application/config/validator.go`

### Authentication

Auth middleware exists (`internal/api/middleware/auth.go`) with API key validation using constant-time comparison, but it is **not mounted** in the router. All endpoints are currently open.

### Persistence

All database access goes through the repository pattern. Transactions are managed by `TxManager`. See [Database Guidelines](./database-guidelines.md).

### Caching

Optional Redis cache for LLM responses. Configured via `cache.redis.*` settings. When enabled, `CachedClient` wraps the raw LLM client with a Redis-backed response cache.

### Node Execution

Custom nodes defined in `configs/custom_nodes.yaml` are executed via HTTP. The `httpnode.Adapter` sends requests to configured endpoints with resilient HTTP clients (circuit breaker + exponential backoff).

### Configuration Hot-Reload

Provider configuration changes are broadcast to connected frontend clients via SSE (`/api/v1/config/stream`). The `configservice.Watcher` observes changes and pushes events.
