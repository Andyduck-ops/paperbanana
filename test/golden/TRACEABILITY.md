# Core Baseline Traceability

This document connects the Phase `01.1` baseline IDs (`PB-*`) to the existing Golden Data inventory (`GD-*`).

Use it to answer three questions quickly:

1. Which core journeys are already protected?
2. Which journeys are only partially protected?
3. Which journeys still have no constitutional Golden Data at all?

## Coverage Rules

- `covered`: existing `GD-*` cases are strong enough to use as the current baseline
- `partial`: some behavior is protected, but the exact repo baseline is still under-specified
- `missing`: no existing `GD-*` case currently protects this journey

## Matrix

| Baseline ID | Journey | Status | Existing GD cases | Immediate action |
|---|---|---|---|---|
| `PB-CORE-001` | single generate happy path | covered | `GD-001`, `GD-UI-001`, `GD-UI-003`, `GD-API-001` | keep as constitutional P0 |
| `PB-CORE-002` | planner fallback / empty planner output | partial | `GD-002`, `GD-002-M-fail-at-planner`, `GD-001-M-partial-plan-from-planner` | add explicit silent-fallback golden case |
| `PB-CORE-003` | retrieval references remain causally attached | partial | `GD-004`, retrieval variants | add restorable reference-bundle case |
| `PB-BATCH-001` | per-candidate batch visibility | covered | `GD-BATCH-004`, `GD-BATCH-006`, `GD-UI-005`, `GD-UI-BATCH-001` | keep as constitutional P0 |
| `PB-BATCH-002` | mixed batch success/failure truthfulness | covered | `GD-BATCH-007`, `GD-UI-005-*` | keep as constitutional P0 |
| `PB-BATCH-003` | promote candidate into active branch | missing | none | add new P0 golden case before branch-review repair |
| `PB-REFINE-001` | refine returns image artifact | covered | `GD-API-016` | keep as constitutional P0 |
| `PB-REFINE-002` | refine iteration flags honored | covered | `GD-API-017` | keep as constitutional P0 |
| `PB-REFINE-003` | refine failure semantics | missing | none | add new P1 case |
| `PB-HIST-001` | restore generated session | covered | `GD-003`, `GD-API-005`, `GD-UI-004` | keep as constitutional P0 |
| `PB-HIST-002` | restore batch session | partial | batch truthfulness cases only | add explicit batch-restore case |
| `PB-HIST-003` | restore refine session | covered | `GD-UI-021` | keep as constitutional P0 |
| `PB-CONFIG-001` | provider key CRUD contract | missing | none | add API golden cases first |
| `PB-CONFIG-002` | query/gen role assignment affects runtime | partial | `GD-UI-006` | add API + runtime integration case |
| `PB-CONFIG-003` | config hot reload sync | missing | none | add config stream case |

## Practical Repair Order

### Wave 1
- `PB-REFINE-001`
- `PB-REFINE-002`
- `PB-HIST-003`

Reason:
- refine is currently the clearest broken product journey
- refine and restore are tightly coupled in the active UI shell

### Wave 2
- `PB-CONFIG-001`
- `PB-CONFIG-002`
- `PB-CONFIG-003`

Reason:
- current provider/model configuration has contract drift and weak runtime proof

### Wave 3
- `PB-BATCH-003`
- `PB-HIST-002`

Reason:
- candidate promotion and batch restore define whether batch is a real workspace flow or just a temporary progress display

### Wave 4
- `PB-CORE-002`
- `PB-CORE-003`

Reason:
- these are semantic quality protections and should be repaired after the core runnable product paths are trustworthy

## Rule

From this point on, every repair should name:

- which `PB-*` baseline IDs it touches
- which existing `GD-*` cases it reuses
- which new `GD-*` cases it must add
