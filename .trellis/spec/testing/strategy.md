# Testing Strategy

## Testing Pyramid

PaperBanana follows a pragmatic testing pyramid with three tiers:

```
        ┌──────────┐
        │   E2E    │  Playwright (Chromium only)
        │  Tests   │  web/test-ui.spec.ts, web/e2e/*.spec.ts
        ├──────────┤
        │Integration│  Backend: httptest + in-memory SQLite
        │   Tests   │  Frontend: SSE boundary (web/src/test/integration/)
        ├──────────┤
        │   Unit   │  Backend: go test + testify
        │  Tests   │  Frontend: Vitest + Testing Library + jsdom
        └──────────┘
```

### Unit Tests (Base Layer)

- **Backend:** Go `testing` package + `github.com/stretchr/testify`
  - Table-driven test pattern with `t.Run` subcases
  - `require` for setup/fatal checks, `assert` for detailed assertions
  - `cmp.Diff` for structured comparison
  - In-memory SQLite and `t.TempDir()` for persistence tests

- **Frontend:** Vitest 3.2.4 + `@testing-library/react` + jsdom
  - Co-located test files (`Foo.test.tsx` next to `Foo.tsx`)
  - `render` for components, `renderHook` for hooks
  - `waitFor` / `act` for async state transitions
  - Globals enabled; shared setup/cleanup in `web/src/test/setup.ts`

### Integration Tests (Middle Layer)

- **Backend:** `httptest.NewRecorder()` for handler tests with real Gin router;
  `httptest.NewServer()` for transport-layer tests
- **Frontend:** SSE flow integration via `web/src/test/integration/sse-flow.test.ts`
  (mocks the network boundary, not the real backend)

### E2E Tests (Top Layer)

- **Runner:** `@playwright/test 1.58.2`
- **Config:** `web/playwright.config.ts` (Chromium only, screenshots on failure)
- **Gap:** Config only matches `test-ui.spec.ts`; specs under `web/e2e/` are NOT
  included in the default Playwright run unless the config changes

## Framework Choices

| Layer | Tool | Version | Rationale |
|-------|------|---------|-----------|
| Backend unit | Go testing + testify | stdlib + v1.9+ | Idiomatic Go; table-driven tests |
| Frontend unit | Vitest | 3.2.4 | Fast, Vite-native, ESM support |
| Frontend DOM | Testing Library | latest | User-centric queries, accessible |
| E2E | Playwright | 1.58.2 | Cross-browser capable; currently Chromium-only |
| Frontend assertion | jest-dom matchers | latest | DOM-specific matchers (`toBeInTheDocument`, etc.) |

## Coverage Policy

- **No enforced coverage thresholds** -- neither frontend nor backend
- **No CI pipeline detected** -- no GitHub Actions, no coverage gates
- Coverage can be viewed locally but is not wired into any checked-in script

## TDD Workflow (ADR-004)

Both stacks follow Red-Green-Refactor:

1. **Red:** Write a failing test first
2. **Green:** Implement the minimal code to pass
3. **Refactor:** Clean up with the test safety net

See [ADR-004](../adr/ADR-004-tdd-workflow.md) for the full decision record.

## Run Commands

```bash
# Backend
go test ./...

# Frontend (once)
cd web && npm run test:run

# Frontend (watch)
cd web && npm test

# E2E
cd web && npx playwright test -c playwright.config.ts
```
