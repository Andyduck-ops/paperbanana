# Image Display Flow Analysis

## Executive Summary

This document traces the complete data flow from image generation to display, identifying **four critical breakpoints** that explain why generated images may not display correctly.

---

## Data Flow Diagram

```
+------------------+     +-------------------+     +------------------+
|   LLM Provider   |     |   Visualizer      |     |   Orchestrator   |
| (Gemini/OpenAI)  |---->|   Agent           |---->|   Runner         |
|  Image Response  |     |   executeDiagram  |     |   Session        |
+------------------+     +-------------------+     +------------------+
        |                        |                         |
        v                        v                         v
  image bytes               Artifact{                   SessionState{
  in response.Parts          ID: "visualizer-diagram-rendered"   FinalOutput.GeneratedArtifacts[]
  MIMEType: image/png        Kind: "rendered_figure"             [Artifact...]
                            MIMEType: "image/png"               }
                            URI: "memory://visualizer/..."
                            Bytes: [raw bytes]
                          }

        |                                                   |
        | SSE Event: "result"                               |
        v                                                   v
+-------------------+     +-------------------+     +------------------+
|   generate.go     |     |   JSON Serialize  |     |   Frontend       |
|   buildGenerate   |---->|   Artifact struct |---->|   SSE Client     |
|   Response()      |     |   cloneArtifacts  |     |   onResult       |
+-------------------+     +-------------------+     +------------------+
        |                                                   |
        | GenerateResponse{                                 |
        |   generated_artifacts: [Artifact]                  |
        | }                                                 |
        v                                                   v
+------------------+     +-------------------+     +------------------+
|   Frontend       |     |   ArtifactPreview |     |   Browser DOM    |
|   useGenerate.ts |---->|   imageUrl calc   |---->|   <img src="...">|
|   onResult       |     +-------------------+     +------------------+
+------------------+
```

---

## Critical Findings

### 1. Artifact Struct Definition (Backend)

**Location**: `internal/domain/agent/types.go:87-95`

```go
type Artifact struct {
    ID       string            `json:"id"`
    Kind     ArtifactKind      `json:"kind"`
    MIMEType string            `json:"mime_type"`
    URI      string            `json:"uri"`
    Content  string            `json:"content,omitempty"`
    Bytes    []byte            `json:"bytes,omitempty"`        // <-- BASE64 ENCODED BY GIN
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

**Key observation**: The struct has `Bytes` field but NO `asset_id` field. The JSON serialization uses `bytes` (base64 encoded by Gin).

---

### 2. Visualizer Agent Output (Backend)

**Location**: `internal/application/agents/visualizer/agent.go:420-430`

```go
func renderedArtifact(mode domainagent.VisualMode, mimeType string, bytes []byte) domainagent.Artifact {
    return domainagent.Artifact{
        ID:       fmt.Sprintf("visualizer-%s-rendered", mode),
        Kind:     domainagent.ArtifactKindRenderedFigure,
        MIMEType: mimeType,
        URI:      fmt.Sprintf("memory://visualizer/%s/rendered", mode),
        Bytes:    append([]byte(nil), bytes...),  // Raw bytes stored
        Metadata: map[string]string{"mode": string(mode)},
    }
}
```

**Key observation**: Artifacts are created with raw `Bytes` and a `memory://` URI. No `asset_id` is ever set during generation.

---

### 3. Frontend Artifact Processing (Frontend)

**Location**: `web/src/hooks/useGenerate.ts:170-176`

```typescript
artifacts: data.generated_artifacts.map((a) => ({
    kind: a.kind,
    mimeType: a.mime_type,
    summary: a.summary,
    data: a.data,                                        // <-- Expecting base64 here
    assetId: (a as { asset_id?: string }).asset_id,      // <-- NEVER POPULATED!
})),
```

**Location**: `web/src/components/ArtifactPreview.tsx:21-25`

