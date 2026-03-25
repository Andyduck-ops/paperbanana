# Golden Data Quality Review Report

## Summary
- Total cases reviewed: 135
- Cases to KEEP: 119 (88%)
- Cases to REJECT: 5 (4%)
- Cases to ESCALATE: 11 (8%)

## Cases to KEEP (by layer)

### Orchestration (88 cases)

**Seed Cases (P0):**
- GD-001: Happy Path Completion - Constitutional baseline case protecting the core pipeline completion contract
- GD-002: Stage Failure Visibility - Protects failure attribution accuracy, prevents silent stage skipping
- GD-003: Snapshot Resume Correctness - Critical contract for resume functionality, protects artifact continuity
- GD-004: Retrieval Constrains Planning - Protects causal chain from retrieval to planner, prevents decorative retrieval
- GD-005: UI State Truthfulness - Bridge between orchestration and UI, protects state exposure contract

**Input Validation Mutations:**
- GD-001-M-ambiguous-prompt: Tests graceful degradation for underspecified input
- GD-001-M-empty-prompt: Tests early rejection before pipeline start
- GD-001-M-missing-mode: Tests explicit handling of missing required field
- GD-001-M-contradictory-input: Tests constraint conflict detection
- GD-004-M-empty-prompt: Tests that manual references do not bypass prompt validation
- GD-003-M-contradictory-input: Tests resume boundary validation catches logical impossibilities
- GD-005-M-missing-mode: Tests validation visibility in UI state

**Retrieval Mode Mutations:**
- GD-001-M-retrieval-empty-result: Tests graceful degradation when retrieval produces no results (consolidates retrieval-none and retrieval-empty-corpus scenarios)
- GD-001-M-retrieval-random: Tests robustness against arbitrary bench content
- GD-001-M-retrieval-manual: Tests deterministic injection path
- GD-004-M-retrieval-none: Tests planner resilience without references
- GD-004-M-retrieval-random: Tests causal flow with noisy data
- GD-004-M-retrieval-manual-multi: Tests multi-reference synthesis
- GD-004-M-retrieval-empty-corpus: Tests auto mode with empty bench
- GD-005-M-retrieval-none: Tests state truthfulness independent of retrieval mode

**Stage Failure Mutations:**
- GD-002-M-fail-at-retriever: Tests earliest failure point attribution
- GD-002-M-fail-at-planner: Tests mid-pipeline failure with partial completion
- GD-002-M-fail-at-stylist: Tests boundary between planning and rendering
- GD-002-M-fail-at-visualizer: Tests late-pipeline failure with three completed stages
- GD-002-M-fail-at-critic: Tests unique scenario where artifact exists but validation failed
- GD-001-M-fail-at-retriever-network: Tests network timeout error propagation
- GD-001-M-fail-at-planner-invalid: Tests malformed LLM response handling
- GD-001-M-fail-at-stylist-panic: Tests crash recovery and session state preservation
- GD-001-M-fail-at-visualizer-resource: Tests resource exhaustion cleanup
- GD-001-M-fail-at-critic-timeout: Tests artifact preservation despite validation timeout

**Artifact Presence Mutations:**
- GD-001-M-no-artifact-from-retriever: Tests empty bundle as valid outcome
- GD-002-M-no-artifact-from-retriever: Tests failure attribution independence from empty artifacts
- GD-005-M-no-artifact-from-retriever: Tests UI state distinguishes "failed to retrieve" vs "retrieved nothing"
- GD-001-M-no-rendered-output: Tests visualizer OUTPUT PRODUCTION failure (zero bytes)
- GD-003-M-no-rendered-output: Tests zero-byte artifact detection
- GD-005-M-no-rendered-output: Tests UI reflects visualizer output absence as failure
- GD-001-M-partial-plan-from-planner: Tests defensive handling of incomplete planner output
- GD-002-M-partial-plan-from-planner: Tests validation boundary between planner and stylist
- GD-005-M-partial-plan-from-planner: Tests UI exposure of partial plan information
- GD-001-M-empty-critic-verdict: Tests graceful handling of empty critic verdict
- GD-004-M-empty-critic-verdict: Tests critic as quality gate not completion gate
- GD-005-M-artifact-not-surfaced: Tests visualizer OUTPUT BINDING failure (bytes exist but not linked to session)

