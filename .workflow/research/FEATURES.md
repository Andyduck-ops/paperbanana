# Codebase Features Analysis

## Overview

PaperBanana is an **image refinement workspace** for scientific visualization generation. The application implements a multi-agent pipeline architecture where specialized agents (Retriever, Planner, Stylist, Visualizer, Critic) collaborate to generate scientific diagrams and plots from textual descriptions. The system supports both single and batch generation modes, with streaming SSE events for real-time progress feedback.

---

## Core Features

### 1. Multi-Agent Pipeline Generation

**Description**: Orchestrated execution of 5 specialized agents in sequence to transform text descriptions into scientific visualizations.

**Agents in Pipeline**:
| Agent | Role | Output |
|-------|------|--------|
| Retriever | Retrieves reference examples from PaperBananaBench | `RetrievedReference[]` |
| Planner | Generates detailed visualization plan | `Plan` (text content) |
| Stylist | Applies style guidelines (optional) | Enhanced plan |
| Visualizer | Renders the final image | `RenderedFigure` artifact |
| Critic | Reviews and iterates on output | Critique feedback |

**File References**:
- `internal/application/orchestrator/runner.go:37-48` - Canonical pipeline construction
- `internal/application/orchestrator/runner.go:130-197` - Pipeline execution logic
- `internal/domain/agent/agent.go:6-12` - BaseAgent interface

**Modes**:
- `diagram` - Scientific diagram illustration via LLM image generation
- `plot` - Statistical plots via Python matplotlib code generation

---

### 2. Retrieval System

**Description**: Reference example retrieval from PaperBananaBench dataset to guide generation quality.

**Retrieval Modes**:
| Mode | Behavior |
|------|----------|
| `auto` | LLM-based selection of top relevant examples |
| `manual` | Pre-curated high-quality examples |
| `random` | Random selection from candidate pool |
| `none` | Skip retrieval entirely |

**File References**:
- `internal/application/agents/retriever/agent.go:26-33` - RetrievalMode constants
- `internal/application/agents/retriever/agent.go:189-246` - Mode execution logic
- `internal/application/agents/retriever/agent.go:351-369` - ParseTopReferences for LLM response parsing

---

### 3. Visualizer Execution Paths

**Description**: Multiple execution strategies for visualization rendering.

**Execution Paths**:
1. **LLM Image Path** (diagram mode) - Direct image generation via vision-capable LLM
2. **LLM Plot Path** (plot mode) - Code generation + Python matplotlib execution
3. **Node Runner Path** - Custom HTTP node execution for specialized renderers
4. **Reuse Path** - Skip re-rendering when Critic requests no changes

**File References**:
- `internal/application/agents/visualizer/agent.go:62-98` - Execute dispatch logic
- `internal/application/agents/visualizer/agent.go:114-151` - Diagram execution
- `internal/application/agents/visualizer/agent.go:153-204` - Plot execution
- `internal/application/agents/visualizer/node_runner.go` - Custom node execution

---

### 4. Critique and Iteration Loop

**Description**: Automated quality review with revision feedback for iterative refinement.

**Critique Capabilities**:
- Content fidelity verification against methodology/caption
- Text QA for typos and labels
- Presentation clarity assessment
- Revision description generation

**File References**:
- `internal/application/agents/critic/prompt.go:102-141` - Diagram system prompt
- `internal/application/agents/critic/prompt.go:143-185` - Plot system prompt
- `internal/domain/agent/events.go:14` - Critique round events

**Iteration Control**:
- Configurable critic rounds via `critic_rounds` parameter
- "No changes needed" detection for early termination
- `internal/application/agents/visualizer/agent.go:319-346` - Reuse detection

---

### 5. Batch Generation

**Description**: Parallel multi-candidate generation with shared retriever results.

**Batch Architecture**:
- Single shared Retriever execution
- Parallel Planner-Visualizer-Critic pipelines
- Concurrent candidate limit: 10 (configurable)
- Results stored for retrieval by batch ID

