# Conventions

> Go coding conventions and project-specific patterns for PaperBanana.

---

## Naming

### Files and Packages

- File names: `lowercase_snake_case.go` (e.g., `workspace_repository.go`, `snapshot_store.go`)
- Package names: `lowercase` single word (e.g., `sqlite`, `aesgcm`, `openai`)
- Test files: `*_test.go` alongside the source file

### Functions

- Constructor functions: `NewXxx` (e.g., `NewHandler`, `NewRunner`, `NewProjectRepository`)
- Factory functions: `NewXxxWithOptions` when additional configuration is needed
- Unexported helpers: `lowercaseCamelCase` (e.g., `buildAgentInput`, `cloneStringMap`)

### Variables

- Local variables: `camelCase` (e.g., `sessionID`, `providerConfig`)
- Package-level: `camelCase` for unexported, `PascalCase` for exported
- Sentinel errors: `ErrXxx` (e.g., `ErrResumeRequiresSession`, `ErrNotFound`)
- Constants: `PascalCase` for exported, `camelCase` for unexported

### Structs

- `PascalCase` for all structs (e.g., `GenerateRequest`, `RunResult`, `AgentState`)
- GORM models: `XxxModel` suffix (e.g., `ProjectModel`, `SessionModel`)
- Config structs: `XxxConfig` suffix (e.g., `BootstrapConfig`, `ServerConfig`)

---

## Code Style

- **Formatter**: `gofmt` (standard Go formatting)
- **Indentation**: Tabs
- **Imports**: Grouped: stdlib, then external, then internal
- **Line length**: No hard limit, but prefer readability

### Import Grouping

```go
import (
    // stdlib
    "context"
    "fmt"
    "time"

    // external
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "gorm.io/gorm"

    // internal
    "github.com/paperbanana/paperbanana/internal/domain/agent"
    domainagent "github.com/paperbanana/paperbanana/internal/domain/agent"
)
```

---

## Import Aliases

Aliases are used when package names clash or when the default name is ambiguous. The project has established these aliases:

| Alias | Package | Reason |
|-------|---------|--------|
| `domainagent` | `internal/domain/agent` | Avoids clash with `internal/agents/` and `internal/application/agents/` |
| `domainllm` | `internal/domain/llm` | Avoids clash with `internal/infrastructure/llm/` |
| `llminfra` | `internal/infrastructure/llm` | Avoids clash with `internal/domain/llm/` |
| `criticagent` | `internal/application/agents/critic` | Avoids clash with other agent packages |
| `planneragent` | `internal/application/agents/planner` | Avoids clash with other agent packages |
| `retrieveragent` | `internal/application/agents/retriever` | Avoids clash with other agent packages |
| `stylistagent` | `internal/application/agents/stylist` | Avoids clash with other agent packages |
| `visualizeragent` | `internal/application/agents/visualizer` | Avoids clash with other agent packages |
| `configservice` | `internal/application/config` | Avoids clash with `internal/config/` |
| `rediscache` | `internal/infrastructure/cache/redis` | Avoids clash with `github.com/redis/go-redis/v9` |
| `agentstate` | `internal/infrastructure/agentstate` | Avoids ambiguity with `domain/agent` |
| `domainworkspace` | `internal/domain/workspace` | Used in handlers to avoid clash with handler-level types |
| `pbconfig` | `internal/config` | Used in infrastructure/llm to avoid clash with domain/config |
| `openaiclient` | `internal/infrastructure/llm/openai` | Avoids clash with sashabaranov/openai |

---

## Error Handling

See [Error Handling](./error-handling.md) for full details. Summary:

- **Wrapping**: `fmt.Errorf("context: %w", err)` -- always add context
- **Sentinel errors**: `var ErrXxx = errors.New("package: description")` for cross-package comparison
- **Cleanup**: `errors.Join(err, cleanupErr)` for combining multiple failures
- **Classification**: `domainagent.ClassifyError(err)` for pipeline event errors
- **Anti-pattern**: Never use `errors.Is(err, errors.New("..."))` -- always false

---

## Logging

See [Logging Guidelines](./logging-guidelines.md) for full details. Summary:

