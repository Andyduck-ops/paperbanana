# PaperBanana Domain Model

> **Pre-Tauri Note**: This document describes the current web-app architecture. The bounded contexts and entity relationships remain valid, but the transport layer (REST + SSE) will be replaced by Tauri IPC in the migration.

## Overview

PaperBanana is an AI-powered academic paper figure generation tool. The domain is organized into five bounded contexts, each owning a distinct area of responsibility.

```
┌─────────────────────────────────────────────────────────────────┐
│                        PaperBanana                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Agent     │  │  Workspace  │  │       Config            │ │
│  │   Domain    │  │   Domain    │  │       Domain            │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐                              │
│  │    LLM      │  │   Crypto    │                              │
│  │   Domain    │  │   Domain    │                              │
│  └─────────────┘  └─────────────┘                              │
└─────────────────────────────────────────────────────────────────┘
```

## Bounded Contexts

### 1. Agent Domain (`internal/domain/agent/`)

The core orchestration domain. Manages the multi-agent pipeline that produces figures from prompts.

**Key types** (from `internal/domain/agent/types.go`):

- **`StageName`** -- Enum for pipeline stages: `retriever`, `planner`, `stylist`, `visualizer`, `critic`, `polish`
- **`RunStatus`** -- Enum for run states: `pending`, `running`, `completed`, `failed`, `canceled`
- **`VisualMode`** -- Enum for output types: `diagram`, `plot`
- **`ArtifactKind`** -- Enum for artifact types: `reference_bundle`, `plan`, `rendered_figure`, `prompt_trace`, `critique`, `polished_image`
- **`PipelineMode`** -- Constants: `full`, `planner-critic`, `vanilla`

**Core entities:**

```go
// Artifact -- a named, typed output produced by a pipeline stage
type Artifact struct {
    ID        string
    Kind      ArtifactKind
    MIMEType  string
    URI       string
    Content   string            // Text content
    Bytes     []byte            // Legacy binary data (deprecated)
    Shared    *SharedBytes      // Reference-counted binary data
    AssetID   string            // Links to workspace Asset
    ProjectID string            // For URL construction
    Metadata  map[string]string
}
```

```go
// SessionState -- the full state of a pipeline run
type SessionState struct {
    SchemaVersion string
    SessionID     string
    RequestID     string
    Status        RunStatus
    CurrentStage  StageName
    FailedStage   StageName
    Pipeline      []StageName
    InitialInput  AgentInput
    StageStates   []AgentState
    FinalOutput   AgentOutput
    Error         *ErrorDetail
    Restore       RestoreMetadata
    Metadata      map[string]string
    StartedAt     time.Time
    UpdatedAt     time.Time
    CompletedAt   time.Time
}
```

```go
// BatchResult -- aggregated results from a batch execution
type BatchResult struct {
    BatchID    string
    Results    []CandidateResult
    Successful int
    Failed     int
    Timing     BatchTiming
}
```

**Events** (from `internal/domain/agent/events.go`):

| Event Type | Description |
|------------|-------------|
| `run_started` | Pipeline execution begins |
| `stage_started` | A stage begins processing |
| `stage_completed` | A stage completes successfully |
| `stage_failed` | A stage fails |
| `run_completed` | Pipeline execution finishes |
| `run_failed` | Pipeline execution fails |
| `run_canceled` | Pipeline execution is canceled |
| `resume_start` | Resuming from a snapshot |
| `batch_start` | Batch execution begins |
| `candidate_start` | A candidate within a batch begins |
| `candidate_complete` | A candidate within a batch completes |
| `batch_complete` | Batch execution finishes |

**Error classification** (from `internal/domain/agent/errors.go`):

- Error codes: `llm_timeout`, `rate_limit`, `service_unavailable`, `network_error`, `invalid_input`, `invalid_config`, `unsupported_type`, `resource_not_found`, `missing_api_key`, `invalid_model`, `execution_failed`, `stage_timeout`, `cancelled`, `internal_error`, `unknown`
- Error categories: `transient`, `permanent`, `configuration`, `internal`