```typescript
const imageUrl = artifact.data
    ? `data:${artifact.mimeType};base64,${artifact.data}`
    : artifact.assetId
    ? `/api/v1/assets/${artifact.assetId}`               // <-- NEVER REACHED
    : null;
```

**Key observation**: Frontend expects either:
1. `data` field (base64 encoded) - should work
2. `assetId` field - **NEVER populated by backend**

---

## BREAKPOINT #1: Field Name Mismatch

**Backend sends**: `bytes` (base64 encoded by Gin JSON serializer)

**Frontend expects**: `data` (base64 string)

### Evidence:

| Backend JSON Key | Frontend Expected Key | Status |
|-----------------|----------------------|--------|
| `bytes` | `data` | **MISMATCH** |
| `id` | `assetId` | **MISMATCH** |

### Code Flow:

```
Backend Artifact.Bytes (raw bytes)
    |
    v (Gin JSON serialization)
Backend JSON: {"bytes": "base64encoded...", ...}
    |
    v (SSE transmission)
Frontend SSE parsing: a.data = undefined (no "data" field)
    |
    v
ArtifactPreview: artifact.data is undefined, falls through to assetId
    |
    v
ArtifactPreview: assetId is undefined, imageUrl = null
    |
    v
NO IMAGE DISPLAYED
```

---

## BREAKPOINT #2: Asset ID Never Populated

**Location**: Asset endpoint path construction

**Backend route**: `GET /api/v1/assets/:project_id/:asset_id`

**Frontend call**: `/api/v1/assets/${artifact.assetId}` (single parameter)

### The Problem:

The asset endpoint requires TWO parameters:
- `project_id` - the project the asset belongs to
- `asset_id` - the unique asset identifier

But `ArtifactPreview.tsx` only provides ONE parameter:
```typescript
`/api/v1/assets/${artifact.assetId}`  // Missing project_id!
```

Even if `assetId` were populated, the API call would be malformed.

---

## BREAKPOINT #3: No Persistence for In-Memory Artifacts

**The visualizer creates artifacts with**:
- `URI: "memory://visualizer/..."` - indicates in-memory storage
- `Bytes: [raw bytes]` - stored in memory only

**Missing**: No code path that:
1. Stores artifacts to `localstore`
2. Creates `workspace.Asset` records
3. Populates `asset_id` or `storage_key`

The `AssetService` exists but is never called during generation:
- `internal/application/persistence/asset_service.go:Write()`
- `internal/infrastructure/assets/localstore/store.go:Write()`

---

## BREAKPOINT #4: Batch Generation Artifact Structure

**Location**: `web/src/types/batch.ts:1-5`

```typescript
export interface BatchArtifact {
  id: string;
  kind: string;
  mimeType: string;
  // NO 'data' field!
  // NO 'assetId' field!
}
```

**Location**: `internal/api/dto/batch.go:44`

```go
type CandidateResultDTO struct {
    CandidateID int                      `json:"candidate_id"`
    SessionID   string                   `json:"session_id"`
    Status      string                   `json:"status"`
    Artifacts   []domainagent.Artifact   `json:"artifacts,omitempty"`  // Uses same Artifact struct
    Error       *domainagent.ErrorDetail `json:"error,omitempty"`
}
```

Batch artifacts also use `domainagent.Artifact` with `bytes` field, but frontend `BatchArtifact` type has no data field.

---

## Asset API Endpoint Verification

**Location**: `internal/api/router.go:118-120`

```go
v1.GET("/assets/:project_id/:asset_id", assetHandler.GetAsset)
v1.GET("/assets/:project_id/:asset_id/download", assetHandler.DownloadAsset)
v1.GET("/assets/version/:project_id/:version_id", assetHandler.ListAssetsByVersion)
```

**Confirmed**: The `/api/v1/assets/:project_id/:asset_id` endpoint EXISTS, but:
1. It requires `project_id` parameter
2. Frontend only sends `assetId` (missing project context)
3. No artifacts are ever persisted to this system