**File References**:
- `internal/application/orchestrator/batch_runner.go:24-48` - BatchRunner struct
- `internal/application/orchestrator/batch_runner.go:117-132` - StartBatch
- `internal/application/orchestrator/batch_runner.go:134-266` - executeBatch

---

### 6. Image Refinement

**Description**: Standalone image enhancement feature for uploaded images.

**File References**:
- `internal/api/router.go:177-180` - Refine endpoint registration
- `web/src/hooks/useRefine.ts` - Frontend refine hook
- `web/src/lib/refine.ts` - Refine API client

---

### 7. Streaming Events (SSE)

**Description**: Real-time progress updates via Server-Sent Events.

**Event Types**:
| Event | When |
|-------|------|
| `run_started` | Pipeline begins |
| `stage_started` | Agent begins execution |
| `stage_completed` | Agent finishes successfully |
| `stage_failed` | Agent encounters error |
| `run_completed` | Pipeline finishes |
| `run_failed` | Pipeline fails |
| `run_canceled` | User cancellation |

**Batch Events**:
- `batch_start`, `candidate_start`, `candidate_complete`, `batch_complete`

**File References**:
- `internal/domain/agent/events.go:5-21` - EventType constants
- `internal/application/orchestrator/session.go:34-42` - RunHandle Events channel
- `web/src/lib/sse.ts` - Frontend SSE client
- `web/src/hooks/useGenerate.ts:79-146` - SSE event handling

---

## API Endpoints

### Generation Endpoints

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | `/api/v1/generate` | `handlers.Generate` | Synchronous generation |
| POST | `/api/v1/generate/stream` | `handlers.StreamGenerate` | SSE streaming generation |
| POST | `/api/v1/generate/batch` | `handlers.StreamBatchGenerate` | Batch generation with SSE |
| POST | `/api/v1/batch/download` | `handlers.DownloadBatchZip` | Download batch results |
| POST | `/api/v1/refine` | `handlers.Refine` | Image refinement |

**File References**:
- `internal/api/router.go:52-74` - Basic router (generate only)
- `internal/api/router.go:78-126` - Full router with persistence
- `internal/api/router.go:134-183` - Complete router with config/batch/refine

### Workspace/Project Endpoints

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | `/api/v1/projects` | `workspaceHandler.CreateProject` | Create project |
| GET | `/api/v1/projects` | `workspaceHandler.ListProjects` | List all projects |
| GET | `/api/v1/projects/:project_id` | `workspaceHandler.GetProject` | Get project details |
| POST | `/api/v1/folders` | `workspaceHandler.CreateFolder` | Create folder |
| POST | `/api/v1/visualizations` | `workspaceHandler.CreateVisualization` | Create visualization |
| GET | `/api/v1/folders/contents` | `workspaceHandler.ListFolderContents` | List folder contents |
| POST | `/api/v1/workspace/move` | `workspaceHandler.MoveItem` | Move/reparent item |
| POST | `/api/v1/workspace/trash` | `workspaceHandler.TrashItem` | Soft delete |
| POST | `/api/v1/workspace/restore` | `workspaceHandler.RestoreItem` | Restore from trash |

**File References**:
- `internal/api/handlers/workspace.go` - Workspace handlers
- `internal/domain/workspace/entities.go` - Project, Folder, Visualization entities

### History/Session Endpoints

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/v1/history` | `historyHandler.ListHistory` | List sessions |
| GET | `/api/v1/history/:project_id/:version_id` | `historyHandler.GetVersion` | Get version details |
| GET | `/api/v1/session/latest` | `historyHandler.GetLatestSession` | Get latest session |
| GET | `/api/v1/session/:session_id` | `historyHandler.GetSession` | Get session details |

**File References**:
- `internal/api/router.go:110-116` - History route registration
- `internal/application/persistence/history_service.go` - History service

### Asset Endpoints

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/v1/assets` | `assetHandler.ListAssets` | List assets |
| GET | `/api/v1/assets/:project_id/:asset_id` | `assetHandler.GetAsset` | Get asset metadata |
| GET | `/api/v1/assets/:project_id/:asset_id/download` | `assetHandler.DownloadAsset` | Download asset |
| GET | `/api/v1/assets/version/:project_id/:version_id` | `assetHandler.ListAssetsByVersion` | List version assets |

