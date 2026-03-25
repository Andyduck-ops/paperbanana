# Codebase Architecture Analysis

## Overview

PaperBanana implements a **Clean Architecture** pattern with clear layer separation. The codebase follows the hexagonal architecture style where the domain core is independent of infrastructure concerns. The application is a visualization generation system that uses an LLM-powered multi-agent pipeline to produce diagrams and plots from natural language descriptions.

**Architecture Style**: Clean Architecture / Hexagonal Architecture
**Primary Language**: Go 1.23
**Web Framework**: Gin
**Database**: SQLite (via GORM with glebarez/sqlite driver)
**Key Pattern**: Multi-agent pipeline orchestration

---

## Layer Structure

### Domain Layer (`internal/domain/`)

The innermost layer containing pure business logic with no external dependencies.

**Submodules**:

| Module | Purpose | Key Types |
|--------|---------|-----------|
| `agent/` | Agent lifecycle and state management | `BaseAgent`, `AgentInput`, `AgentOutput`, `AgentState`, `SessionState` |
| `llm/` | LLM client abstraction | `LLMClient`, `GenerateRequest`, `GenerateResponse` |
| `workspace/` | Workspace entities and repository interfaces | `Project`, `Folder`, `Visualization`, `VisualizationVersion`, `SessionRecord`, `Asset` |
| `config/` | Configuration domain types | `Provider`, `APIKey` |
| `crypto/` | Encryption service interfaces | `EncryptionService` |

**Key Abstractions**:

- `BaseAgent` interface (`internal/domain/agent/agent.go:6-12`): Defines the agent lifecycle contract
  ```go
  type BaseAgent interface {
      Initialize(ctx context.Context) error
      Execute(ctx context.Context, input AgentInput) (AgentOutput, error)
      Cleanup(ctx context.Context) error
      GetState() AgentState
      RestoreState(state AgentState) error
  }
  ```

- `LLMClient` interface (`internal/domain/llm/client.go:9-13`): LLM provider abstraction
  ```go
  type LLMClient interface {
      Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
      GenerateStream(ctx context.Context, req GenerateRequest) (<-chan StreamChunk, <-chan error)
      Provider() string
  }
  ```

- Repository interfaces in `internal/domain/workspace/repositories.go`: `ProjectRepository`, `FolderRepository`, `VisualizationRepository`, `VersionRepository`, `SessionRepository`, `AssetRepository`

### Application Layer (`internal/application/`)

Contains use cases, orchestration, and application-specific business rules.

**Submodules**:

| Module | Purpose |
|--------|---------|
| `orchestrator/` | Pipeline execution engine, session management, batch processing |
| `agents/` | Concrete agent implementations (retriever, planner, stylist, visualizer, critic) |
| `persistence/` | Application-level services wrapping repositories |
| `config/` | Configuration service with validation and watching |

**Key Components**:

- `Runner` (`internal/application/orchestrator/runner.go:30-35`): The main orchestration engine that executes the agent pipeline
  - Manages session lifecycle and state tracking
  - Supports resume from snapshot
  - Emits events for streaming responses

- `sessionTracker` (`internal/application/orchestrator/session.go:51-54`): Tracks session state through pipeline stages

- Agent implementations in `internal/application/agents/`:
  - `retriever/`: Retrieves reference examples from knowledge base
  - `planner/`: Plans visualization structure
  - `stylist/`: Applies style guides
  - `visualizer/`: Generates visualization code (Mermaid/Matplotlib)
  - `critic/`: Evaluates output quality

- `WorkspaceService` (`internal/application/persistence/workspace_service.go:18-25`): Application service for workspace operations with transaction management

### Infrastructure Layer (`internal/infrastructure/`)

Contains external service integrations, persistence implementations, and adapters.

**Submodules**:

| Module | Purpose |
|--------|---------|
| `persistence/sqlite/` | SQLite repositories, GORM models, bootstrap, transaction management |
| `llm/` | LLM client implementations (OpenAI, Gemini, Anthropic, OpenRouter) |
| `cache/redis/` | Redis cache implementation |
| `crypto/aesgcm/` | AES-GCM encryption service |
| `crypto/keyderivation/` | Argon2id key derivation |
| `assets/localstore/` | Local filesystem asset storage |
| `agentstate/` | Snapshot store for session persistence |
| `nodes/httpnode/` | HTTP node adapter for external visualization services |
| `resilience/` | Circuit breaker and retry patterns |

**Key Patterns**:

- Factory pattern for LLM clients (`internal/infrastructure/llm/factory.go:35-46`)
- Repository pattern for persistence
- Adapter pattern for external services

### API Layer (`internal/api/`)

HTTP delivery mechanism with handlers, DTOs, and middleware.

**Structure**:

| Component | Purpose |
|-----------|---------|
| `handlers/` | HTTP request handlers for generate, workspace, history, assets, provider, batch |
| `dto/` | Data transfer objects for API request/response |
| `middleware/` | Request logging middleware |
| `router.go` | Route registration and router setup |

**Router Evolution** (`internal/api/router.go`):

- `SetupRouter()`: Phase 1-2 compatibility, generate endpoints only
- `SetupRouterWithPersistence()`: Phase 3, adds workspace/history/asset endpoints
- `SetupRouterWithConfigAndBatch()`: Full router with provider management, batch, and refine endpoints

---

## Module Dependency Graph

```
cmd/server
    |
    v
internal/api (API Layer)
    |
    +---> internal/application (Application Layer)
    |         |
    |         +---> internal/domain (Domain Layer)
    |         |
    |         +---> internal/infrastructure (Infrastructure Layer)
    |                   |
    |                   +---> internal/domain (Domain interfaces)
    |
    +---> internal/config (Configuration)
              |
              +---> internal/domain (Domain config types)
```

**Dependency Rules**:

1. Domain layer has NO dependencies on other layers
2. Application layer depends only on domain interfaces
3. Infrastructure layer depends on domain interfaces (implements them)
4. API layer depends on application services

---

## Data Flow

### Request Flow (Generate Pipeline)

```
HTTP Request
    |
    v
handlers.Generate() / handlers.StreamGenerate()
    |
    v
orchestrator.Runner.Start()
    |
    v
Pipeline Execution (stages in order):
    1. Retriever Agent --> Fetch reference examples
    2. Planner Agent --> Create visualization plan
    3. Stylist Agent --> Apply style guide
    4. Visualizer Agent --> Generate code/execute
    5. Critic Agent --> Evaluate quality
    |
    v
Event Streaming (SSE for streaming mode)
    |
    v
Session/Version Persistence
    |
    v
HTTP Response
```

### Agent Pipeline State Machine

```
SessionState
    |
    +-- StageStates[] (one per agent)
    |       |
    |       +-- Stage: StageName (retriever, planner, etc.)
    |       +-- Status: RunStatus (pending, running, completed, failed, canceled)
    |       +-- Input: AgentInput
    |       +-- Output: AgentOutput
    |       +-- Timing: StartedAt, CompletedAt, Duration
    |
    +-- FinalOutput (merged outputs from all stages)
```

The pipeline order is defined in `internal/domain/agent/types.go:20-26`:

```go
var pipelineOrder = []StageName{
    StageRetriever,
    StagePlanner,
    StageStylist,
    StageVisualizer,
    StageCritic,
}
```

### Workspace Data Model

```
Project (1) ----< (N) Folder
    |                    |
    +----< (N) Visualization
                          |
                          +----< (N) VisualizationVersion (immutable)
                          |           |
                          |           +----< (N) VersionArtifact
                          |
                          +----< (N) Session
                          |
                          +----< (N) Asset
```

---

## Key Abstractions

### Agent State Management

- **AgentInput/AgentOutput**: Immutable data structures passed between pipeline stages
- **SessionState**: Complete snapshot of pipeline execution state
- **Deep cloning**: All state mutations use defensive cloning (`cloneAgentInput`, `cloneAgentOutput`, etc. in `session.go`)

### Transaction Management