---

## MIME Type Flow

**Visualizer agent sets MIMEType**:
```go
// From imageArtifactFromResponse()
mimeType := part.MIMEType
if mimeType == "" {
    mimeType = "image/png"
}
```

**Asset store preserves MIME type in sidecar file**:
```go
// localstore/store.go:108-114
if mimeType != "" {
    mimePath := fullPath + ".mime"
    os.WriteFile(mimePath, []byte(mimeType), 0644)
}
```

**MIME type flow**: Correct - the visualizer sets it, and asset store can preserve it. However, since artifacts are never stored, this path is never exercised.

---

## Root Cause Summary

| Breakpoint | Root Cause | Impact |
|------------|------------|--------|
| #1 | Backend uses `bytes`, frontend expects `data` | Image data never reaches display |
| #2 | Backend uses `id`, frontend expects `assetId` | Asset lookup would fail |
| #3 | No persistence layer called during generation | No retrievable asset ID exists |
| #4 | BatchArtifact type missing data fields | Batch images cannot display |

---

## Recommended Fixes

### Fix #1: Align JSON field names (Quick Fix)

**Option A - Backend** (`internal/domain/agent/types.go`):
```go
type Artifact struct {
    ID       string            `json:"id"`
    Kind     ArtifactKind      `json:"kind"`
    MIMEType string            `json:"mime_type"`
    URI      string            `json:"uri"`
    Content  string            `json:"content,omitempty"`
    Bytes    []byte            `json:"data,omitempty"`  // Changed from "bytes" to "data"
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

**Option B - Frontend** (`web/src/hooks/useGenerate.ts`):
```typescript
artifacts: data.generated_artifacts.map((a) => ({
    kind: a.kind,
    mimeType: a.mime_type,
    summary: a.summary,
    data: a.bytes,  // Changed from a.data to a.bytes
    assetId: a.id,  // Changed from asset_id to id
})),
```

### Fix #2: Implement artifact persistence (Complete Fix)

1. Add artifact storage step after visualizer execution
2. Store bytes via `AssetService.Write()`
3. Set `AssetID` on the artifact
4. Frontend constructs proper URL: `/api/v1/assets/${projectId}/${assetId}`

### Fix #3: Fix asset URL construction

**Frontend** (`web/src/components/ArtifactPreview.tsx`):
```typescript
const imageUrl = artifact.data
    ? `data:${artifact.mimeType};base64,${artifact.data}`
    : artifact.assetId && artifact.projectId
    ? `/api/v1/assets/${artifact.projectId}/${artifact.assetId}/download`
    : null;
```

---

## Test Verification Points

1. **SSE Event Inspection**: Check `result` event payload for `bytes` vs `data` field
2. **Network Tab**: Verify `/api/v1/assets/...` call (should 404 with current code)
3. **Console Logging**: Add `console.log(artifact)` in `onResult` handler
4. **Backend Logging**: Add logging in `buildGenerateResponse()` to see artifact structure

---

## Files Analyzed

| File | Purpose |
|------|---------|
| `internal/api/handlers/generate.go` | SSE handler, builds response |
| `internal/application/orchestrator/session.go` | Session tracking |
| `internal/infrastructure/assets/localstore/store.go` | Asset storage |
| `internal/application/agents/visualizer/agent.go` | Image generation |
| `internal/domain/agent/types.go` | Artifact struct definition |
| `internal/api/handlers/assets.go` | Asset endpoint handlers |
| `internal/api/router.go` | Route definitions |
| `web/src/hooks/useGenerate.ts` | Result processing |
| `web/src/hooks/useBatchGeneration.ts` | Batch result processing |
| `web/src/components/ArtifactPreview.tsx` | Image display |
| `web/src/lib/sse.ts` | SSE client |
| `web/src/types/batch.ts` | Batch types |
| `internal/api/dto/batch.go` | Batch DTOs |
