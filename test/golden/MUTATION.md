# Golden Data Mutation Guide

## Purpose

This document tells agent workers exactly how to generate new Golden Data candidate cases.
It is the instruction manual for large-scale case expansion.

When you receive this document, your job is:

1. Read the existing P0 seed cases in `test/golden/cases/`
2. Apply the mutation operators defined below
3. Produce new candidate YAML files following the schema in `HARNESS.md`
4. Mark every generated case with `source: mutation` and the operator name used
5. Do NOT decide what passes the rubric yourself — submit all candidates for human review

## Seed Cases

The five P0 seed cases are:

| ID | Title | Layer |
|----|-------|-------|
| GD-001 | Happy Path Completion | orchestration |
| GD-002 | Stage Failure Visibility | orchestration |
| GD-003 | Snapshot Resume Correctness | orchestration |
| GD-004 | Retrieval Constrains Planning | orchestration |
| GD-005 | UI State Truthfulness | ui |

These are the constitutional baseline. Mutations should extend them, not replace them.

## Mutation Axes

Each axis describes a dimension along which a seed case can be varied.
Apply one or more axes to a seed to produce a candidate.

### Axis 1: Input Ambiguity

Vary the quality and completeness of the user input.

Operators:
- `ambiguous-prompt`: replace the input prompt with one that is underspecified or vague
- `missing-mode`: omit the generation mode and require the system to infer or reject
- `contradictory-input`: provide conflicting constraints in the same prompt
- `empty-prompt`: submit a blank or near-blank prompt

Example: take GD-001 and apply `ambiguous-prompt`

```yaml
id: GD-001-M-ambiguous-prompt
source: mutation
operator: ambiguous-prompt
parent: GD-001
intent: System should handle an underspecified prompt gracefully without silent failure
input:
  prompt: "Make something about science."
  mode: diagram
  retrieval_mode: auto
expected:
  # system should either ask for clarification or produce a best-effort result with observable state
  final_status: completed | failed
  must_expose_progress: true
anti_patterns:
  - silent success with empty output
  - crash without error state
```

### Axis 2: Retrieval Mode Variation

Vary how references are sourced.

Operators:
- `retrieval-none`: set retrieval_mode to none, verify pipeline still runs
- `retrieval-random`: set retrieval_mode to random, verify planner still receives something
- `retrieval-manual`: inject specific references manually, verify planner uses them
- `retrieval-empty-corpus`: set retrieval_mode to auto with an empty bench, verify graceful degradation

Example: take GD-004 and apply `retrieval-none`

```yaml
id: GD-004-M-retrieval-none
source: mutation
operator: retrieval-none
parent: GD-004
intent: When retrieval is disabled, pipeline must still complete without retrieval artifacts
input:
  prompt: "Diagram showing a transformer attention mechanism."
  mode: diagram
  retrieval_mode: none
expected:
  final_status: completed
  stage_order:
    - retriever
    - planner
    - stylist
    - visualizer
    - critic
  must_have_final_artifact: true
anti_patterns:
  - pipeline crashes because retriever returns nothing
  - planner blocks waiting for references that will never arrive
```

### Axis 3: Failure Location

Inject failure at each stage boundary.

Operators:
- `fail-at-retriever`: retriever returns error
- `fail-at-planner`: planner returns error
- `fail-at-stylist`: stylist returns error
- `fail-at-visualizer`: visualizer returns error
- `fail-at-critic`: critic returns error

Example: take GD-002 and apply `fail-at-visualizer`

```yaml
id: GD-002-M-fail-at-visualizer
source: mutation
operator: fail-at-visualizer
parent: GD-002
intent: Failure at visualizer stage must be attributed correctly and not advance to critic
input:
  prompt: "Plot a time series of population growth."
  mode: plot
  retrieval_mode: auto
  inject_failure_at: visualizer
expected:
  final_status: failed
  failed_stage: visualizer
  must_expose_progress: true
anti_patterns:
  - critic runs after visualizer failure
  - failed_stage reported as empty or generic
  - UI shows no indication of where failure occurred
```

### Axis 4: Resume Boundary

Vary where the snapshot was taken and where resume begins.