### 2. Workspace Domain (`internal/domain/workspace/`)

Manages the organizational structure for user work: projects, folders, visualizations, versions, and assets.

**Core entities** (from `internal/domain/workspace/entities.go`):

```go
// Project -- top-level organizational unit
type Project struct {
    ID          string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Folder -- nested organization within a project
type Folder struct {
    ID        string
    ProjectID string
    ParentID  *string     // nil for root folders
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time  // Soft delete for trash
}

// Visualization -- a single chart or diagram entry
type Visualization struct {
    ID               string
    ProjectID        string
    FolderID         *string    // nil for root level
    Name             string
    CurrentVersionID *string
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        *time.Time // Soft delete
}

// VisualizationVersion -- immutable snapshot of a visualization's state
type VisualizationVersion struct {
    ID              string
    VisualizationID string
    ProjectID       string
    VersionNumber   int
    SessionID       string
    Summary         string
    Artifacts       []VersionArtifact
    CreatedAt       time.Time
    SessionSnapshot *agent.SessionState  // Full session for restore
}

// Asset -- metadata for a file belonging to a visualization
type Asset struct {
    ID              string
    ProjectID       string
    VisualizationID string
    VersionID       *string
    MIMEType        string
    StorageKey      string     // External store key
    ByteSize        int64
    ChecksumSHA256  string
    CreatedAt       time.Time
    DeletedAt       *time.Time
}

// SessionRecord -- persisted session state for restore and audit
type SessionRecord struct {
    ID              string
    ProjectID       string
    VisualizationID *string
    Status          string
    CurrentStage    string
    SchemaVersion   string
    Snapshot        *agent.SessionState
    CreatedAt       time.Time
    UpdatedAt       time.Time
    CompletedAt     *time.Time
}
```

**Repository interfaces** (from `internal/domain/workspace/repositories.go`):

| Interface | Responsibility |
|-----------|---------------|
| `ProjectRepository` | CRUD for projects; soft delete |
| `FolderRepository` | CRUD for folder hierarchies; recursive CTE for descendants |
| `VisualizationRepository` | CRUD for visualizations; version management |
| `VersionRepository` | Append-only version storage |
| `SessionRepository` | Persist session state for restore and audit |
| `AssetRepository` | Track asset metadata; actual bytes stored externally |

All repository queries are project-scoped (enforced by `ProjectScopedQuery` interface).

### 3. Config Domain (`internal/domain/config/`)

Manages LLM provider configuration, including built-in presets and encrypted API keys.

**Core entities** (from `internal/domain/config/provider.go`):

```go
// ProviderType -- enum for supported LLM providers
type ProviderType string  // "openai", "anthropic", "gemini", "deepseek", etc.

// ModelInfo -- a model configuration within a provider
type ModelInfo struct {
    ID             string
    Name           string
    MaxTokens      int
    SupportsVision bool
    Enabled        bool
}

// SystemProviderPreset -- predefined provider configuration
type SystemProviderPreset struct {
    Type           ProviderType
    Name           string
    DisplayName    string
    APIHost        string
    DocsURL        string
    APIKeyURL      string
    DefaultModels  []ModelInfo
    SupportsVision bool
}

// Provider -- a configured LLM provider instance
type Provider struct {
    ID          string
    Type        ProviderType
    Name        string
    DisplayName string
    APIHost     string
    APIKey      string        // Never exposed in JSON
    Models      []ModelInfo
    Enabled     bool
    IsSystem    bool
    IsDefault   bool
    TimeoutMs   int
    QueryModel  string        // Model for retrieval/planning/critique
    GenModel    string        // Model for visualization generation
}
```

**API key management** (from `internal/domain/config/apikey.go`):

```go
// APIKey -- encrypted API key for a provider
type APIKey struct {
    ID           string
    ProviderID   string
    EncryptedKey string    // Never exposed in JSON
    KeyPrefix    string
    KeySuffix    string
    IsActive     bool
    LastUsedAt   *time.Time
}
```

