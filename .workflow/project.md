# PaperBanana Project

## Core Value

PaperBanana is a full-stack scientific visualization workspace that transforms text descriptions into diagrams and plots via a 5-stage multi-agent LLM pipeline.

**Core Promise**: Users input a text prompt, get a professional scientific figure.

---

## Vision

Build a reliable, user-trusting scientific figure generation system where:
- Users always know what the system is doing
- Failures are visible and actionable
- Resumed tasks behave correctly
- Batch operations show per-task status

---

## Requirements

### Validated (P0 - Constitutional)
From Golden Data compliance verification:

| Case | Title | Status | Gap |
|------|-------|--------|-----|
| GD-001 | Happy Path Completion | PARTIAL | stylist stage can be silently skipped |
| GD-002 | Stage Failure Visibility | PASS | - |
| GD-003 | Snapshot Resume Correctness | PASS | - |
| GD-004 | Retrieval Constrains Planning | PASS | - |
| GD-005 | UI State Truthfulness | PASS | - |
| GD-UI-001 | Stage Progress Visible | PASS | - |
| GD-UI-002 | Failure Location Visible | PASS | SSE event type fixed |
| GD-UI-003 | Artifact Surfaced On Completion | PASS | Fixed in session |
| GD-UI-004 | Resumed Task Exposes Resume Semantics | MISSING | Backend missing resume_start event |
| GD-UI-005 | Batch Per-Task Status Visible | PASS | - |

### Active Development Areas

1. **GD-UI-004 Fix**: Backend needs `EventResumeStarted` event type
2. **GD-001 Issue**: stylist stage optional - add warning or validation
3. **Security Gaps**: No authentication, no rate limiting
4. **Production Readiness**: No graceful shutdown, WAL disabled

### Out of Scope (for now)
- Pixel-perfect UI layout
- Model-dependent prose outputs
- External benchmark quality scores

---

## Key Decisions

1. **Golden Data is Constitutional**: All 347 cases define behavioral contracts
2. **Priority Order**: P0 → P1 → P2, fix blockers first
3. **Multi-Agent Architecture**: 5 stages (Retriever → Planner → Stylist → Visualizer → Critic)
4. **State Management**: Session snapshots for resume capability
5. **Streaming**: SSE for real-time progress updates

---

## Success Criteria

### Phase Completion Criteria
- All P0 Golden Data cases pass
- No regressions on existing P0 cases
- Code coverage ≥ 80% on touched modules
- Multi-agent review approval

### Final Delivery Criteria
- 347/347 Golden Data cases implemented
- All compliance reports show PASS
- Production-ready (auth, rate limiting, monitoring)

---

## Constraints

1. **Backward Compatibility**: Never break existing functionality
2. **Golden Data Compliance**: All changes must pass Golden Data tests
3. **Review Required**: Multi-round agent review before delivery
4. **Parallel Execution**: Allow concurrent development where possible

---

## Terminology

| Term | Definition |
|------|------------|
| Golden Data | Behavioral contract test cases |
| P0 | Constitutional - must never regress |
| P1 | High priority - should be fixed soon |
| P2 | Medium priority - nice to have |
| Pipeline | 5-stage agent execution flow |
| Session | Single generation task lifecycle |
| Snapshot | Checkpoint for resume capability |
