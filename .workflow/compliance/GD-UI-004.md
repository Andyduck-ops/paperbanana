# GD-UI-004 Compliance Report

## Golden Case
```yaml
id: GD-UI-004
title: Resumed Task Exposes Resume Semantics
intent: When a task was resumed from a snapshot, the UI must not present it as a fresh execution.
expected:
  ui_must_show:
    - stages completed before resume are shown as already done
    - current stage is the resumed stage, not retriever
    - some indication that this is a resumed run
anti_patterns:
  - resumed task appears identical to a brand new task
  - prior completed stages shown as pending again
```

## Files Analyzed
- `web/src/hooks/useGenerate.ts`
- `web/src/lib/sse.ts`
- `web/src/types/api.ts`
- `web/src/components/ProgressPanel.tsx`
- `web/src/App.tsx`
- `internal/domain/agent/events.go`
- `internal/application/orchestrator/runner.go`
- `internal/application/orchestrator/session.go`
- `internal/domain/agent/types.go`
- `internal/api/handlers/generate.go`

## Verification Results

### [MISSING] Backend emits `resume_start` event
- **Status**: CRITICAL GAP
- **Evidence**: `internal/domain/agent/events.go` defines only these event types:
  ```go
  EventRunStarted     EventType = "run_started"
  EventStageStarted   EventType = "stage_started"
  EventStageCompleted EventType = "stage_completed"
  EventStageFailed    EventType = "stage_failed"
  EventRunCompleted   EventType = "run_completed"
  EventRunFailed      EventType = "run_failed"
  EventRunCanceled    EventType = "run_canceled"
  ```
- **NO `EventResumeStarted` or equivalent event type exists**
- **Consequence**: Frontend's `onResumeStart` handler will NEVER be triggered

### [MISSING] `resumed_from_stage` field in SSE event
- **Status**: CRITICAL GAP
- **Evidence**: Backend has `RestoreMetadata` struct in `types.go`:
  ```go
  type RestoreMetadata struct {
      SnapshotVersion string    `json:"snapshot_version"`
      RestoredFrom    StageName `json:"restored_from"`
      RestoredAt      time.Time `json:"restored_at"`
      ResumeToken     string    `json:"resume_token"`
  }
  ```
- **This metadata exists internally but is NEVER emitted to the frontend**
- Runner's `resumeTracker()` populates `restore.RestoredFrom` but only uses it internally

### [PARTIAL] Frontend `resumeMetadata` handling
- **Status**: PARTIALLY IMPLEMENTED
- **Evidence**: Frontend has complete infrastructure that expects backend data:
  - `web/src/lib/sse.ts` lines 50-54:
    ```typescript
    export interface ResumeStartEvent {
      resumed_from_stage: string;
      stages_completed_before_resume: string[];
      session_id: string;
    }
    ```
  - `web/src/hooks/useGenerate.ts` lines 16-19:
    ```typescript
    export interface ResumeMetadata {
      resumed_from_stage: string;
      stages_completed_before_resume: string[];
    }
    ```
  - `web/src/hooks/useGenerate.ts` lines 178-194 has `onResumeStart` handler
- **Gap**: Handler exists but will never be called because backend doesn't emit the event

### [MISSING] UI consumption of `resumeMetadata`
- **Status**: NOT CONNECTED
- **Evidence**: `App.tsx` line 40-41:
  ```typescript
  const { isGenerating, stages, result, error, generate, reset } = useGenerate();
  ```
- **`resumeMetadata` is NOT destructured or used**
- **`ProgressPanel` does not receive `resumeMetadata` prop**
- **No visual indication of resume status in UI**

## Critical Gap Analysis

### Backend Gap
The orchestrator correctly tracks resume metadata internally:
- `runner.go:398`: `restore.RestoredFrom = snapshot.Stage.Stage`
- `session.go`: `state.Restore` field preserves restore metadata
- `runner.go:315-319`: `newRestoredSessionTracker()` creates tracker with completed stages

**But this data is never exposed via SSE events.**

### Frontend Gap
The frontend infrastructure is ready:
- Event type `resume_start` defined in `sse.ts`
- `ResumeStartEvent` interface matches expected schema
- `onResumeStart` handler in `useGenerate.ts` correctly marks stages complete
- `resumeMetadata` state exists in `GenerateState`

**But the data flow is broken because:**
1. Backend never emits `resume_start` event
2. App.tsx ignores `resumeMetadata` even if populated

## Required Fixes

### 1. Backend: Add `EventResumeStarted` event type
**File**: `internal/domain/agent/events.go`
```go
EventResumeStarted EventType = "resume_start"
```

### 2. Backend: Emit resume event in runner
**File**: `internal/application/orchestrator/runner.go`
- After `newRestoredSessionTracker()`, emit `EventResumeStarted` with:
  - `resumed_from_stage`: snapshot.Stage.Stage
  - `stages_completed_before_resume`: names of completed stages from snapshot
  - `session_id`: snapshot.Session.SessionID

### 3. Frontend: Connect `resumeMetadata` to UI
**File**: `web/src/App.tsx`
- Destructure `resumeMetadata` from `useGenerate()`
- Pass to `ProgressPanel` or render resume indicator separately

### 4. Frontend: Display resume indication
**File**: `web/src/components/ProgressPanel.tsx`
- Add `resumeMetadata?: ResumeMetadata` prop
- Show visual indicator (badge, icon) when `resumeMetadata` is present
- Optionally mark stages that were completed before resume

## Compliance Status
**MISSING** - Critical gaps in both backend and frontend

| Requirement | Backend | Frontend | Status |
|-------------|---------|----------|--------|
| `resumed_from_stage` field | MISSING | DEFINED | NOT WIRED |
| Resume identifier UI | N/A | MISSING | NOT IMPLEMENTED |
| Completed stages stay complete | IMPLEMENTED | IMPLEMENTED | BLOCKED BY EVENT |
| `resume_start` event emission | MISSING | EXPECTED | CRITICAL GAP |

## Test Coverage
- `web/src/test/golden/GD-UI-004.test.tsx` exists with test cases
- Tests use `ResumableStageState` with `wasCompletedBeforeResume` marker
- Tests assume `ProgressPanel` receives resume metadata (not currently implemented)