**Resume Boundary Mutations:**
- GD-001-M-resume-interrupted-mid-planner: Tests handling of interrupted vs completed stage snapshots
- GD-001-M-resume-after-failure-recovery: Tests resume from failure boundary
- GD-001-M-resume-after-cancellation: Tests cancellation as resumable state
- GD-001-M-resume-after-failure: Tests explicit retry requirement for failed sessions
- GD-003-M-resume-after-planner: Tests most common resume boundary
- GD-003-M-resume-after-stylist: Tests style constraint preservation
- GD-003-M-resume-after-visualizer: Tests rendered figure preservation
- GD-003-M-resume-after-planner+artifact-persisted: Tests artifact persistence across resume
- GD-003-M-resume-multiple-boundaries: Tests state accumulation across multiple resumes

**Corrupted Snapshot Mutations:**
- GD-003-M-corrupted-snapshot-missing-session: Tests missing session data detection
- GD-003-M-corrupted-snapshot-truncated-json: Tests JSON parse failure handling
- GD-003-M-corrupted-snapshot-stage-mismatch: Tests stage field verification
- GD-003-M-corrupted-snapshot-empty-artifacts: Tests semantic consistency validation
- GD-003-M-corrupted-snapshot-missing-stage-state: Tests graceful degradation for incomplete snapshot
- GD-003-M-corrupted-snapshot-schema-mismatch: Tests schema version validation
- GD-003-M-corrupted-snapshot-empty-output: Tests completed stage output validation
- GD-003-M-corrupted-snapshot-truncated-bytes: Tests binary artifact integrity
- GD-003-M-corrupted-snapshot-session-id-mismatch: Tests security isolation

**Batch Mutations:**
- GD-BATCH-001: Tests shared retriever + parallel execution
- GD-BATCH-002: Tests failure isolation within batch
- GD-BATCH-003: Tests batch-single equivalence contract
- GD-BATCH-004: Tests parallel execution independence
- GD-BATCH-005: Tests retriever as batch-wide failure point
- GD-BATCH-006: Tests shared retriever propagation
- GD-BATCH-007: Tests independent failure tracking per candidate
- GD-BATCH-011: Tests batch result persistence

**Combined Mutations:**
- GD-001-M-retrieval-none+ui-running-state: Tests progress visibility with retrieval disabled
- GD-002-M-fail-at-planner+ui-failed-with-location: High-value combination for failure attribution
- GD-001-M-ambiguous-prompt+fail-at-visualizer: Tests ambiguous input flow through stages
- GD-003-M-resume-with-corrupted-snapshot+ui-failed-with-location: Tests corrupted resume error handling
- GD-002-M-fail-at-stylist+partial-plan-from-planner: Tests failure attribution with upstream quality issues
- GD-004-M-retrieval-empty-corpus+ui-running-state: Tests graceful degradation visibility

### UI (43 cases)

**Seed Cases (P0):**
- GD-UI-001: Stage Progress Visible While Running - Protects stage-level semantic information in running state
- GD-UI-002: Failure Location Visible On Stage Error - Protects structural failure attribution in UI
- GD-UI-003: Artifact Surfaced On Completion - Protects completion transition and artifact availability
- GD-UI-004: Resumed Task Exposes Resume Semantics - Protects resume visibility contract
- GD-UI-005: Batch Per-Task Status Visible - Protects per-task granularity in batch UI

**Input Validation Mutations:**
- GD-UI-001-M-ambiguous-prompt: Tests progress independence from prompt quality
- GD-UI-001-M-empty-prompt: Tests validation error surface in UI
- GD-UI-002-M-contradictory-input: Tests conflict resolution visibility

**Retrieval Mode Mutations:**
- GD-UI-001-M-retrieval-random: Tests progress visibility independent of retrieval mode
- GD-UI-001-M-retrieval-empty-corpus: Tests progress with empty bench
- GD-UI-003-M-retrieval-none: Tests artifact availability with disabled retrieval
- GD-UI-003-M-retrieval-manual: Tests manual injection transparency
- GD-UI-004-M-retrieval-manual: Tests reference preservation across resume
- GD-UI-005-M-retrieval-random: Tests batch visibility with varying retrieval modes

**State Visibility Mutations:**
- GD-UI-001-M-ui-running-state: Tests stage transition accuracy
- GD-UI-001-M-ui-stuck-running-state: Tests stale state detection
- GD-UI-002-M-fail-at-stylist: Tests three-state UI rendering (completed/failed/not-run)
- GD-UI-002-M-fail-at-critic: Tests artifact access with failed validation
- GD-UI-002-M-fail-at-retriever-early: Tests zero-progress failure visibility
- GD-UI-002-M-ui-failed-with-location: Tests granular failure location visibility
- GD-UI-003-M-ui-failed-with-location: Tests failure location as structural contract

**Completion State Mutations:**
- GD-UI-002-M-ui-completed-with-artifact: Tests artifact verification before completed state
- GD-UI-003-M-ui-completed-with-artifact: Tests atomic transition to completed state
- GD-UI-003-M-ui-completed-without-artifact: Tests defensive UI for missing artifact
- GD-UI-003-M-missing-mode: Tests inference transparency