- `TxManager` interface (`internal/application/persistence/tx.go`): Unit of work pattern
- `RunInTx()`: Read-write transactions
- `ReadOnlyTx()`: Read-only transactions

### Snapshot/Resume

- `SnapshotStore` interface: Persist and restore session state at stage boundaries
- Enables resuming failed pipelines from last successful stage

### Event System

- `Event` type (`internal/domain/agent/events.go`): Typed events for run lifecycle
- Event types: `EventRunStarted`, `EventStageStarted`, `EventStageCompleted`, `EventStageFailed`, `EventRunCompleted`, `EventRunFailed`, `EventRunCanceled`
- Used for SSE streaming responses

---

## Entry Points

### HTTP Endpoints

| Endpoint | Handler | Purpose |
|----------|---------|---------|
| `POST /api/v1/generate` | `generate.go` | Synchronous generation |
| `POST /api/v1/generate/stream` | `generate.go` | Streaming generation (SSE) |
| `POST /api/v1/generate/batch` | `batch.go` | Batch generation |
| `POST /api/v1/refine` | `refine.go` | Image refinement |
| `/api/v1/projects/*` | `workspace.go` | Project CRUD |
| `/api/v1/folders/*` | `workspace.go` | Folder management |
| `/api/v1/visualizations/*` | `workspace.go` | Visualization management |
| `/api/v1/history/*` | `history.go` | Version history |
| `/api/v1/assets/*` | `assets.go` | Asset retrieval |
| `/api/v1/providers/*` | `provider.go` | Provider management |

### Application Bootstrap

Entry point: `cmd/server/main_test.go` (main package tests suggest a main.go should exist)

Configuration flow:
1. `config.Load()` loads from YAML + environment variables
2. SQLite bootstrap via `persistence/sqlite/bootstrap.go`
3. LLM client factory initialization
4. Agent construction with dependencies
5. Runner construction with agent registry
6. Router setup with all services

---

## Configuration

### Config Structure (`internal/config/config.go`)

```go
type Config struct {
    Server      ServerConfig      // Host, Port
    LLM         LLMConfig         // Providers map, Default provider
    Cache       CacheConfig       // Redis settings
    Output      OutputConfig      // DPI, Formats
    Persistence PersistenceConfig // Database path, foreign keys, WAL
    Assets      AssetsConfig      // Storage root, max file size
}
```

### Environment Variables

- `PAPERBANANA_CONFIG_FILE`: Explicit config file path
- `PAPERBANANA_NODE_CONFIG_FILE`: Custom node configuration
- Provider API keys: `GEMINI_API_KEY`, `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `OPENROUTER_API_KEY`

---

## Key Patterns Used

| Pattern | Location | Purpose |
|---------|----------|---------|
| Clean Architecture | Entire codebase | Layer isolation, dependency inversion |
| Repository Pattern | `domain/workspace/repositories.go`, `infrastructure/persistence/sqlite/` | Data access abstraction |
| Factory Pattern | `infrastructure/llm/factory.go` | LLM client creation |
| Unit of Work | `application/persistence/tx.go` | Transaction management |
| Strategy Pattern | Multiple LLM clients | Provider-specific implementations |
| Observer Pattern | Event system in orchestrator | State change notifications |
| State Pattern | Agent state machine | Pipeline stage management |
| Adapter Pattern | `infrastructure/nodes/httpnode/` | External service integration |

---

## Recommendations

1. **Entry Point Clarity**: The main entry point (`cmd/server/main.go`) appears missing from the scan. Verify existence or create proper entry point.

2. **Test Coverage**: Strong test coverage exists in application and infrastructure layers. Consider adding domain layer tests.

3. **Error Handling**: Error handling is consistent with typed `ErrorDetail`. Consider adding error codes for client-side handling.

4. **Configuration**: Provider configuration is split between config file and database (providers table). Consider consolidating.

5. **API Versioning**: Currently at v1. Plan for version negotiation strategy.

6. **Observability**: Consider adding metrics collection (prometheus) and distributed tracing.

7. **Caching**: Redis cache exists but appears optional. Consider caching strategy for frequently accessed visualizations.
