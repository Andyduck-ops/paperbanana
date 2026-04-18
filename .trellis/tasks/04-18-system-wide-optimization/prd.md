# System-wide Project Optimization

## Goal

Systematically resolve all known architectural, engineering, and quality issues across the PaperBanana project (frontend + backend), following the Trellis task workflow.

## Problem Inventory

### Frontend — Architecture
1. **useToast fragmentation** — `useToast()` is called independently in `App.tsx`, `MainWorkspace.tsx`, `useGenerationFlow.ts`, `useGenerationStateMachine.ts`, and multiple pages. Each caller maintains isolated toast state, so toasts triggered deep in the component tree may never render.
2. **Router deficiency** — No React Router / TanStack Router. Routing is manual `if` blocks in `App.tsx` with `window.location.pathname` + `popstate`. No nested routes, params, or navigation guards.
3. **Atomic Design in name only** — `atoms/Button.tsx` exists but `Workspace.tsx` uses raw `<button>` elements. `atoms/`, `molecules/`, `organisms/`, `templates/` directories exist but are inconsistently populated.
4. **Legacy component debt** — `HistorySidebar` and `HistoryItemLegacy` were removed, but re-export indirections (`GeneratePanel.tsx` barrel pointing to `organisms/generation/GeneratePanel.tsx`) still create confusion.

### Frontend — Engineering
5. **ESLint not installed** — `package.json` declares `"lint": "eslint ."` but ESLint is not in `devDependencies`. Code quality enforcement is absent.
6. **Vitest / Playwright overlap** — `vitest.config.ts` excludes `**/*.spec.ts` but no enforcement prevents accidental mixing.
7. **Zone.Identifier pollution** — Windows metadata files (`*.Identifier`) committed to repo.
8. **Font loading bloat** — `index.html` loads 5 Google Fonts simultaneously, many unused per theme.

### Frontend — Testing
9. **Known test failures** — 31 failing Vitest tests and stale golden tests (e.g., `GD-UI-004.test.tsx` assertions on `className` with undefined values).
10. **Snapshot drift** — `__snapshots__` directories scattered; some may be outdated.

### Backend — Stability
11. **Cross-layer contract drift** — `lib/sse.ts` event types must match `internal/domain/agent/events.go`. No automated verification.
12. **Error handling inconsistency** — Some handlers may not wrap errors with `fmt.Errorf("context: %w", err)` per spec.

## Requirements

### R1: Global Toast System
- Replace isolated `useToast()` calls with a **Zustand-backed toast store** (`toastStore.ts`).
- `useToast()` becomes a selector hook reading from the global store.
- Single `<Toast />` instance rendered at app root. Remove all other `<Toast />` instances.

### R2: Install & Configure ESLint
- Add ESLint + `@eslint/js` + `typescript-eslint` + `eslint-plugin-react-hooks` to `devDependencies`.
- Create `eslint.config.js` aligned with project conventions (no path aliases, relative imports, semantic tokens).
- Fix all auto-fixable violations; document remaining manual fixes in task notes.

### R3: Clean Component Architecture
- **Commit to Atomic Design**: ensure `atoms/Button.tsx` is actually used in `organisms/` and `workspace/` components instead of raw `<button>`.
- Remove re-export indirection where safe (`GeneratePanel.tsx` root barrel → direct import from `organisms/generation/`).
- Consolidate orphaned root-level components into feature directories.

### R4: Fix Frontend Tests
- Fix `useTheme.test.ts` (already updated in previous task — verify pass).
- Fix golden tests that assert on undefined `className` (`GD-UI-004.test.tsx`).
- Delete obsolete snapshots.

### R5: Backend Quality Sweep
- Audit all Go handlers for error-wrapping compliance.
- Verify SSE event type union in `lib/sse.ts` matches backend constants.

### R6: Repository Hygiene
- Add `*:Zone.Identifier` to `.gitignore`.
- Remove committed `Zone.Identifier` files.
- Optimize Google Fonts loading in `index.html` (subset + `font-display: swap`).

## Acceptance Criteria

- [ ] `npm run lint` executes successfully with zero auto-fixable errors
- [ ] All toast notifications appear reliably regardless of which component triggers them
- [ ] `npm run test:run` shows zero new failures (existing known failures documented)
- [ ] `atoms/Button.tsx` is used in at least 3 core components
- [ ] No committed `Zone.Identifier` files remain
- [ ] Backend error wrapping audited; SSE types verified

## Technical Notes

- Keep changes scoped to the above requirements. Do not refactor business logic unrelated to these issues.
- Follow existing frontend spec: explicit Props interfaces, barrel exports, relative imports, semantic design tokens.
- Follow existing backend spec: `fmt.Errorf` wrapping, Zap structured logs, repository pattern.
- Update `.trellis/spec/` when new patterns or decisions are made.
