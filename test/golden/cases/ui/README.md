# UI Golden Data Layer

## Purpose

This layer protects the contract between the system's execution state and what the user can see and do.

It does not test visual styling.
It tests whether the UI tells the truth about the system.

## Core Principle

PaperBanana is a long-running, multi-stage generation workspace.
The user submits a task and then waits.
During that wait, the UI is the only window into a complex backend process.

If the UI lies — by hiding a failure, faking progress, or dropping a final artifact — the user loses trust and loses control.

The UI Golden Data layer exists to make that lying impossible to introduce silently.

## What This Layer Protects

### 1. Progress Visibility Contract

When a task is running, the user must be able to see:
- that the system is working
- which stage is currently active
- that progress is real, not a spinner pretending to think

### 2. Failure Surface Contract

When a task fails, the user must be able to see:
- that it failed (not a silent drop or eternal spinner)
- at which stage it failed
- enough information to decide whether to retry or investigate

### 3. Completion and Artifact Contract

When a task completes, the user must be able to see:
- that it is done
- the final output artifact (image, code, or both)
- a stable, non-transient completed state

### 4. Resume Semantics Contract

When a task was resumed from a snapshot, the user should be able to see:
- that this was a resumed run, not a fresh one
- which stage the resume started from
- that prior outputs are available and consistent

### 5. Batch State Contract

When running a batch of tasks, the user must be able to see:
- per-task status
- which tasks are running, completed, or failed
- batch-level summary

## Cases In This Layer

| ID | Title | Priority |
|----|-------|----------|
| GD-UI-001 | Stage Progress Visible While Running | P0 |
| GD-UI-002 | Failure Location Visible On Stage Error | P0 |
| GD-UI-003 | Artifact Surfaced On Completion | P0 |
| GD-UI-004 | Resumed Task Exposes Resume Semantics | P0 |
| GD-UI-005 | Batch Per-Task Status Visible | P1 |

## Evaluation Notes

UI Golden Data cases are evaluated against the frontend component and hook layer.
They should be wired to Vitest tests that assert on rendered state, not visual snapshots.

The validator should check:
- component receives the correct state props
- correct UI elements are rendered given the state
- forbidden states (e.g. completed badge without artifact) never appear together

Do not use pixel diffing.
Use state-based assertions.
