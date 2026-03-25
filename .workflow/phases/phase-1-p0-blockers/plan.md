# Phase 1: P0 Blockers Fix - Execution Plan

## Overview
Fix 2 P0-blocking issues identified in Golden Data compliance review.

## Task Decomposition

### TASK-001: Add EventResumeStarted event type
**Type**: Backend
**Files**: `internal/domain/agent/events.go`
**Implementation**:
```go
// Add to EventType constants
EventResumeStarted EventType = "resume_start"
```
**Verification**: `grep -n "resume_start" internal/domain/agent/events.go`
**Convergence Criteria**: EventResumeStarted constant exists

### TASK-002: Emit resume_start event in runner
**Type**: Backend
**Files**: `internal/application/orchestrator/runner.go`
**Implementation**:
- In `newRestoredSessionTracker()` or immediately after
- Emit `EventResumeStarted` with:
  - `resumed_from_stage`: snapshot.Stage.Stage
  - `stages_completed_before_resume`: completed stage names
  - `session_id`: snapshot.Session.SessionID
**Verification**: `grep -n "resume_start" internal/application/orchestrator/runner.go`
**Convergence Criteria**: Event emission code exists

### TASK-003: Frontend handle resume_start event
**Type**: Frontend
**Files**: `web/src/lib/sse.ts`
**Implementation**:
```typescript
case 'resume_start':
  options.onResumeStart?.(eventData as ResumeStartEvent);
  break;
```
**Verification**: `grep -n "resume_start" web/src/lib/sse.ts`
**Convergence Criteria**: Switch case handles resume_start

### TASK-004: Display resume indicator in UI
**Type**: Frontend
**Files**:
- `web/src/hooks/useGenerate.ts`
- `web/src/components/ProgressPanel.tsx`
**Implementation**:
1. Add `resumeMetadata` to state
2. Handle `onResumeStart` to populate metadata
3. Display badge when `resumeMetadata` is set
**Verification**:
- `grep -n "resumeMetadata" web/src/hooks/useGenerate.ts`
- `grep -n "ResumeIndicator" web/src/components/ProgressPanel.tsx`
**Convergence Criteria**: Resume badge visible in UI

### TASK-005: Add stylist stage validation
**Type**: Backend
**Files**: `internal/application/orchestrator/runner.go`
**Implementation**:
- Option A: Make stylist required (return error if nil)
- Option B: Emit warning event when stylist is nil
**Verification**: `grep -n "stylist.*nil" internal/application/orchestrator/runner.go`
**Convergence Criteria**: Validation or warning exists

### TASK-006: SSE integration tests
**Type**: Test
**Files**: `web/src/test/integration/sse-flow.test.ts`
**Implementation**:
- Test full SSE event flow
- Verify all event types match backend
- Test reconnection scenarios
**Verification**: Test file exists and passes
**Convergence Criteria**: `npm test -- --run sse-flow.test.ts` passes

---

## Execution Wave Plan

### Wave 1 (Sequential - Core Backend)
```
TASK-001 ──→ TASK-002 ──→ TASK-005
```
These tasks modify core backend and must be sequential.

### Wave 2 (Parallel - Frontend)
```
TASK-003 ──┐
           ├──→ TASK-004
TASK-002 ──┘
```
Frontend can start after TASK-002 provides the event.

### Wave 3 (Validation)
```
TASK-004 ──→ TASK-006
```
Integration tests after UI is complete.

---

## Quality Gates

### Per-Task Gates
- [ ] Code compiles without errors
- [ ] Unit tests pass
- [ ] Golden Data case passes

### Phase Gates
- [ ] All 6 tasks completed
- [ ] GD-UI-004 test passes
- [ ] GD-001 test passes
- [ ] No regressions on other P0 cases

---

## Parallel Execution Strategy

**Sequential Core** (TASK-001, TASK-002, TASK-005):
- Single agent, ordered execution
- Core backend changes

**Parallel Frontend** (TASK-003, TASK-004):
- Can start after TASK-002
- Independent from TASK-005

**Final Validation** (TASK-006):
- Runs after all implementation
- Full SSE flow verification

---

## Review Requirements

- [ ] Code review by senior developer
- [ ] Golden Data compliance check
- [ ] Integration test review
- [ ] Documentation update

---

*Generated: 2026-03-23*