- **Framework**: Zap (`go.uber.org/zap`)
- **Structured fields**: Always use `zap.String()`, `zap.Error()`, etc.
- **Never interpolate**: No `fmt.Sprintf` in log messages
- **Pass via DI**: `*zap.Logger` is a constructor parameter, not a global

---

## Comments

### Rationale Comments

Comments should explain *why*, not *what*. Use them for non-obvious decisions:

```go
// Write timeout is disabled because SSE and synchronous generation
// can legitimately exceed several minutes.
WriteTimeout: 0,
```

### Lifecycle Markers

Use comments to mark lifecycle-sensitive code:

```go
// GD-UI-004: Resume event for snapshot restoration
if tracker.state.Restore.RestoredFrom != "" {
    // ...
}

// SECURITY: disabled by default -- plot mode executes Python code
PlotEnabled: false,
```

### Section Markers

Large files use blank-line-separated sections with brief comments:

```go
// --- Constructor ---

func NewRunner(agents map[...]BaseAgent, opts ...RunnerOption) *Runner { ... }

// --- Public API ---

func (r *Runner) Start(...) (*RunHandle, error) { ... }
```

### Go Doc on Exported Types

All exported types and functions have Go doc comments:

```go
// BootstrapResult holds the result of a successful bootstrap.
type BootstrapResult struct {
    DB *gorm.DB
}
```

---

## Function Design

### Large Orchestration Functions

Large orchestration functions are accepted at composition points (e.g., `Runner.execute`, `main.go`). These functions coordinate multiple steps and are inherently procedural. Do not refactor them into tiny methods if the flow is clearer when linear.

### Constructors

Constructors accept their primary dependency + a Config struct:

```go
func NewAgent(client domainllm.LLMClient, cfg Config) *Agent { ... }
func NewHandler(runner Runner, logger *zap.Logger) *Handler { ... }
```

### Option Pattern

For extensible configuration, use the functional options pattern:

```go
type RunnerOption func(*Runner)

func WithSnapshotStore(store SnapshotStore) RunnerOption { ... }
func WithStageTimeouts(timeouts StageTimeouts) RunnerOption { ... }
func WithGracefulDegradation(enabled bool) RunnerOption { ... }
```

---

## Module Design

### Interfaces Close to Consumers

Repository interfaces are defined in the domain package (close to the consumer), not in the infrastructure package (close to the implementation):

```go
// internal/domain/workspace/repositories.go -- Interface definition (consumer)
type ProjectRepository interface { ... }

// internal/infrastructure/persistence/sqlite/workspace_repository.go -- Implementation
type ProjectRepository struct { db *gorm.DB }
```

### Small Consumer-Owned Interfaces

The `BaseAgent` interface in `internal/domain/agent/` is intentionally small (5 methods). Each agent implements only what the pipeline needs. Similarly, `LLMClient` in `internal/domain/llm/` defines only the methods that consumers call.

---

## Backend Patterns

### HTTP DTOs in Handlers

Request and response structs are defined in the handler or DTO package, not in the domain. The domain layer is unaware of HTTP concerns:

```go
// internal/api/handlers/generate.go -- HTTP-specific DTO
type GenerateRequest struct {
    Prompt string `json:"prompt"`
    // ...
}

// internal/domain/agent/types.go -- Domain type (no JSON tags)
type AgentInput struct {
    SessionID string
    Content   string
    // ...
}
```

### Orchestration in Runner

The `Runner` is the single orchestration point. It does not delegate orchestration logic to agents; agents are stage executors only.

### Provider Behavior in Infrastructure Packages

Each LLM provider has its own package under `internal/infrastructure/llm/`. The factory (`factory.go`) selects the appropriate implementation based on the provider name.

---

## Risky Inconsistencies

### `errors.Is(err, errors.New(...))` Anti-Pattern

Present in `handlers/history.go`, `handlers/workspace.go`, and `handlers/assets.go`. These comparisons always return `false`. See [Error Handling](./error-handling.md) for details.

### Mixed Error Handling in Handlers

Some handlers log and return errors; others only return. This leads to inconsistent error logging and potential duplicate entries.

### Duplicate Asset Stores

Two implementations with different feature sets. See [Database Guidelines](./database-guidelines.md) for details.
