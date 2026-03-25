# Golden Data Tests Report

## Summary

Generated 5 Vitest test files based on UI golden data cases from `test/golden/cases/ui/`.

## Generated Test Files

| File | Golden Case | Priority | Test Count |
|------|-------------|----------|------------|
| `web/src/test/golden/GD-UI-001.test.tsx` | Stage Progress Visible While Running | P0 | 8 tests |
| `web/src/test/golden/GD-UI-002.test.tsx` | Failure Location Visible On Stage Error | P0 | 10 tests |
| `web/src/test/golden/GD-UI-003.test.tsx` | Artifact Surfaced On Completion | P0 | 10 tests |
| `web/src/test/golden/GD-UI-004.test.tsx` | Resumed Task Exposes Resume Semantics | P0 | 9 tests |
| `web/src/test/golden/GD-UI-005.test.tsx` | Batch Per-Task Status Visible | P1 | 13 tests |

**Total: 50 tests**

## Test Coverage by Golden Case

### GD-UI-001: Stage Progress Visible While Running (P0)

**Intent:** While a generation task is running, the UI must display real stage-level progress and not show only a generic loading indicator.

**Tests verify:**
- Current active stage name is displayed
- Stages already completed are shown
- Overall pipeline structure is visible (not just a spinner)
- Completed badge is NOT shown (task is still running)
- Final artifact is NOT shown
- Failure message is NOT shown
- Completed stages are distinguished from pending stages
- No generic spinner without stage info

**Components tested:** `ProgressPanel`, `StageCard`

---

### GD-UI-002: Failure Location Visible On Stage Error (P0)

**Intent:** When a task fails, the UI must display where in the pipeline the failure occurred and must not show only a generic error message.

**Tests verify:**
- Failed status indicator is shown
- Name of the failed stage is displayed
- Stages that completed before failure are shown
- Clear distinction between completed stages and stages that did not run
- Completed badge is NOT shown
- Final artifact is NOT shown
- Running spinner is NOT shown
- Error is NOT generic (has stage attribution)
- NOT all stages shown as failed
- Failed stage is distinct from pending stage

**Components tested:** `ProgressPanel`, `StageCard`

---

### GD-UI-003: Artifact Surfaced On Completion (P0)

**Intent:** When a task completes, the UI must surface the final artifact in a usable form and must transition cleanly out of the running state.

**Tests verify:**
- Completed status indicator is shown
- All stages shown as completed
- Running spinner is NOT shown
- Failure message is NOT shown
- Final artifact is displayed
- Artifact area is NOT empty
- Completed state is stable (non-transient)

**Components tested:** `ProgressPanel`, `ResultPanel`, `ArtifactPreview`

---

### GD-UI-004: Resumed Task Exposes Resume Semantics (P0)

**Intent:** When a task was resumed from a snapshot, the UI must not present it as a fresh execution. Prior completed stages and resume origin must be distinguishable.

**Tests verify:**
- Stages completed before resume are shown as already done
- Current stage is stylist (resume point), NOT retriever (start)
- Some indication that this is a resumed run
- Retriever/planner NOT shown as currently running
- Pipeline NOT presented as starting from scratch
- Resume metadata NOT discarded
- User can see which stages were inherited from snapshot
- User can see which stages are being run in this resume

**Components tested:** `ProgressPanel` (with extended `ResumableStageState`)

**Note:** This test introduces an extended interface `ResumableStageState` that adds `wasCompletedBeforeResume` to track resume semantics. The actual UI implementation should adopt this or a similar pattern.

---

### GD-UI-005: Batch Per-Task Status Visible (P1)

**Intent:** When running a batch of generation tasks, the UI must show per-task status so the user can track individual task progress without the batch collapsing into a single opaque indicator.

**Tests verify:**
- Task-001 shown as completed with artifact
- Task-002 shown as running with current stage
- Task-003 shown as failed with failed stage
- Batch-level summary showing mixed states
- NO single unified spinner for entire batch
- NOT all tasks shown as same status
- Failed task NOT hidden behind generic batch error
- User can distinguish each task's individual status
- User can act on completed tasks while others running
- Batch summary reflects actual task state distribution
- Correct status icons for each state (completed/running/failed)

**Components tested:** `BatchProgressPanel`

---

## Testing Methodology

### State-Based Assertions

All tests use state-based assertions, NOT pixel diffing:

```typescript
// State-based assertion (correct approach)
expect(screen.getByText('Planner')).toBeInTheDocument();
expect(screen.getByText('✓')).toBeInTheDocument();

// NOT pixel diffing (anti-pattern avoided)
// expect(component).toMatchSnapshot();
```

### Anti-Pattern Checks

Each test file includes explicit anti-pattern checks from the golden cases:

- GD-UI-001: No generic spinner without stage info
- GD-UI-002: No generic error message without stage attribution
- GD-UI-003: No completed badge with empty artifact area
- GD-UI-004: No resume metadata discarded
- GD-UI-005: No single unified spinner for entire batch

### Interaction Contracts

Tests verify interaction contracts defined in golden cases:

- User can identify which agent is currently working
- User has enough information to decide whether to retry
- User can view or download the final artifact
- User can see which stages were inherited from snapshot
- User can distinguish each task's individual status

## Running Tests

```bash
# Run all golden tests
cd web
npm test -- --run src/test/golden

# Run specific golden test
npm test -- --run src/test/golden/GD-UI-001.test.tsx

# Run with coverage
npm test -- --run --coverage src/test/golden
```

## Test Dependencies

All tests mock the `useLanguage` hook to avoid i18n dependencies:

```typescript
vi.mock('../../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
}));
```

## Notes

1. **GD-UI-004** introduces an extended interface for resume semantics. The actual `ProgressPanel` component may need to be updated to accept this metadata.

2. **GD-UI-005** tests the `BatchProgressPanel` component which already supports per-task status. Tests verify the existing implementation matches the golden case requirements.

3. All tests follow the existing patterns from:
   - `web/src/hooks/useGenerate.test.ts`
   - `web/src/components/ProgressPanel.test.tsx`

## File Locations

- Golden data cases: `test/golden/cases/ui/GD-UI-*.yaml`
- Generated tests: `web/src/test/golden/GD-UI-*.test.tsx`
- This report: `.workflow/golden-tests-report.md`
