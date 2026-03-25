# PaperBanana Golden Data Manual

## Purpose

This manual defines the first-pass Golden Data strategy for PaperBanana.
Its job is not to judge artistic beauty or model eloquence.
Its job is to lock the system's behavioral contract before large-scale UI, backend, and multi-agent iteration begins.

In this project, Golden Data is the executable form of product intent.
Code may be refactored, agents may be swapped, and prompts may mutate.
The behavioral contract should remain stable.

## Phase 01.1 Baseline

The current repo baseline is formalized in:

- `.planning/phases/01.1-core-rules-and-golden-baseline/BASELINE-SPEC.md`
- `test/golden/manifests/core-baseline.yaml`
- `test/golden/TRACEABILITY.md`

These files define the canonical supported journeys and map them onto the existing `GD-*` case inventory.

Do not start large-scale repair work without checking whether the target behavior already has:

- a `PB-*` baseline ID
- an existing `GD-*` golden case
- a stated coverage status of `covered`, `partial`, or `missing`

## Project Topology

PaperBanana is a full-stack scientific visualization workspace driven by a multi-agent pipeline:

1. `retriever`
2. `planner`
3. `stylist`
4. `visualizer`
5. `critic`

The product is not a generic chat app.
It is a stateful generation system with:

- agent orchestration
- stage-by-stage event streaming
- session persistence and resume
- workspace and version history
- UI visibility into a long-running generation task

Because of that, Golden Data must focus on system behavior, not only snapshots of text.

## What Golden Data Should Protect

Golden Data in PaperBanana should primarily protect five things:

1. Pipeline correctness
2. Failure visibility
3. Resume correctness
4. Retrieval-to-planning constraint flow
5. UI state semantics

If these five drift silently, the system can appear to work while actually losing its core contract.

## What Golden Data Should Not Protect First

Do not start by freezing these:

- pixel-perfect UI layout
- incidental wording changes
- fully model-dependent prose outputs
- broad benchmark quality scores that require external corpora
- style details that are subjective and expected to evolve quickly

These may matter later, but they are not the first constitutional layer.

## Golden Data Layers

### Layer 1: Orchestration Golden Data

This is the highest-value layer.
It protects the multi-agent execution contract.

Typical assertions:

- stages run in the canonical order
- stage transitions are visible to the system
- intermediate artifacts are passed forward correctly
- final status is truthful
- failure is attached to the correct stage
- resume continues from the correct boundary

### Layer 2: Product Interaction Golden Data

This layer protects user-visible state semantics.
It is not about visual polish.
It is about whether the UI tells the truth.

Typical assertions:

- running tasks show stage progress
- failed tasks show where and why they failed
- resumed tasks expose resumed state instead of pretending to be fresh
- completed tasks surface final artifacts and final status consistently

### Layer 3: API and Persistence Golden Data

This layer protects the backend contract between orchestration and product surfaces.

Typical assertions:

- session records are created and updated
- version and artifact persistence follow the expected shape
- API responses expose stable state fields needed by the UI
- restore metadata survives persistence boundaries

## P0 Golden Cases

These are the first five cases that should exist before massive agent-driven mutation begins.

### GD-001 Happy Path Completion

**Goal**: A valid generation request completes through the canonical pipeline.

**Why it matters**:
This is the baseline constitutional case. If this drifts, everything else becomes noisy.

**Given**:

- a valid generation prompt
- a valid mode such as `diagram` or `plot`
- normal pipeline execution

**Expect**:

- stages execute in canonical order
- final session status is `completed`
- final output exists
- stage progress is visible to downstream consumers
- final artifacts are present and typed coherently

**Anti-patterns**:

- skipped stage with fake success
- final success without final artifact
- hidden intermediate failure

### GD-002 Stage Failure Visibility

**Goal**: If one stage fails, the system reports a truthful failure boundary.

**Why it matters**:
A multi-agent system becomes ungovernable if failure location is opaque.

**Given**:

- a task that fails at one defined stage

**Expect**:

- session status is failed
- failed stage is identifiable
- downstream stages do not pretend to complete
- UI-consumable state can explain where the run stopped

**Anti-patterns**:

- generic error with no stage attribution
- pipeline keeps advancing after a failed critical stage
- UI says only "generation failed" with no structure

### GD-003 Snapshot Resume Correctness

