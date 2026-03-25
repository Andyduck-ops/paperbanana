# Security Issues Requiring Follow-up

## Critical Issues

### ISSUE-001: Python RCE Vulnerability
**Status**: OPEN
**Severity**: CRITICAL
**File**: `internal/application/agents/visualizer/plot_executor.go`
**Action Required**: Implement sandboxing before any production deployment

**Sub-tasks**:
- [ ] Research RestrictedPython feasibility
- [ ] Design Docker execution environment
- [ ] Implement resource limits
- [ ] Add execution timeout handling
- [ ] Create allowlist for matplotlib operations

## High Issues

### ISSUE-002: Encryption Key Management
**Status**: OPEN
**Severity**: HIGH
**File**: `internal/infrastructure/crypto/aesgcm/service.go`
**Action Required**: Add production mode validation

**Sub-tasks**:
- [ ] Add `--production` flag or `PAPERBANANA_PRODUCTION=true` env check
- [ ] Fail-fast when encryption key not set in production
- [ ] Replace `fmt.Printf` with structured logging

### ISSUE-003: CORS Configuration
**Status**: OPEN
**Severity**: HIGH
**File**: `internal/api/middleware/cors.go`
**Action Required**: Implement environment-aware CORS

**Sub-tasks**:
- [ ] Add `ALLOWED_ORIGINS` configuration option
- [ ] Default to strict mode in production builds
- [ ] Document CORS configuration in deployment guide

## Medium Issues

### ISSUE-004: Input Validation Gaps
**Status**: OPEN
**Severity**: MEDIUM
**File**: `internal/api/handlers/generate.go`
**Action Required**: Enhance input validation

**Sub-tasks**:
- [ ] Add maximum prompt length validation
- [ ] Validate model names against allowlist
- [ ] Add content encoding validation

## Low Issues

### ISSUE-005: Missing Security Headers
**Status**: OPEN
**Severity**: LOW
**File**: `internal/api/router.go`
**Action Required**: Add security headers middleware

**Sub-tasks**:
- [ ] Create SecurityHeaders middleware
- [ ] Add CSP header
- [ ] Add HSTS header (for HTTPS)

---

## Batch Persistence Issues

### ISSUE-006: Batch Results Not Persisted
**Status**: OPEN
**Severity**: HIGH
**File**: `internal/application/orchestrator/batch_runner.go`
**Action Required**: Implement database persistence for batch results

**Impact**:
- All batch results lost on server restart
- ZIP download returns 404 after restart
- Cannot resume interrupted batches

**Sub-tasks**:
- [ ] Create `BatchResultModel` in database schema
- [ ] Implement `BatchRepository` interface
- [ ] Modify `storeResult()` to persist to database
- [ ] Modify `GetBatchResult()` to fall back to database
- [ ] Add `expires_at` column for TTL cleanup
- [ ] Create database migration script

### ISSUE-007: Missing Batch-to-Session Mapping
**Status**: OPEN
**Severity**: MEDIUM
**File**: `internal/application/orchestrator/batch_runner.go`
**Action Required**: Track which sessions belong to which batch

**Impact**:
- Cannot reconstruct batch results from individual sessions
- Limited analytics capability
- No way to query all sessions in a batch

**Sub-tasks**:
- [ ] Create `BatchSessionModel` mapping table
- [ ] Add batch_id to session metadata during creation
- [ ] Implement query to get all sessions for a batch

---

## Image Display Flow Issues

### ISSUE-008: JSON Field Name Mismatch
**Status**: OPEN
**Severity**: CRITICAL
**File**: `internal/domain/agent/types.go`, `web/src/hooks/useGenerate.ts`
**Action Required**: Align field names between backend and frontend

**Description**: Backend `Artifact.Bytes` serializes as `bytes`, frontend expects `data`. Image data is never received by display component.

**Evidence**:
- Backend: `Bytes []byte `json:"bytes,omitempty"``
- Frontend: `data: a.data` (undefined)
- Frontend: `assetId: (a as { asset_id?: string }).asset_id` (undefined)

**Impact**: Generated images never display after generation completes.

**Sub-tasks**:
- [ ] Change backend to use `json:"data,omitempty"` OR
- [ ] Change frontend to read `a.bytes`
- [ ] Add shared type definitions to prevent regression

### ISSUE-009: Asset ID Never Populated
**Status**: OPEN
**Severity**: HIGH
**File**: `internal/application/agents/visualizer/agent.go`
**Action Required**: Implement artifact persistence step

**Description**: No code path sets `asset_id` on generated artifacts. The field is typed but never assigned.