**Config events** (from `internal/domain/config/events.go`):

| Event | Description |
|-------|-------------|
| `provider_created` | New provider added |
| `provider_updated` | Provider configuration changed |
| `provider_deleted` | Provider removed |
| `key_added` | API key added |
| `key_deleted` | API key removed |
| `key_toggled` | API key active status changed |

**Model selection** (from `internal/domain/config/model_defaults.go`):

- `SupportsImageGeneration(modelID)` -- identifies image-generation models
- `SelectPreferredQueryModel(configured, models)` -- selects the best model for planning/retrieval/critique
- `SelectPreferredGenerationModel(providerType, name, configured, models)` -- selects the best model for image generation

### 4. LLM Domain (`internal/domain/llm/`)

Abstracts LLM interaction across all providers.

**Core interface** (from `internal/domain/llm/client.go`):

```go
// LLMClient -- the contract shared by all provider implementations
type LLMClient interface {
    Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
    GenerateStream(ctx context.Context, req GenerateRequest) (<-chan StreamChunk, <-chan error)
    Provider() string
}

// ImageGenerator -- optional capability for providers that return generated images
type ImageGenerator interface {
    GenerateImage(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
}

// ModelLister -- optional capability for providers that list available models
type ModelLister interface {
    ListModels(ctx context.Context) ([]ModelInfo, error)
}
```

**Message types:**

```go
type Message struct {
    Role  Role    // "user" or "assistant"
    Parts []Part  // Text or image parts
}

type Part struct {
    Type     PartType  // "text" or "image"
    Text     string
    MIMEType string
    Data     []byte    // Inline image data
    URL      string    // URL-based image
}
```

### 5. Crypto Domain (`internal/domain/crypto/`)

Provides encryption and key derivation abstractions for secure API key storage.

**Core interfaces** (from `internal/domain/crypto/encryption.go`):

```go
// EncryptionService -- secure encryption/decryption for sensitive data
type EncryptionService interface {
    Encrypt(ctx context.Context, plaintext string) (string, error)
    Decrypt(ctx context.Context, ciphertext string) (string, error)
    Mask(plaintext string) string  // "sk-abc123xyz" -> "sk-abc****xyz"
}

// KeyDerivationService -- derives encryption keys from passwords/secrets
type KeyDerivationService interface {
    DeriveKey(password string, salt []byte) []byte  // Argon2id
    GenerateSalt() ([]byte, error)
}
```

## Key Interfaces

### BaseAgent (`internal/domain/agent/agent.go`)

The lifecycle interface every pipeline stage must implement:

```go
type BaseAgent interface {
    Initialize(ctx context.Context) error
    Execute(ctx context.Context, input AgentInput) (AgentOutput, error)
    Cleanup(ctx context.Context) error
    GetState() AgentState
    RestoreState(state AgentState) error
}
```

### Runner (`internal/application/orchestrator/runner.go`)

Executes the canonical pipeline and streams stage events:

```go
// Create a runner with all five agents
func NewCanonicalRunner(retriever, planner, stylist, visualizer, critic BaseAgent, opts ...RunnerOption) *Runner

// Start a new pipeline run
func (r *Runner) Start(ctx context.Context, input AgentInput) (*RunHandle, error)

// Resume a previously failed/canceled run from a snapshot
func (r *Runner) Resume(ctx context.Context, input AgentInput) (*RunHandle, error)
```

Runner options:
- `WithEventBuffer(size)` -- configure event channel buffer
- `WithSnapshotStore(store)` -- enable snapshot persistence for resume
- `WithStageTimeouts(timeouts)` -- per-stage timeout configuration
- `WithGracefulDegradation(enabled)` -- continue pipeline when non-critical stages fail

### BatchRunner (`internal/application/orchestrator/batch_runner.go`)

Executes multiple candidates in parallel with a shared retriever:

```go
// Create a batch runner with an agent factory
func NewBatchRunner(factory AgentFactory, opts ...BatchOption) *BatchRunner

// Start batch execution
func (r *BatchRunner) StartBatch(ctx context.Context, inputs []AgentInput) (*BatchHandle, error)

// Retrieve batch results
func (r *BatchRunner) GetBatchResult(batchID string) (*BatchResult, error)

// Check batch progress
func (r *BatchRunner) GetBatchProgress(batchID string) (*BatchProgress, error)
```

### RuntimeClient (`internal/domain/llm/client.go`)

Resolves provider/model at request time:

```go
func ResolveModel(requestModel, defaultModel string) string
```

## Data Flow

### Single Generate (5-Stage Pipeline)

The canonical pipeline runs stages in order: `retriever` -> `planner` -> `stylist` -> `visualizer` -> `critic`.

```
User Prompt
    │
    ▼
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌────────────┐    ┌──────────┐
│ Retriever│───▶│ Planner  │───▶│ Stylist  │───▶│ Visualizer │───▶│  Critic  │
└──────────┘    └──────────┘    └──────────┘    └────────────┘    └──────────┘
     │               │               │                │                │
     ▼               ▼               ▼                ▼                ▼
  References        Plan        Styled plan      Rendered figure   Critique
```

Pipeline modes (controlled by `metadata["config.pipeline_mode"]`):
- **`full`** -- all five stages (default)
- **`planner-critic`** -- only planner and critic
- **`vanilla`** -- only visualizer

Graceful degradation: when enabled, non-critical stages (`retriever`, `stylist`, `critic`) can fail without stopping the pipeline. Critical stages (`planner`, `visualizer`) always cause pipeline failure.

### Batch (Shared Retriever + Parallel Candidates)

```
               ┌──────────┐
               │ Retriever │  (runs once, shared)
               └─────┬─────┘
                     │
         ┌───────────┼───────────┐
         ▼           ▼           ▼
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │ Cand. 0 │ │ Cand. 1 │ │ Cand. N │  (parallel)
    │ Planner │ │ Planner │ │ Planner │
    │ Stylist │ │ Stylist │ │ Stylist │
    │Visualizr│ │Visualizr│ │Visualizr│
    │ Critic  │ │ Critic  │ │ Critic  │
    └─────────┘ └─────────┘ └─────────┘
```

The `BatchRunner` uses an `AgentFactory` to create fresh agent instances per candidate.

### Refine (Separate Polish Path)

The refine flow uses the `polish` stage to iterate on an existing visualization based on user feedback. The pipeline is shorter: it takes an existing artifact and feedback, then runs a polish agent to produce an improved version.

## Unified Language (Ubiquitous Language)

| Term | Definition |
|------|-----------|
| **Generation** | The process of creating a new figure using AI |
| **Refinement** | Iterative improvement of a figure based on feedback |
| **Asset** | A generated figure file (image, code, or data) |
| **Version** | An immutable historical snapshot of an asset |
| **Project** | A user's top-level organizational unit |
| **Channel** | A specific role + model configuration mapping |
| **Pipeline** | The multi-agent execution flow that produces a figure |
| **Retrieval** | Fetching relevant information from reference data |
| **Artifact** | A named, typed output produced by a pipeline stage |
| **Session** | The full state of a single pipeline run, persistable for restore |
| **Candidate** | One of multiple parallel generation attempts in a batch |
| **Provider** | An LLM service configuration (OpenAI, Gemini, Anthropic, etc.) |
| **Critique** | The critic agent's evaluation of a generated figure |

## Integration Points

1. **Frontend <-> Backend**: REST API + Server-Sent Events (SSE)
2. **Backend -> LLM**: Provider-specific SDKs via the `LLMClient` interface
3. **Backend <-> Database**: GORM + SQLite
4. **Inter-domain**: Domain events via channels (`ConfigWatcher.Subscribe()`, `Event` streaming)