**Goal**: A resumed run continues from the right boundary with intact state.

**Why it matters**:
PaperBanana already treats snapshot/resume as a core system ability.
This should be constitutional, not incidental.

**Given**:

- an interrupted session with a stored snapshot

**Expect**:

- prior successful stages are restored, not rerun blindly
- resume metadata is preserved
- the run continues from the correct next stage
- final outcome is consistent with resumed execution

**Anti-patterns**:

- full rerun disguised as resume
- lost metadata across restore
- duplicate artifact chains caused by bad resume boundaries

### GD-004 Retrieval Constrains Planning

**Goal**: Retriever output must have observable influence on planner input or output.

**Why it matters**:
Otherwise the pipeline becomes decorative instead of causal.

**Given**:

- a request with retrieved references or examples

**Expect**:

- planner receives the relevant retrieved material
- generated plan reflects the presence of retrieval context
- artifact flow preserves reference bundles into later stages

**Anti-patterns**:

- retriever runs but planner ignores it
- references are generated but dropped from artifact flow
- retrieval mode changes but planning behavior is unaffected

### GD-005 UI State Truthfulness

**Goal**: The UI contract must reflect real execution state.

**Why it matters**:
A long-running generation workspace lives or dies on user trust.
The UI must not hallucinate progress.

**Given**:

- running, failed, completed, and resumed tasks

**Expect**:

- stage progress is visible while running
- completion exposes final output and status
- failure exposes location and reason class
- resumed tasks expose resumed semantics where relevant

**Anti-patterns**:

- spinner without stage meaning
- completed badge without artifact availability
- failed state with no structured reason
- resumed task presented as brand new execution

## Recommended Golden Data Schema

Each case should be stored as structured data plus human explanation.
Use a schema like this:

```yaml
id: GD-001
priority: P0
layer: orchestration
title: Happy Path Completion
intent: A valid request completes through the canonical pipeline
input:
  mode: diagram
  prompt: Create a diagram explaining a scientific agent workflow.
  retrieval_mode: auto
expected:
  final_status: completed
  stage_order:
    - retriever
    - planner
    - stylist
    - visualizer
    - critic
  must_expose_progress: true
  must_have_final_artifact: true
anti_patterns:
  - skipped stage with fake success
  - completed without final artifact
notes:
  - This is the constitutional baseline case.
```

## Mutation Strategy For Multi-Agent Expansion

When you later dispatch many agents to generate new candidate cases, mutate around the contract instead of mutating randomly.

Recommended mutation axes:

1. Input ambiguity
2. Retrieval mode
3. Stage failure location
4. Resume boundary
5. Artifact presence or absence
6. UI state surface requirements
7. Batch vs single run
8. Persistence boundary conditions

For each new candidate, ask:

- Does it protect a real behavioral contract?
- Is it distinguishable from existing cases?
- Would a regression here hurt product trust or system truthfulness?
- Is it stable enough to survive refactors?

Only keep cases that answer yes to most of these.

## Selection Rubric

Use this rubric to filter agent-generated candidate cases.

### Keep

- cases that protect core orchestration truth
- cases that expose silent regressions
- cases that pin state semantics across backend and UI
- cases that remain valid even if implementation changes

### Reject

- duplicate paraphrases of an existing case
- cases that freeze unstable prompt wording without protecting behavior
- cases that depend on subjective style judgment only
- cases that require an external full benchmark corpus to evaluate

### Escalate For Human Review

- cases that mix multiple contracts into one noisy scenario
- cases that reflect product strategy decisions not yet settled
- cases where the expected outcome is still ambiguous

## Suggested Directory Shape

A repo-safe starting point could look like this:

```text
test/
  golden/
    README.md
    cases/
      orchestration/
      ui/
      api/
    manifests/
```

This keeps Golden Data separate from package-local unit tests and consistent with the existing test architecture notes.

## First Authoring Workflow

1. Write 5 P0 cases by hand.
2. Review them for overlap and ambiguity.
3. Turn them into a minimal machine-readable schema.
4. Build a tiny validator against one layer first.
5. Only then start large-scale agent mutation.
6. Use the selection rubric to keep the set small and high-signal.

## Practical Rule

For PaperBanana, the first question is not:

"Did the model say something pretty?"

The first question is:

"Did the system behave truthfully, causally, and observably?"

That is the spirit of Golden Data in this repository.