**Evidence**:
- Visualizer creates artifacts with only: `ID`, `Kind`, `MIMEType`, `URI`, `Bytes`, `Metadata`
- No call to `AssetService.Write()` during generation
- `asset_id` field in JSON is always null/undefined

**Impact**: Asset-based display fallback is impossible - no asset exists to retrieve.

**Sub-tasks**:
- [ ] Add artifact storage step after visualizer execution
- [ ] Store bytes via `AssetService.Write()`
- [ ] Set `AssetID` on the artifact
- [ ] Add persistence configuration flag

### ISSUE-010: Malformed Asset API URL
**Status**: OPEN
**Severity**: HIGH
**File**: `web/src/components/ArtifactPreview.tsx`
**Action Required**: Fix asset URL construction

**Description**: Frontend constructs asset URL with single parameter, endpoint requires two parameters.

**Evidence**:
- Frontend: `/api/v1/assets/${artifact.assetId}` (1 param)
- Backend route: `/assets/:project_id/:asset_id` (2 params)

**Impact**: Asset retrieval would fail with 400/404 even if implemented.

**Sub-tasks**:
- [ ] Add `projectId` to artifact type
- [ ] Construct URL with both parameters
- [ ] Use `/download` endpoint for binary content

### ISSUE-011: BatchArtifact Type Incomplete
**Status**: OPEN
**Severity**: MEDIUM
**File**: `web/src/types/batch.ts`
**Action Required**: Add missing fields to BatchArtifact interface

**Description**: `BatchArtifact` type lacks `data` and `assetId` fields needed for display.

**Evidence**:
```typescript
export interface BatchArtifact {
  id: string;
  kind: string;
  mimeType: string;
  // Missing: data, assetId
}
```

**Impact**: Batch generation results cannot display images.

**Sub-tasks**:
- [ ] Add `data?: string` field
- [ ] Add `assetId?: string` field
- [ ] Update batch result processing to map fields correctly

### ISSUE-012: No Persistence Integration
**Status**: OPEN
**Severity**: MEDIUM
**File**: `internal/application/orchestrator/runner.go`
**Action Required**: Connect asset storage to generation pipeline

**Description**: Complete asset storage infrastructure exists but is not connected to generation pipeline.

**Evidence**:
- `localstore.Store` with `Write()` exists
- `AssetService` with `Write()` exists
- Asset endpoints exist
- No call from orchestrator/visualizer to storage

**Impact**: Images are ephemeral - lost after SSE response completes.

**Sub-tasks**:
- [ ] Add persistence hook in runner after visualizer
- [ ] Make persistence configurable (memory vs storage)
- [ ] Update artifact with storage key after persistence

---

## Data Dependency Issues

### ISSUE-013: PaperBananaBench Reference Data Missing
**Status**: OPEN
**Severity**: CRITICAL
**Files**: `data/PaperBananaBench/diagram/ref.json`, `data/PaperBananaBench/plot/ref.json`
**Action Required**: Populate benchmark data or add data acquisition mechanism

**Description**: All PaperBananaBench reference data files contain empty arrays (`[]`), resulting in zero-shot generation instead of few-shot.

**Evidence**:
- `data/PaperBananaBench/diagram/ref.json`: 3 bytes, content `[]`
- `data/PaperBananaBench/plot/ref.json`: 3 bytes, content `[]`
- Expected size: ~4.5MB (diagram) + ~900KB (plot)
- Missing: 241 diagram images + 480 plot images

**Impact**:
- No reference examples for few-shot learning
- No reference images for visual guidance
- Generation quality severely degraded
- Users unaware of degradation (silent failure)

**Sub-tasks**:
- [ ] Copy benchmark data from repo-cn
- [ ] Create data download script `scripts/download-benchmark.sh`
- [ ] Add startup data validation check
- [ ] Add frontend warning when retrieval returns 0 results
- [ ] Document data acquisition in README

### ISSUE-014: Lite Retrieval Mode Not Implemented
**Status**: OPEN
**Severity**: HIGH
**File**: `internal/application/agents/retriever/agent.go`
**Action Required**: Implement lite retrieval mode for 96% token savings

