# Deep Dive: Image Display and Download Flow Analysis

**Report ID**: DEEP-006
**Date**: 2026-03-25
**Analyst**: Team Worker (analyst role)
**Status**: Completed

---

## Executive Summary

This report provides a comprehensive analysis of the image display, rendering, and download pipeline in paperbanana-clean. The analysis reveals **4 critical breakpoints** in the image display flow and identifies key architectural gaps between the frontend expectations and backend implementations.

### Key Findings

| Issue | Severity | Root Cause | Impact |
|-------|----------|------------|--------|
| Field name mismatch (`bytes` vs `data`) | **P0 Critical** | Backend uses `bytes`, frontend expects `data` | Images never display |
| Missing `asset_id` in SSE events | **P0 Critical** | No persistence step after generation | No fallback display path |
| Asset API URL missing `project_id` | **P1 High** | Frontend URL format incorrect | 404 errors even with valid assetId |
| Batch results in-memory only | **P1 High** | No database persistence | Lost on server restart |

---

## 1. Complete Data Flow Diagram

### 1.1 Generation to Display Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           BACKEND - Generation Phase                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐    ┌─────────────────┐    ┌────────────────────────────┐ │
│  │ Visualizer   │───►│ Artifact        │───►│ SSE Event (result)        │ │
│  │ Agent        │    │ {               │    │ {                         │ │
│  │              │    │   ID: string    │    │   session_id: string,     │ │
│  │ Creates:     │    │   Kind: string  │    │   generated_artifacts: [{ │ │
│  │ - Bytes      │    │   MIMEType: str │    │     kind: string,         │ │
│  │ - URI        │    │   URI: string   │    │     mime_type: string,    │ │
│  │ - Content    │    │   Bytes: []byte │    │     summary: string,      │ │
│  │              │    │   Content: str  │    │     data?: string         │ ◄── MISSING: bytes serialized
│  │ Note:        │    │ }               │    │   }]                      │     as base64 with field "bytes"
│  │ NO asset_id! │    │                 │    │ }                         │     but frontend expects "data"!
│  └──────────────┘    └─────────────────┘    └────────────────────────────┘ │
│         │                                           │                       │
│         │ ❌ NO PERSISTENCE                         │                       │
│         ▼                                           ▼                       │
│  ┌──────────────┐                          ┌────────────────────────────┐  │
│  │ Asset Store  │                          │ Gin JSON Serialization    │  │
│  │ EXISTS but   │                          │ - Bytes []byte → base64   │  │
│  │ NEVER USED   │                          │ - Field name: "bytes"     │  │
│  │ in pipeline  │                          └────────────────────────────┘  │
│  └──────────────┘                                                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ SSE Stream
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           FRONTEND - Display Phase                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────────────────┐    ┌────────────────────────────────────────┐   │
│  │ useGenerate.ts        │    │ ArtifactPreview.tsx                    │   │
│  │ onResult callback     │    │                                        │   │
│  │                       │    │ imageUrl = artifact.data               │   │
│  │ Expects:              │    │   ? `data:${mime};base64,${data}`      │   │
│  │ - kind: string        │    │   : artifact.assetId                   │   │
│  │ - mime_type: string   │    │   ? `/api/v1/assets/${assetId}`        │   │
│  │ - summary: string     │    │   : null                               │   │
│  │ - data?: string       │    │                                        │   │
│  │ - asset_id?: string   │    │ Result:                                │   │
│  │                       │    │ - data = undefined (field is "bytes")  │   │
│  │ Gets:                 │    │ - assetId = undefined (never set)      │   │
│  │ - kind, mime_type,    │    │ - imageUrl = null                      │   │
│  │   summary OK          │    │ - IMAGE DOES NOT DISPLAY               │   │
│  │ - data: undefined     │    │                                        │   │
│  │ - asset_id: undefined │    └────────────────────────────────────────┘   │
│  └───────────────────────┘                                                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Download Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SINGLE DOWNLOAD                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ExportModal.tsx                                                            │
│  ├─ Requires: imageData (base64 string)                                     │
│  ├─ Uses: exportAsPng(canvas, { dpi })                                      │
│  └─ Problem: artifact.data is undefined → no imageData → cannot export      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           BATCH DOWNLOAD (ZIP)                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  BatchProgressPanel.tsx                                                      │
│  └─ handleDownloadZip()                                                      │
│      └─ POST /api/v1/batch/download { batch_id }                            │
│                                                                              │
│  batch.go:DownloadBatchZip                                                   │
│  ├─ Retrieves result from BatchRunner.results (in-memory map)               │
│  ├─ Creates ZIP with artifact.Bytes                                          │
│  └─ Problem: BatchRunner.results is memory-only, lost on restart            │
│                                                                              │
│  ZIP Contents:                                                               │
│  ├─ candidate_0_artifact_0.png                                               │
│  ├─ candidate_1_artifact_0.png                                               │
│  └─ metadata.json                                                            │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Critical Breakpoints in Image Display