**File References**:
- `internal/api/handlers/assets.go` - Asset handlers
- `internal/application/persistence/asset_service.go` - Asset service

### Provider Configuration Endpoints

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/v1/providers/presets` | `providerHandler.ListPresets` | List provider presets |
| GET | `/api/v1/providers` | `providerHandler.ListProviders` | List configured providers |
| GET | `/api/v1/providers/:id` | `providerHandler.GetProvider` | Get provider details |
| POST | `/api/v1/providers` | `providerHandler.CreateProvider` | Create provider |
| PUT | `/api/v1/providers/:id` | `providerHandler.UpdateProvider` | Update provider |
| DELETE | `/api/v1/providers/:id` | `providerHandler.DeleteProvider` | Delete provider |
| POST | `/api/v1/providers/:id/default` | `providerHandler.SetDefaultProvider` | Set default |
| GET | `/api/v1/providers/:id/keys` | `providerHandler.ListAPIKeys` | List API keys |
| POST | `/api/v1/providers/:id/keys` | `providerHandler.AddAPIKey` | Add API key |
| DELETE | `/api/v1/providers/:id/keys/:keyId` | `providerHandler.DeleteAPIKey` | Delete key |
| PATCH | `/api/v1/providers/:id/keys/:keyId` | `providerHandler.ToggleAPIKey` | Toggle key |
| GET | `/api/v1/providers/:id/models` | `providerHandler.ListModels` | List models |
| POST | `/api/v1/providers/:id/test` | `providerHandler.TestExistingProvider` | Test provider |
| POST | `/api/v1/providers/test` | `providerHandler.TestProvider` | Test new provider |
| POST | `/api/v1/providers/reset` | `providerHandler.ResetSystemProviders` | Reset to defaults |
| GET | `/api/v1/config/stream` | `configSSEHandler.StreamConfigChanges` | SSE config updates |

**File References**:
- `internal/api/router.go:139-167` - Provider route registration
- `internal/domain/config/provider.go` - Provider domain entity
- `internal/domain/config/apikey.go` - API key entity

---

## Frontend Components

### Main Application Structure

**File**: `web/src/App.tsx:34-375`

**Pages**:
- `main` - Generation workspace (default)
- `settings` - Provider configuration
- `provider-new` - New provider form
- `provider-edit` - Edit provider form

**Main Tabs**:
- `generate` - Image generation panel
- `refine` - Image refinement panel

### Key Components

| Component | File | Purpose |
|-----------|------|---------|
| `GeneratePanel` | `web/src/components/GeneratePanel.tsx` | Prompt input + config options |
| `ProgressPanel` | `web/src/components/ProgressPanel.tsx` | Pipeline stage progress |
| `ResultPanel` | `web/src/components/ResultPanel.tsx` | Generated artifact display |
| `BatchProgressPanel` | `web/src/components/BatchProgressPanel.tsx` | Batch candidate progress |
| `RefinePanel` | `web/src/components/RefinePanel.tsx` | Image upload + refinement |
| `HistorySidebar` | `web/src/components/HistorySidebar.tsx` | Session history |
| `ExportModal` | `web/src/components/ExportModal.tsx` | Export format options |
| `ConfigPanel` | `web/src/components/ConfigPanel.tsx` | Generation configuration |
| `DualInputPanel` | `web/src/components/DualInputPanel.tsx` | Method + caption inputs |
| `SettingsPage` | `web/src/pages/SettingsPage.tsx` | Provider management UI |
| `ProviderEditPage` | `web/src/pages/ProviderEditPage.tsx` | Provider form |

### React Hooks

| Hook | File | Purpose |
|------|------|---------|
| `useGenerate` | `web/src/hooks/useGenerate.ts` | Single generation state + SSE |
| `useBatchGeneration` | `web/src/hooks/useBatchGeneration.ts` | Batch generation state |
| `useRefine` | `web/src/hooks/useRefine.ts` | Refinement state |
| `useProviders` | `web/src/hooks/useProviders.ts` | Provider CRUD |
| `useHistory` | `web/src/hooks/useHistory.ts` | Session history |
| `useToast` | `web/src/hooks/useToast.ts` | Toast notifications |
| `useLanguage` | `web/src/hooks/useLanguage.ts` | i18n |

---

## Domain Entities

### Workspace Domain

**File**: `internal/domain/workspace/entities.go`

| Entity | Description |
|--------|-------------|
| `Project` | Top-level organizational unit |
| `Folder` | Nested organization within project |
| `Visualization` | Single chart/diagram entry |
| `VisualizationVersion` | Immutable snapshot of visualization |
| `VersionArtifact` | Asset metadata attached to version |
| `SessionRecord` | Full session state for restore |
| `Asset` | File metadata with storage key |

### Agent Domain

**File**: `internal/domain/agent/agent.go`, events.go

| Entity | Description |
|--------|-------------|
| `AgentInput` | Input to agent Execute |
| `AgentOutput` | Output from agent Execute |
| `AgentState` | Agent state snapshot |
| `SessionState` | Full pipeline session state |
| `Event` | Pipeline event for SSE |
| `BatchEvent` | Batch execution event |
| `Artifact` | Generated artifact (image, code, etc.) |
| `RetrievedReference` | Reference example from bench |
| `CritiqueRound` | Critique iteration feedback |
| `VisualIntent` | Generation intent (mode, goal, style) |
| `PromptMetadata` | Prompt template info |

---

## Configuration

### Generation Configuration

**File**: `web/src/components/ConfigPanel.tsx`, `web/src/components/GeneratePanel.tsx`

| Option | Type | Description |
|--------|------|-------------|
| `aspect_ratio` | string | Output image aspect ratio |
| `critic_rounds` | int | Number of critique iterations |
| `retrieval_mode` | string | Reference retrieval strategy |
| `pipeline_mode` | string | Full vs. partial pipeline |
| `query_model` | string | Model for query/planning |
| `gen_model` | string | Model for visualization |
| `visualizer_node` | string | Custom visualizer node |
| `num_candidates` | int | Batch candidate count |

### Provider Configuration

**File**: `internal/domain/config/provider.go`, `internal/infrastructure/persistence/sqlite/provider_model.go`

| Field | Description |
|-------|-------------|
| `type` | Provider type (openai, gemini, anthropic, openrouter) |
| `name` | Display name |
| `base_url` | API endpoint override |
| `query_model` | Default query model |
| `gen_model` | Default generation model |
| `timeout` | Request timeout |
| `enabled` | Active status |
| `is_default` | Default provider flag |

---

## Key Patterns

### Pipeline Orchestration Pattern
- **Location**: `internal/application/orchestrator/runner.go`
- **Frequency**: Every generation request
- **Pattern**: Sequential agent execution with state passing via `AgentInput`/`AgentOutput` chain

### Event Streaming Pattern
- **Location**: `internal/application/orchestrator/session.go:34-42`, `internal/domain/agent/events.go`
- **Frequency**: All generation modes
- **Pattern**: Go channels for event propagation, SSE for client delivery

### Agent State Clone Pattern
- **Location**: `internal/application/orchestrator/session.go:144-294`
- **Frequency**: All state transitions
- **Pattern**: Deep copy of state objects to prevent mutation

### Model Selection Pattern
- **Location**: `internal/application/agents/modelselection/modelselection.go`
- **Frequency**: Every LLM call
- **Pattern**: Metadata override -> Config default -> Provider default

---

## Recommendations

1. **Add API Documentation**: Generate OpenAPI spec from router definitions for external consumers

2. **Batch Result Persistence**: Currently batch results stored in memory only; consider persistence for long retrieval

3. **Resume Feature Completion**: `Resume` method in runner exists but may need additional handler exposure

4. **WebSocket Alternative**: SSE works well for one-way streaming; consider WebSockets for bidirectional control (cancel, pause)

5. **Rate Limiting**: Add rate limiting to generation endpoints to prevent abuse

6. **Result Caching**: Consider caching retriever results by prompt hash for repeated similar queries