**Description**: Go only has one retrieval mode (equivalent to Python's `auto-full`). Python's `auto` (lite) mode reduces token usage from ~80K to ~3K.

**Evidence**:
```go
// Current modes
const (
    RetrievalModeAuto   RetrievalMode = "auto"
    RetrievalModeManual RetrievalMode = "manual"
    RetrievalModeRandom RetrievalMode = "random"
    RetrievalModeNone   RetrievalMode = "none"
)
// Missing: RetrievalModeAutoFull (full content)
// Auto should be lite (IDs only)
```

**Impact**:
- Unnecessary token consumption (25x more tokens)
- Higher costs per generation
- Slower response times

**Sub-tasks**:
- [ ] Rename current `auto` to `auto-full`
- [ ] Implement `auto` as lite mode (IDs only)
- [ ] Add mode selection to frontend
- [ ] Update documentation

### ISSUE-015: Missing Data Warning Not Implemented
**Status**: OPEN
**Severity**: MEDIUM
**File**: `internal/application/agents/retriever/agent.go`
**Action Required**: Add user-visible warning when data is missing or empty

**Description**: Python's retriever warns and falls back to `none` mode when data is missing. Go silently returns empty slices.

**Evidence**:
```python
# Python (repo-cn)
if retrieval_setting in ["auto", "auto-full", "random"] and not ref_file.exists():
    print(f"Warning: Reference file not found at {ref_file}. Falling back to retrieval_setting='none'.")
```

```go
// Go (paperbanana-clean)
if errors.Is(err, os.ErrNotExist) {
    return nil, nil  // Silent return
}
```

**Impact**: Users have no visibility into data quality issues.

**Sub-tasks**:
- [ ] Add warning log when candidates list is empty
- [ ] Return metadata flag indicating data availability
- [ ] Add frontend notification when retrieval degraded

---

## UI/UX Issues

### ISSUE-016: Header Missing History and Settings Buttons
**Status**: OPEN
**Severity**: HIGH
**File**: `web/src/components/Header.tsx`
**Action Required**: Add missing header action buttons

**Description**: Header component only displays theme and language switchers, but App.tsx passes `onHistoryClick` and `onSettingsClick` props that are never used.

**Impact**:
- Users cannot access history panel from header
- Users cannot access settings drawer from header
- Features exist but are undiscoverable

**Sub-tasks**:
- [ ] Add History icon button with count badge
- [ ] Add Settings icon button
- [ ] Connect to App.tsx callback props

### ISSUE-017: EmptyState Example Click Non-functional
**Status**: OPEN
**Severity**: MEDIUM
**File**: `web/src/components/workspace/EmptyState.tsx`
**Action Required**: Implement example prompt loading

**Description**: EmptyState dispatches `workspace:loadExample` custom event when example cards are clicked, but no parent component listens for this event.

**Impact**:
- Example cards appear clickable but do nothing
- Users cannot quick-start with example prompts

**Sub-tasks**:
- [ ] Add event listener in App.tsx or Workspace.tsx
- [ ] Populate GeneratePanel fields from event detail
- [ ] Alternatively, pass callback prop instead of using custom event

### ISSUE-018: Unfriendly Error Messages
**Status**: OPEN
**Severity**: MEDIUM
**File**: `web/src/hooks/useGenerate.ts`, `web/src/hooks/useBatchGeneration.ts`
**Action Required**: Add actionable error messages

**Description**: Error states display raw HTTP status codes or generic messages without guidance on how to resolve them.

**Impact**:
- Users see "HTTP 401" instead of "API Key invalid"
- Users see "HTTP 429" instead of "Too many requests, please wait"
- Poor user experience for error recovery

**Sub-tasks**:
- [ ] Create error message mapping table
- [ ] Add error code to user-friendly message converter
- [ ] Include actionable suggestions in error display

### ISSUE-019: Refine Mode Lacks Stage Progress
**Status**: OPEN
**Severity**: MEDIUM
**File**: `web/src/components/workspace/Workspace.tsx`
**Action Required**: Add progress stages for refinement

**Description**: Generation mode shows detailed stage progress (Retriever, Planner, etc.), but Refine mode only shows a spinner text.

**Impact**:
- Users have no visibility into refinement progress
- Cannot identify where refinement fails

**Sub-tasks**:
- [ ] Define refinement stages (upload, process, generate)
- [ ] Update useRefine hook to report stages
- [ ] Render stages in Workspace component

### ISSUE-020: Missing Token Cost Warnings
**Status**: OPEN
**Severity**: MEDIUM
**File**: `web/src/components/ConfigPanel.tsx`
**Action Required**: Add token consumption warnings

**Description**: Unlike Streamlit version which clearly warns about token costs ("auto-full: ~800K tokens per candidate"), the React UI provides no cost information.

**Impact**:
- Users may accidentally incur high API costs
- No transparency about retrieval mode costs

**Sub-tasks**:
- [ ] Add token estimates for each retrieval mode
- [ ] Display warning for expensive options
- [ ] Consider adding cost summary before submission