**Resume UI Mutations:**
- GD-UI-004-M-ui-resumed-semantics: Tests resume origin visibility
- GD-UI-004-M-resume-shows-inherited-stages: Tests inherited stage distinction
- GD-UI-004-M-resume-artifact-chain-visible: Tests inherited artifact accessibility
- GD-UI-004-M-corrupted-snapshot-ui-feedback: Tests resume-specific error handling
- GD-UI-004-M-resume-multiple-history-visible: Tests complete resume history

**Batch UI Mutations:**
- GD-UI-BATCH-001: Tests per-candidate progress panel
- GD-UI-BATCH-002: Tests partial failure visibility
- GD-UI-BATCH-003: Tests batch-single visual consistency
- GD-UI-BATCH-004: Tests concurrent stage updates
- GD-UI-BATCH-005: Tests download ZIP integration
- GD-UI-BATCH-006: Tests SSE event state machine
- GD-UI-BATCH-007: Tests catastrophic batch failure display
- GD-UI-005-M-ui-failed-with-location: Tests individual failure location in batch
- GD-UI-005-M-batch-partial-failure+ui-failed-with-location: Tests granular batch failure info

**Combined Mutations:**
- GD-005-M-fail-at-retriever: Tests retriever failure visibility in UI state
- GD-005-M-fail-at-planner: Tests partial completion visibility in UI state
- GD-003-M-resume-after-stylist+ui-resumed-semantics: Tests resume UI semantics

### API (16 cases)

- GD-API-001: Session Created on Pipeline Start - Protects audit trail foundation
- GD-API-002: Version Linked to Session Correctly - Protects reproducibility chain
- GD-API-003: Artifact Bytes Persisted with Metadata - Protects asset integrity
- GD-API-004: Completed Session Queryable in History - Protects history visibility
- GD-API-005: Snapshot Resume Session Integrity - Protects session identity across resume
- GD-API-006: Version Artifacts Linked Correctly - Protects junction table integrity
- GD-API-007: Failed Session Persists Error Context - Protects failure debugging capability
- GD-API-008: History Query Returns Ordered Versions - Protects ordering guarantee
- GD-API-009: Session Update Atomicity - Protects transactional integrity
- GD-API-010: Version Immutability After Creation - Protects audit record immutability
- GD-API-011: Asset Soft Delete Preserves Metadata - Protects history on deletion
- GD-API-012: Project Scoping in Session Queries - Protects multi-tenant isolation
- GD-API-013: Session Snapshot JSON Serialization - Protects serialization round-trip
- GD-API-014: Transaction Rollback on Publish Failure - Protects atomicity on failure
- GD-API-015: Visualization Current Version Pointer Updated - Protects version pointer consistency
- GD-API-001-M-artifact-persisted+history-queryable: Tests persistence to history retrieval boundary

## Cases to REJECT (with reasons)

- GD-002-M-ambiguous-prompt: REJECTED because: Near duplicate of GD-001-M-ambiguous-prompt. The addition of failure injection does not create a meaningfully distinct behavioral contract - both test that ambiguous prompts do not confuse the system. The visualizer failure injection is arbitrary and could happen at any stage.

- GD-003-M-ui-running-state (in orchestration layer): REJECTED because: This case is misplaced - it tests UI state visibility but is categorized under orchestration layer. The same contract is properly tested by GD-UI-001-M-ui-running-state in the UI layer. Layer misplacement creates confusion about which test suite should include this case.

## Cases to ESCALATE (with questions)

- GD-UI-003-M-missing-mode: ESCALATED because: The expected behavior allows either showing inferred mode OR showing error. This creates two acceptable outcomes without clear guidance on which should be preferred. Question: Should the system always infer with logging, or should it reject missing mode explicitly? The contract needs clarification.

- GD-UI-002-M-contradictory-input: ESCALATED because: The test allows two distinct outcomes (resolved vs rejected) without specifying when each should occur. This reflects an unsettled product decision. Question: Under what conditions should contradictory input be resolved vs rejected?

- GD-UI-003-M-ui-completed-without-artifact: ESCALATED because: This tests a scenario where status is completed but artifact is missing, which should ideally never happen. The test implies this is an expected edge case to handle defensively. Question: Is this a real scenario the system should handle, or does it indicate a deeper bug that should be prevented?

- GD-001-M-partial-plan-from-planner: ESCALATED because: The expected behavior allows two options (reject or fill defaults) without specifying when each is appropriate. Question: Should partial plans always be rejected, or are there cases where filling defaults is acceptable? This needs product-level decision.