Operators:
- `resume-after-planner`: snapshot captured after planner, resume from stylist
- `resume-after-stylist`: snapshot captured after stylist, resume from visualizer
- `resume-after-visualizer`: snapshot captured after visualizer, resume from critic
- `resume-with-corrupted-snapshot`: snapshot is incomplete or malformed

Example: take GD-003 and apply `resume-after-stylist`

```yaml
id: GD-003-M-resume-after-stylist
source: mutation
operator: resume-after-stylist
parent: GD-003
intent: Resume from stylist checkpoint must not rerun retriever or planner
input:
  snapshot_stage: stylist
  resume_from: visualizer
expected:
  stages_rerun: []
  stages_continued:
    - visualizer
    - critic
  final_status: completed
  must_have_final_artifact: true
anti_patterns:
  - full pipeline rerun after resume
  - visualizer receives empty planner output due to lost state
```

### Axis 5: Artifact Presence

Vary whether artifacts are present, partial, or absent at stage boundaries.

Operators:
- `no-artifact-from-retriever`: retriever completes but returns no bundle
- `partial-plan-from-planner`: planner returns a plan with missing fields
- `no-rendered-output`: visualizer completes but produces no image bytes
- `empty-critic-verdict`: critic returns verdict with no structured content

### Axis 6: UI State Surface

Vary what the UI must display under each system condition.

Operators:
- `ui-running-state`: system is mid-execution, verify progress visibility
- `ui-completed-with-artifact`: system is done, verify artifact is surfaced
- `ui-failed-with-location`: system failed, verify failure location is visible
- `ui-resumed-semantics`: system resumed, verify resumed badge or indicator
- `ui-batch-progress`: batch of tasks running, verify per-task status visibility

### Axis 7: Batch vs Single

Vary whether the task is single or part of a batch.

Operators:
- `batch-all-succeed`: all batch tasks complete
- `batch-partial-failure`: some batch tasks fail, others succeed
- `batch-single-item`: batch of one item behaves same as single
- `batch-concurrent-states`: tasks reach different stages simultaneously

### Axis 8: Persistence Boundary

Vary whether persistence is exercised across restore points.

Operators:
- `session-created-on-start`: verify session record exists after run begins
- `version-linked-to-session`: verify version record links to session correctly
- `artifact-persisted`: verify artifact bytes and metadata survive persistence
- `history-queryable`: verify completed sessions appear in history query

## Combination Rules

You may combine up to two axes per candidate case.
Combining three or more axes in a single case produces scenarios that are hard to evaluate and maintain.

Good combination examples:
- `retrieval-none` + `ui-running-state`: verify UI still shows progress when retrieval is off
- `fail-at-planner` + `ui-failed-with-location`: verify failure at planner surfaces correctly in UI
- `resume-after-planner` + `artifact-persisted`: verify artifacts survive a resume boundary

Bad combination examples:
- `empty-prompt` + `fail-at-visualizer` + `resume-after-stylist`: too many moving parts in one case

## Output Format For Agent Workers

For each candidate you generate, produce:

1. A YAML file following the schema in `HARNESS.md`
2. A one-line intent statement that would pass the selection rubric
3. At least two anti-patterns
4. A `source: mutation` field and the operator name

Submit all candidates regardless of your own confidence.
Do not self-filter aggressively — the human reviewer and rubric will do that.

## Selection Rubric (Reminder)

The human reviewer will apply this rubric after you submit:

**Keep if:**
- protects a real behavioral contract
- is distinguishable from all existing cases
- would catch a real regression
- stays valid across refactors

**Reject if:**
- is a paraphrase of an existing case
- freezes unstable prompt wording without protecting behavior
- requires full external benchmark corpus to evaluate
- protects only subjective style

**Escalate if:**
- mixes too many contracts in one scenario
- reflects unsettled product decisions
- has ambiguous expected outcome

## Naming Convention

```text
{parent-id}-M-{operator-name}.yaml

Examples:
GD-001-M-ambiguous-prompt.yaml
GD-002-M-fail-at-visualizer.yaml
GD-003-M-resume-after-stylist.yaml
GD-UI-001-M-ui-failed-with-location.yaml
```

For cases that combine two operators:

```text
GD-002-M-fail-at-planner+ui-failed-with-location.yaml
```
