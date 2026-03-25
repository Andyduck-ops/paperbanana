# Phase 2: Core Pipeline Hardening - Plan

**Created:** 2026-03-24
**Status:** Ready for execution

## Goals

1. Standardize error handling across all agents
2. Ensure session state consistency with atomic updates
3. Implement configurable timeouts and proper cancellation propagation

## Tasks

### Task 2.1: Error Handling Standardization

**Priority:** P0
**Files:**
- `internal/domain/agent/errors.go` (new) - Define standard error types and codes
- `internal/domain/agent/events.go` - Add error classification
- `internal/application/orchestrator/runner.go` - Use standardized errors

**Implementation:**
1. Create `errors.go` with:
   - Error codes enum (ErrCodeLLMTimeout, ErrCodeRateLimit, ErrCodeInvalidInput, etc.)
   - `NewErrorDetail(code, message, retryable)` constructor
   - `WrapAgentError(err, stage, code)` helper

2. Update `ErrorDetail` to support error classification:
   - Add `Category` field (transient, permanent, configuration)
   - Add `Suggestion` field for user-facing guidance

3. Ensure all stages emit proper failure events with classified errors

**Acceptance Criteria:**
- All agent errors use standardized codes
- Error events include actionable suggestions
- Tests verify error classification

### Task 2.2: Session State Consistency

**Priority:** P0
**Files:**
- `internal/application/orchestrator/session.go` - Atomic state updates
- `internal/infrastructure/agentstate/snapshot.go` - Add integrity checks

**Implementation:**
1. Add checksum calculation for snapshots:
   - Compute SHA256 of state before save
   - Verify checksum on restore

2. Ensure atomic state transitions:
   - Use mutex for sessionTracker state mutations
   - Add state transition validation (pending→running→completed/failed)

3. Add cross-tab state sync support:
   - Add `SessionVersion` field for optimistic locking
   - Emit `session_version_mismatch` event if conflict detected

**Acceptance Criteria:**
- Snapshot integrity verified on restore
- State transitions are atomic and validated
- Tests cover concurrent access scenarios

### Task 2.3: Timeout & Cancellation

**Priority:** P0
**Files:**
- `internal/config/config.go` - Add timeout configuration
- `internal/application/orchestrator/runner.go` - Stage timeout handling
- `internal/domain/agent/events.go` - Add timeout event metadata

**Implementation:**
1. Add configurable stage timeouts:
   - Add `StageTimeouts` to config (map[StageName]time.Duration)
   - Default: retriever=30s, planner=60s, stylist=30s, visualizer=120s, critic=60s

2. Wrap each stage execution with timeout context:
   ```go
   stageCtx, cancel := context.WithTimeout(ctx, timeout)
   defer cancel()
   output, err := stageAgent.Execute(stageCtx, stageInput)
   ```

3. Add timeout information to error messages:
   - Include timeout duration in ErrorDetail
   - Emit `stage_timeout` event with timing info

4. Ensure proper cancellation propagation:
   - All goroutines respect context.Done()
   - Cleanup runs even on timeout/cancel

**Acceptance Criteria:**
- Each stage has configurable timeout
- Timeout errors include duration and stage info
- Tests verify timeout behavior

## Verification

- All existing tests pass
- New tests cover:
  - Error classification
  - Snapshot integrity verification
  - Concurrent state access
  - Stage timeout handling
  - Cancellation propagation
- Golden Data cases still pass

## Files Changed

| File | Change |
|------|--------|
| `internal/domain/agent/errors.go` | New - Standardized error types |
| `internal/domain/agent/types.go` | Modify - Add Category, Suggestion to ErrorDetail |
| `internal/application/orchestrator/session.go` | Modify - Add mutex, state validation |
| `internal/application/orchestrator/runner.go` | Modify - Timeout handling, error wrapping |
| `internal/infrastructure/agentstate/snapshot.go` | Modify - Checksum verification |
| `internal/config/config.go` | Modify - Stage timeout config |