### 2.1 Breakpoint #1: Field Name Mismatch

**Location**: `internal/domain/agent/types.go:87-95` vs `web/src/lib/sse.ts:34-42`

**Backend Artifact Structure**:
```go
type Artifact struct {
    ID       string            `json:"id"`
    Kind     ArtifactKind      `json:"kind"`
    MIMEType string            `json:"mime_type"`
    URI      string            `json:"uri"`
    Content  string            `json:"content,omitempty"`
    Bytes    []byte            `json:"bytes,omitempty"`      // ← Field name: "bytes"
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

**Frontend ResultEvent Interface**:
```typescript
export interface ResultEvent {
  session_id: string;
  generated_artifacts: Array<{
    kind: string;
    mime_type: string;
    summary: string;
    data?: string;              // ← Field name: "data"
  }>;
}
```

**Impact**: Gin's JSON serializer converts `[]byte` to base64 automatically, but under the field name `bytes`. Frontend expects `data` and never receives it.

### 2.2 Breakpoint #2: Missing Persistence Step

**Location**: `internal/application/agents/visualizer/agent.go:420-431`

```go
func renderedArtifact(mode domainagent.VisualMode, mimeType string, bytes []byte) domainagent.Artifact {
    return domainagent.Artifact{
        ID:       fmt.Sprintf("visualizer-%s-rendered", mode),
        Kind:     domainagent.ArtifactKindRenderedFigure,
        MIMEType: mimeType,
        URI:      fmt.Sprintf("memory://visualizer/%s/rendered", mode),  // ← memory:// URI
        Bytes:    append([]byte(nil), bytes...),
        Metadata: map[string]string{
            "mode": string(mode),
        },
    }
}
```

**Analysis**:
- Artifacts are created with `memory://` URIs indicating no persistence
- The full `AssetStore` infrastructure exists (`localstore/store.go`) but is never called
- No `asset_id` is ever generated or attached to artifacts

### 2.3 Breakpoint #3: Asset API URL Format Mismatch

**Location**: `web/src/components/ArtifactPreview.tsx:21-25` vs `internal/api/handlers/assets.go:124-126`

**Frontend URL Construction**:
```typescript
const imageUrl = artifact.data
  ? `data:${artifact.mimeType};base64,${artifact.data}`
  : artifact.assetId
  ? `/api/v1/assets/${artifact.assetId}`  // ← Missing project_id!
  : null;
```

**Backend Route Definition**:
```go
// GET /api/v1/assets/:project_id/:asset_id
func (h *AssetHandler) GetAsset(c *gin.Context) {
    projectID := c.Param("project_id")
    assetID := c.Param("asset_id")
    // Both required!
}
```

**Impact**: Even if `assetId` were populated, the URL would be incorrect. The correct format requires both `project_id` and `asset_id`.

### 2.4 Breakpoint #4: Batch Results Memory-Only Storage

**Location**: `internal/application/orchestrator/batch_runner.go`

```go
type BatchRunner struct {
    agentFactory  AgentFactory
    maxConcurrent int
    eventBuffer   int
    results       map[string]*domainagent.BatchResult  // ← In-memory map only!
    mu            sync.RWMutex
}
```

**Impact**:
- Batch results lost on server restart
- Download ZIP unavailable after restart
- Cannot resume batch operations

---

## 3. Download Functionality Analysis

### 3.1 Single Image Download

| Component | Status | Notes |
|-----------|--------|-------|
| ExportModal.tsx | Partially functional | Requires `artifact.data` (base64) to work |
| export.ts | Functional | `exportAsPng`, `exportAsSvg`, `exportAsPdf` all work |
| Current state | **Broken** | No `artifact.data` available from SSE events |

**Code Path**:
```typescript
// ExportModal.tsx:32-40
if (format === 'png') {
  if (canvas) {
    await exportAsPng(canvas, { dpi });
  } else if (imageData) {  // ← imageData comes from artifact.data
    const img = new window.Image();
    img.src = `data:image/png;base64,${imageData}`;
    await new Promise((resolve) => { img.onload = resolve; });
    await exportAsPng(img, { dpi });
  }
}
```

