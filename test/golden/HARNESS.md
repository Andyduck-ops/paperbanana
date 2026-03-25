# Golden Data Harness Specification

## Purpose

This document defines how Golden Data cases are stored, validated, and enforced.
Data without a harness is documentation.
Data with a harness is a contract.

## Case Schema

Every Golden Data case is a YAML file under `test/golden/cases/`.
The schema must be followed exactly so validators can consume cases programmatically.

```yaml
# Required fields
id: string                  # e.g. GD-001, GD-UI-003
priority: P0 | P1 | P2     # P0 = must never regress
layer: orchestration | ui | api
title: string
intent: string              # one sentence describing what this case protects

input:
  mode: diagram | plot      # generation mode where applicable
  prompt: string            # user-visible natural language input
  retrieval_mode: auto | manual | random | none
  # add other input fields as needed per layer

expected:
  # orchestration layer
  final_status: completed | failed | paused
  stage_order: list         # canonical stage names in expected order
  must_expose_progress: bool
  must_have_final_artifact: bool
  failed_stage: string      # only for failure cases
  # ui layer
  ui_state: running | completed | failed | resumed
  must_show_stage_progress: bool
  must_show_artifact: bool
  must_show_failure_location: bool
  # api layer
  response_fields: list     # fields that must be present
  must_persist_session: bool

anti_patterns:
  - string                  # describe what a regression looks like

notes:
  - string                  # optional explanation or edge case notes

tags:
  - string                  # e.g. orchestration, resume, retrieval, ui-state
```

## Directory Layout

```text
test/
  golden/
    README.md               <- Golden Data Manual (start here)
    HARNESS.md              <- this file
    MUTATION.md             <- mutation guide for agent-driven expansion
    cases/
      orchestration/
        GD-001.yaml
        GD-002.yaml
        GD-003.yaml
        GD-004.yaml
      ui/
        README.md
        GD-UI-001.yaml
        GD-UI-002.yaml
        GD-UI-003.yaml
        GD-UI-004.yaml
        GD-UI-005.yaml
      api/
        GD-API-001.yaml
    manifests/
      p0.yaml               <- list of all P0 case IDs for CI reference
```

## Validator Interface

The validator is a small program that reads case YAML files and checks that
the implementation satisfies each expected field.

Minimum validator responsibilities:

- Parse all `.yaml` files under `test/golden/cases/`
- For each case, assert expected fields against a target output
- Report which fields passed, which failed, and which were not evaluated
- Exit non-zero if any P0 case fails

Validator entry point (to be implemented):

```text
test/golden/validate.go      <- Go harness for orchestration and API cases
test/golden/validate.ts      <- TypeScript harness for UI cases (Vitest)
```

Output format:

```text
[PASS] GD-001: Happy Path Completion
[FAIL] GD-002: Stage Failure Visibility
  - expected: failed_stage = planner
  - got:      failed_stage = (empty)
[SKIP] GD-UI-003: Artifact Display Contract (not yet wired)
```

## CI Gate

Golden Data gates should block merges when P0 cases regress.

Proposed gate logic:

- On every pull request, run the golden validator against affected layers
- If any P0 case reports FAIL, block the merge
- P1 cases report warnings but do not block
- P2 cases are advisory only

This can be wired into GitHub Actions or any CI runner:

```yaml
# .github/workflows/golden.yml sketch
steps:
  - name: Run golden data validation
    run: go test ./test/golden/... -run TestGoldenP0
```

## Evaluation Mode

For cases that cannot be evaluated deterministically (model-dependent outputs),
use a judge-based evaluation mode:

- Provide the case intent and expected contract to an LLM judge
- The judge returns pass / fail / needs-human
- Any `needs-human` result is escalated to the human reviewer queue
- Do not block CI on judge-based evaluations unless they are P0 and fully deterministic

## When To Add A New Case

Add a new case when:

- a real regression was discovered that had no prior case
- a mutation agent proposes a case that passes the selection rubric
- a product decision changes what the system must do in a named scenario

Do not add a case:

- as a duplicate paraphrase of an existing case
- because it sounds interesting but protects no real behavioral boundary
- when the expected outcome is still genuinely ambiguous

## When To Delete A Case

Delete a case when:

- the feature it protects has been intentionally removed
- it has been superseded by a more precise case
- it has become permanently unevaluable due to architecture changes

Record deletions in a `CHANGELOG.md` under `test/golden/`.
