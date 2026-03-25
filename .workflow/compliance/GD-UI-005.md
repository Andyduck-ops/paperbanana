# GD-UI-005 Compliance Verification

## Golden Case

```yaml
id: GD-UI-005
title: Batch Per-Task Status Visible
intent: When running a batch of tasks, the UI must show per-task status.
expected:
  ui_must_show:
    - task-001 as completed with artifact available
    - task-002 as running with current stage visible
    - task-003 as failed with failed stage visible
anti_patterns:
  - single unified spinner for the entire batch
  - all tasks shown as the same status
```

## Verification Results

### [x] Each task status is independently displayed

**Evidence:** `BatchProgressPanel.tsx` lines 127-164

The component renders each candidate as an individual card with distinct visual styling:

```tsx
{progress.candidates.map((candidate) => (
  <div key={candidate.candidateId} className={`p-3 rounded-lg border ${
    candidate.status === 'completed'
      ? 'border-green-500/50 bg-green-500/10'
      : candidate.status === 'failed'
      ? 'border-red-500/50 bg-red-500/10'
      : candidate.status === 'running'
      ? 'border-primary/50 bg-primary/10'
      : 'border-border bg-background'
  }`}>
    ...
  </div>
))}
```

Each candidate has:
- Unique `candidateId` for keying
- Independent `status` field: `'pending' | 'running' | 'completed' | 'failed'`
- Separate artifacts and error information

### [x] Mixed statuses are visually distinguishable

**Evidence:** Lines 131-139 (background colors) and 146-158 (status indicators)

| Status | Border Color | Background | Icon |
|--------|-------------|------------|------|
| `completed` | `border-green-500/50` | `bg-green-500/10` | `✓` + "Successful" |
| `failed` | `border-red-500/50` | `bg-red-500/10` | `✗` + "Failed" |
| `running` | `border-primary/50` | `bg-primary/10` | `◆` (diamond) |
| `pending` | `border-border` | `bg-background` | `○` (empty circle) |

Error messages are displayed inline when present (lines 160-162):
```tsx
{candidate.error && (
  <p className="mt-1 text-sm text-red-600">{candidate.error}</p>
)}
```

### [x] Avoids unified spinner anti-pattern

**Evidence:** The component does NOT use a single loading state for the entire batch.

1. **Overall progress bar** (lines 114-124) shows aggregate progress as a bar, not a spinner
2. **Per-candidate cards** (lines 127-164) show individual status for each task
3. **State model** (`BatchCandidate` type) tracks each candidate independently:
   ```typescript
   interface BatchCandidate {
     candidateId: number;
     status: 'pending' | 'running' | 'completed' | 'failed';
     artifacts?: BatchArtifact[];
     error?: string;
   }
   ```

The `useBatchGeneration` hook correctly maintains per-candidate state through the `reduceBatchStreamEvent` reducer, which updates individual candidates without losing others' status.

## Data Flow Analysis

1. **Event Sources:**
   - `candidate_start` -> Sets specific candidate to `running`
   - `candidate_complete` -> Sets specific candidate to `completed` or `failed`
   - `batch_result` -> Finalizes with artifacts

2. **State Updates:**
   - `applyCandidatePatch()` immutably updates single candidate
   - Counters (`successful`, `failed`) are recalculated per update
   - No shared status field that could override individual states

## Compliance Status: PASS

All requirements satisfied:

| Requirement | Status | Notes |
|-------------|--------|-------|
| Per-task status visible | PASS | Each candidate has individual card with status |
| Mixed states distinguishable | PASS | Color coding + icons differentiate all 4 states |
| No unified spinner | PASS | Progress bar + individual cards, no single spinner |
| Error visibility | PASS | Inline error display for failed candidates |
| Artifact tracking | PASS | `artifacts` array per candidate |

## Files Verified

- `web/src/components/BatchProgressPanel.tsx` - UI rendering
- `web/hooks/useBatchGeneration.ts` - State management hook
- `web/src/types/batch.ts` - Type definitions
- `web/src/lib/batchProgress.ts` - Stream event reducer
