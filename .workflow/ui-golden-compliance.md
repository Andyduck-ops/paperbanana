# UI Golden Data Compliance Report

**Date**: 2026-03-23
**Files Modified**:
- `web/src/hooks/useGenerate.ts`
- `web/src/components/StageCard.tsx`
- `web/src/components/ProgressPanel.tsx`
- `web/src/components/EvolutionTimeline.tsx`
- `web/src/lib/sse.ts`

---

## Summary of Changes

### 1. Fixed stageOrder (GD-UI-001)
**Before**: `['retriever', 'planner', 'visualizer', 'critic']` (4 stages)
**After**: `['retriever', 'planner', 'stylist', 'visualizer', 'critic']` (5 stages)

### 2. Added Resume Metadata Support (GD-UI-004)
- New `ResumeMetadata` interface with `resumed_from_stage` and `stages_completed_before_resume`
- New `resume_start` SSE event type
- `onResumeStart` handler marks previously completed stages

### 3. Added `not_run` Stage Status (GD-UI-002)
- New `not_run` status in `StageStatus` type
- Visual styling (gray, dash icon) for stages that won't run after failure
- `stagesNotRun` array in state to track skipped stages

### 4. Enhanced Error Handling (GD-UI-002)
- `onError` now marks failed stage with `error` status
- Subsequent stages are marked as `not_run`
- `stagesNotRun` array populated with stages after failure point

---

## Golden Case Compliance

### GD-UI-001: Stage Progress Visible While Running

| Requirement | Status | Notes |
|-------------|--------|-------|
| Display current active stage name | ✅ | `onStageStart` sets status to `running` |
| Display completed stages | ✅ | `onStageComplete` sets status to `complete` with summary |
| Display pipeline structure | ✅ | 5-stage pipeline now complete: retriever, planner, stylist, visualizer, critic |
| Avoid generic spinner without stage info | ✅ | Each stage has name, agent, and status |

**Verdict**: COMPLIANT

---

### GD-UI-002: Failure Location Visible On Stage Error

| Requirement | Status | Notes |
|-------------|--------|-------|
| Display failure status | ✅ | Failed stage has `status: 'error'` with error message |
| Display failed stage name | ✅ | `data.stage` identifies which stage failed |
| Display completed stages before failure | ✅ | Completed stages retain `status: 'complete'` |
| Distinguish completed vs not-run stages | ✅ | New `not_run` status for stages after failure point |

**Anti-patterns Avoided**:
- ❌ Generic "generation failed" message → Specific stage failure with error details
- ❌ All stages show as failed → Only failed stage shows error, completed stay complete, unstarted show not_run

**Verdict**: COMPLIANT

---

### GD-UI-003: Artifact Surfaced On Completion

| Requirement | Status | Notes |
|-------------|--------|-------|
| Display completion status | ✅ | `onResult` sets `isGenerating: false` with result |
| Final artifact accessible | ✅ | `GenerateResult` contains artifacts array |
| All stages show complete | ✅ | Each stage transitions through `running` → `complete` |

**Anti-patterns Avoided**:
- ❌ Completion mark but empty artifact area → `result.artifacts` populated from SSE

**Verdict**: COMPLIANT

---

### GD-UI-004: Resumed Task Exposes Resume Semantics

| Requirement | Status | Notes |
|-------------|--------|-------|
| Display resume indication | ✅ | `resumeMetadata` contains `resumed_from_stage` |
| Completed stages stay complete | ✅ | `onResumeStart` marks `stages_completed_before_resume` as complete |
| Resume from correct stage | ✅ | Backend provides `resumed_from_stage` to continue |

**Anti-patterns Avoided**:
- ❌ Resumed task looks like fresh start → Resume metadata and pre-completed stages distinguish from new runs

**Verdict**: COMPLIANT (frontend ready, backend must emit `resume_start` event)

---

### GD-UI-005: Batch Per-Task Status Visible

| Requirement | Status | Notes |
|-------------|--------|-------|
| Display each task status | ⚠️ | Current UI shows single pipeline stages, not batch task list |
| Show running/complete/failed per task | ⚠️ | Batch mode requires additional UI component |

**Analysis**: The current implementation focuses on single-pipeline stage progress. Batch processing with multiple parallel tasks would require:
1. New UI component for batch task list
2. Different state shape to track multiple tasks
3. Backend SSE events for batch task updates

**Verdict**: PARTIALLY COMPLIANT - Single pipeline fully supported, batch mode needs additional work

---

## Outstanding Items

| Item | Priority | Description |
|------|----------|-------------|
| Batch UI Component | Medium | Create `BatchTaskList` component for GD-UI-005 |
| Backend `resume_start` Event | High | Backend must emit this for GD-UI-004 to work |
| i18n for Resume UI | Low | Add translation keys for "Resumed from..." indicator |

---

## Files Changed

```
web/src/hooks/useGenerate.ts    - Core state management, 5 stages, resume, error handling
web/src/components/StageCard.tsx - Added 'not_run' status styling
web/src/components/ProgressPanel.tsx - Re-export StageStatus type
web/src/components/EvolutionTimeline.tsx - Added 'not_run' status support
web/src/lib/sse.ts              - Added resume_start event type and interface
```