### 3.2 Batch ZIP Download

| Component | Status | Notes |
|-----------|--------|-------|
| BatchProgressPanel.tsx | Functional | UI and download trigger work |
| POST /api/v1/batch/download | Functional | Creates valid ZIP |
| BatchRunner.results | **Memory-only** | Lost on restart |
| metadata.json | Included | Contains batch metadata |

**ZIP Structure**:
```
paperviz_candidates_20260325_143022.zip
├── candidate_0_artifact_0.png
├── candidate_1_artifact_0.png
├── candidate_2_artifact_0.png
└── metadata.json
```

### 3.3 Asset API Download

| Endpoint | Method | Parameters | Status |
|----------|--------|------------|--------|
| `/api/v1/assets/:project_id/:asset_id` | GET | project_id, asset_id | **Unused** |
| `/api/v1/assets/:project_id/:asset_id/download` | GET | project_id, asset_id | **Unused** |

**Note**: These endpoints exist and are functional, but are never called because:
1. No `asset_id` is ever generated
2. Frontend doesn't have `project_id` context for URL construction

---

## 4. Image Rendering Comparison: paperbanana-clean vs repo-cn

### 4.1 paperbanana-clean (Go)

| Aspect | Implementation | Issue |
|--------|---------------|-------|
| Image creation | Visualizer agent generates `Artifact.Bytes` | Raw bytes stored |
| Transport | SSE with JSON serialization | `bytes` field (not `data`) |
| Persistence | None (memory:// URIs) | No asset storage |
| Display fallback | None | Must have `data` or `assetId` |
| Download | Uses in-memory bytes | Works but lost on restart |

### 4.2 repo-cn (Python)

| Aspect | Implementation | Notes |
|--------|---------------|-------|
| Image creation | Visualizer generates image bytes | Same approach |
| Transport | JSON file persistence | Incremental saves every 10 results |
| Persistence | File-based JSON | Results survive restart |
| Display | Console/CLI only | No web interface |
| Download | File-based output | Results written to disk |

### 4.3 Key Differences

```
paperbanana-clean                         repo-cn
================                         =======
SSE streaming for web                    Console-only output
Memory-only batch results                File-persisted results
Session-level DB persistence             No session concept
Asset store EXISTS but UNUSED            No asset store
Sophisticated pipeline                   Simple linear pipeline
```

---

## 5. Display Scenarios Analysis

### 5.1 Successful Generation

| Step | Expected | Actual |
|------|----------|--------|
| Backend creates artifact | Bytes populated | Bytes populated |
| SSE sends result | `bytes` field serialized | `bytes` field serialized |
| Frontend receives | `data` field populated | `data` = undefined |
| Image displays | Shows image | **No image** |

### 5.2 Failed Generation

| Step | Expected | Actual |
|------|----------|--------|
| Backend returns error | Error event with details | Works correctly |
| Frontend shows error | Error message displayed | Works correctly |
| Partial results | Show what completed | Works if artifacts present |

### 5.3 Batch Generation

| Step | Expected | Actual |
|------|----------|--------|
| Progress updates | Real-time SSE events | Works correctly |
| Candidate completion | Each candidate shown | Status shown, but no image |
| Download ZIP | ZIP with all images | Works if server not restarted |

---

## 6. Root Cause Analysis

### 6.1 Primary Root Causes

1. **JSON Field Naming Convention Mismatch**
   - Go struct tags: `json:"bytes"`
   - Frontend interface: `data?: string`
   - No shared type definitions between frontend and backend

2. **Missing Integration Point**
   - Asset persistence infrastructure exists
   - No code path calls `AssetStore.Write()` after visualization
   - Pipeline ends at artifact creation, not artifact storage

3. **Context Propagation Gap**
   - `project_id` is captured in request metadata
   - Not propagated to artifact creation or display components
   - Frontend lacks project context for asset URL construction

### 6.2 Architectural Gap

```
Current Flow:
  Generation → Artifact.Bytes → SSE → Frontend → NO DISPLAY

Expected Flow:
  Generation → Artifact.Bytes → AssetStore.Write() → AssetID
            → SSE (with asset_id) → Frontend → Display via Asset API
```

---

## 7. Recommendations

### 7.1 Critical Fixes (P0)

#### Fix #1: Field Name Alignment

**Option A**: Rename backend field (Recommended)
```go
type Artifact struct {
    // ...
    Bytes []byte `json:"data,omitempty"`  // Changed from "bytes"
}
```

**Option B**: Update frontend interface
```typescript
generated_artifacts: Array<{
    kind: string;
    mime_type: string;
    summary: string;
    bytes?: string;  // Changed from "data"
}>;
```

#### Fix #2: Add Persistence Step

Location: After visualizer agent execution in `runner.go`

```go
// Pseudocode
func (r *Runner) persistArtifacts(ctx context.Context, output AgentOutput, projectID string) {
    for i, artifact := range output.GeneratedArtifacts {
        if len(artifact.Bytes) > 0 {
            key, err := r.assetStore.Write(ctx, projectID, sessionID, artifact.MIMEType, artifact.Bytes)
            if err == nil {
                output.GeneratedArtifacts[i].Metadata["asset_id"] = key
            }
        }
    }
}
```

### 7.2 High Priority Fixes (P1)

#### Fix #3: Update Asset URL Format

```typescript
// ArtifactPreview.tsx
const imageUrl = artifact.data
  ? `data:${artifact.mimeType};base64,${artifact.data}`
  : artifact.assetId
  ? `/api/v1/assets/${projectId}/${artifact.assetId}/download`  // Added project_id
  : null;
```

#### Fix #4: Add Batch Result Persistence

1. Create `BatchResultModel` in database schema
2. Update `BatchRunner.storeResult()` to persist to DB
3. Update `GetBatchResult()` to check DB if not in memory

### 7.3 Medium Priority (P2)

- Generate TypeScript types from Go structs (openapi-generator)
- Add asset cleanup job for old artifacts
- Implement batch-to-session mapping table

---

## 8. Test Verification Plan

### 8.1 Field Name Fix Verification

```bash
# After fix, SSE should return:
curl -N -X POST http://localhost:8080/api/v1/generate/stream \
  -H "Content-Type: application/json" \
  -d '{"prompt": "test diagram"}' | jq '.generated_artifacts[0]'

# Expected output should include "data" field with base64 content:
{
  "kind": "rendered_figure",
  "mime_type": "image/png",
  "summary": "...",
  "data": "iVBORw0KGgoAAAANSUhEUgAA..."  // ← Should be present
}
```

### 8.2 Asset API Verification

```bash
# Test asset download endpoint
curl http://localhost:8080/api/v1/assets/{project_id}/{asset_id}/download \
  --output test.png

# Verify file is valid PNG
file test.png  # Should show: PNG image data
```

### 8.3 Batch Download Verification

```bash
# Run batch generation
curl -X POST http://localhost:8080/api/v1/generate/batch \
  -H "Content-Type: application/json" \
  -d '{"prompt": "test", "num_candidates": 3}' \
  -o batch_stream.txt

# Extract batch_id from stream, then download
curl -X POST http://localhost:8080/api/v1/batch/download \
  -H "Content-Type: application/json" \
  -d '{"batch_id": "extracted_batch_id"}' \
  -o test.zip

# Verify ZIP contents
unzip -l test.zip
```

---

## 9. Related Files

| Category | File Path | Purpose |
|----------|-----------|---------|
| Backend Types | `internal/domain/agent/types.go` | Artifact struct definition |
| Backend Handler | `internal/api/handlers/generate.go` | SSE streaming handler |
| Backend Handler | `internal/api/handlers/batch.go` | Batch download handler |
| Backend Handler | `internal/api/handlers/assets.go` | Asset API endpoints |
| Backend Agent | `internal/application/agents/visualizer/agent.go` | Artifact creation |
| Backend Store | `internal/infrastructure/assets/localstore/store.go` | Asset persistence |
| Frontend SSE | `web/src/lib/sse.ts` | SSE client and types |
| Frontend Hook | `web/src/hooks/useGenerate.ts` | Generation state management |
| Frontend Display | `web/src/components/ArtifactPreview.tsx` | Image rendering |
| Frontend Panel | `web/src/components/BatchProgressPanel.tsx` | Batch progress UI |
| Frontend Modal | `web/src/components/ExportModal.tsx` | Export functionality |
| Frontend Export | `web/src/lib/export.ts` | Export utilities |
| Reference | `repo-cn/main.py` | Python implementation comparison |

---

## 10. Conclusion

The image display and download system has a fundamental architectural gap: the backend creates artifacts with raw bytes but never persists them, and the frontend expects either base64 data or asset IDs that don't exist. The immediate fix is to align field names (`bytes` → `data`), and the proper fix is to implement artifact persistence with asset ID generation.

The download functionality (single and batch) is architecturally sound but non-functional due to the display pipeline breakage. Once images display correctly, downloads will work automatically for batch operations, and single-image exports will work with the existing `export.ts` utilities.