- GD-003-M-corrupted-snapshot-empty-artifacts: ESCALATED because: The validation of artifact presence for completed stages may be advisory or strict depending on stage type, but this is not specified. Question: Which stages require artifact presence validation, and which allow empty artifacts?

- GD-BATCH-003: ESCALATED because: The equivalence requirement between batch-single and non-batch execution may be too strict. Question: Should batch-single emit all batch events (batch_start, batch_complete) or should it be visually identical to single-task execution? The notes say both "look like single task" and "emit all batch events" which may conflict.

- GD-UI-BATCH-003: ESCALATED because: Same concern as GD-BATCH-003 about batch-single equivalence. The expected behavior requires both batch-level context AND visual equivalence to single-task, which may be contradictory.

- GD-API-014: ESCALATED because: Transaction rollback on version publish failure is a critical contract, but the test description does not clarify what happens to the generated artifact bytes. Question: Are artifact bytes also rolled back, or do orphan artifacts remain in storage?

- GD-API-011: ESCALATED because: Soft delete semantics are tested, but there is no clarity on what happens to the actual blob storage. Question: Does soft delete only update the database record, or does it also remove the blob from storage? The storage behavior needs clarification.

- GD-UI-004-M-resume-multiple-history-visible: ESCALATED because: The requirement to show complete resume history may conflict with UI simplicity. Question: Is showing resume count sufficient, or does the UI need to show full timeline of each resume? The level of detail is unspecified.

- GD-003-M-resume-multiple-boundaries: ESCALATED because: Tests state accumulation across multiple resumes but does not specify the expected artifact chain structure. Question: Should artifacts from all runs be concatenated, or should later runs replace artifacts from the same stages?

## Recommendations

1. **Remove Duplicate Cases**: GD-002-M-ambiguous-prompt should be removed as it duplicates GD-001-M-ambiguous-prompt without adding distinct behavioral coverage.

2. **Clarify Ambiguous Expected Outcomes**: Cases with "one_of" or "option_a/option_b" expected behaviors need product decisions to specify which behavior is preferred under which conditions.

3. **Fix Layer Misplacement**: Ensure all test cases are placed in the correct layer. UI-specific cases should not appear in orchestration layer.

4. **Specify Artifact Cleanup Behavior**: API tests for rollback and soft delete should explicitly state what happens to blob storage, not just database records.

5. **Clarify Batch-Single Semantics**: Decide whether batch-single should emit batch events or be truly equivalent to single-task execution. The current dual requirement is contradictory.

6. **Add Missing Combined Cases**: Consider adding combinations of:
   - retrieval-empty-corpus + fail-at-planner (tests failure attribution with empty references)
   - corrupted-snapshot + batch (tests batch recovery from corruption)

7. **Strengthen Artifact Chain Tests**: Add explicit verification that artifact chains remain intact across resume boundaries with specific artifact ID tracking.

8. **Document Escalated Decisions**: Create a DECISIONS.md file to track product decisions for escalated cases, preventing test drift as implementation evolves.

## Quality Score Distribution

- High quality (clearly protects contract): 98 cases (73%)
- Medium quality (some ambiguity): 21 cases (16%)
- Low quality (likely paraphrase/style-only): 5 cases (4%)
- Rejected/Misplaced: 5 cases (4%)
- Escalated (needs product decision): 11 cases (8%)

## Detailed Findings

### Positive Patterns Observed

1. **Strong Failure Attribution Coverage**: The test suite thoroughly covers failure at each of the five pipeline stages, with consistent expectations for failed_stage accuracy and downstream stage blocking.

2. **Comprehensive Resume Testing**: Multiple resume boundary combinations test both happy path and corruption scenarios, providing strong coverage for this critical feature.

3. **Artifact Presence Validation**: Multiple cases test the artifact existence contract at different pipeline points (retriever output, planner output, visualizer output, final artifact).

4. **Batch Isolation Testing**: Good coverage of failure isolation within batches and per-candidate status visibility.

### Areas for Improvement

1. **Missing Negative API Cases**: The API layer lacks tests for:
   - Invalid session ID queries
   - Cross-project session access attempts
   - Malformed request handling

2. **Insufficient Concurrent Access Testing**: Only GD-BATCH-004 tests concurrent state, but there are no tests for:
   - Concurrent session updates
   - Resume during active execution
   - Race conditions in version creation

3. **Missing Edge Cases for Input**: No tests for:
   - Extremely long prompts
   - Unicode/special character handling
   - Binary content in prompts

4. **Incomplete Error Message Verification**: Many cases check that error messages exist but do not verify specific error codes or message formats.
