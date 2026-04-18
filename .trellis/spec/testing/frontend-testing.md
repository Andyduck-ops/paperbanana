# Frontend Testing

## Runner and Setup

**Runner:** Vitest 3.2.4

**Run commands:**

```bash
cd web && npm run test:run    # Run once
cd web && npm test            # Watch mode
```

**Configuration:** `web/vitest.config.ts` and `web/vite.config.ts`

**Setup file:** `web/src/test/setup.ts`

- jsdom environment
- Vitest globals enabled
- `@testing-library/jest-dom` matchers imported
- `afterEach` cleanup for Testing Library

```typescript
// web/src/test/setup.ts (conceptual)
import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

afterEach(() => {
  cleanup();
});
```

## Test File Organization

Tests are co-located with the source they test:

```text
web/src/components/GeneratePanel.tsx
web/src/components/GeneratePanel.test.tsx
web/src/hooks/useGenerate.ts
web/src/hooks/useGenerate.test.ts
web/src/test/golden/GD-UI-001.test.tsx
web/src/test/integration/sse-flow.test.ts
```

### Naming Conventions

| Type | Pattern | Example |
|------|---------|---------|
| Component test | `*.test.tsx` | `GeneratePanel.test.tsx` |
| Hook test | `*.test.ts` | `useGenerate.test.ts` |
| Golden test | `GD-*.test.tsx` | `GD-UI-001.test.tsx` |
| Integration test | `*.test.ts` | `sse-flow.test.ts` |
| Playwright spec | `*.spec.ts` | `test-ui.spec.ts` |

## Patterns

### Component Testing

```typescript
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('../../hooks', () => ({
  useLanguage: () => ({ t: (key: string) => key }),
}));

describe('ProgressPanel', () => {
  it('should show current active stage name', () => {
    render(<ProgressPanel stages={runningState} />);
    expect(screen.getByText('Planner')).toBeInTheDocument();
  });
});
```

### Hook Testing

```typescript
import { renderHook, waitFor } from '@testing-library/react';
import { useGenerate } from './useGenerate';

describe('useGenerate', () => {
  it('should return expected value', async () => {
    const { result } = renderHook(() => useGenerate());
    await waitFor(() => {
      expect(result.current.result).not.toBeNull();
    });
  });
});
```

### Async State Transitions

Use `act` and `waitFor` for async state changes:

```typescript
const { result } = renderHook(() => useGenerate());

await act(async () => {
  result.current.generate('test prompt');
});

await waitFor(() => {
  expect(result.current.result).not.toBeNull();
});
```

### Setup/Teardown

```typescript
beforeEach(() => {
  // Reset mocks, stub globals, etc.
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});
```

## Mocking

### Framework

`vi.mock`, `vi.spyOn`, and `vi.stubGlobal` for frontend mocking.

### What to Mock

- **Hooks** when isolating a component from hook dependencies
- **Translations** via `vi.mock` on i18n modules
- **Network/SSE boundaries** when isolating from backend
- **Global APIs** (e.g., `fetch`, `EventSource`) via `vi.stubGlobal`

```typescript
vi.mock('../lib/sse', () => ({
  streamGenerate: vi.fn(async (_data, options) => {
    options.onStageStart?.({ stage: 'retriever', agent: 'Retriever' });
    options.onResult?.({ session_id: 'test-session', generated_artifacts: [] });
  }),
}));
```

### What NOT to Mock

- **Basic React rendering** in golden tests -- the golden suite intentionally
  renders real components and asserts user-visible behavior
- **DOM structure** -- prefer accessible queries over implementation details

## TDD Workflow (from ADR-004)

Follow Red-Green-Refactor with Vitest:

1. **Red:** Write a failing test with Vitest
2. **Green:** Implement the minimal component/hook to pass
3. **Refactor:** Clean up with test safety

Target pattern:

```typescript
// 1. RED: Write failing test first
describe('useHook', () => {
  it('should return expected value', async () => {
    const { result } = renderHook(() => useHook());
    await waitFor(() => {
      expect(result.current.data).toEqual(expected);
    });
  });
});

// 2. GREEN: Implement minimal code to pass

// 3. REFACTOR: Clean up while tests stay green
```

## Fixtures

Frontend fixtures are mostly inline in the test file:

```typescript
const runningState: StageState[] = [
  { stage: 'retriever', agent: 'Retriever', status: 'complete' },
  { stage: 'planner', agent: 'Planner', status: 'running' },
  { stage: 'stylist', agent: 'Stylist', status: 'pending' },
];
```

Golden scenario identifiers map to documentation under `test/golden/` and
`test/golden/cases/ui/`.

## Known Issues

As of 2026-03-25, `cd web && npm run test:run` reports **31 failing tests** across
**16 failed files** (out of 54 files, 487 tests total).

### Root Causes

| Cause | Example | Impact |
|-------|---------|--------|
| Playwright/Vitest conflict | `web/vite.config.ts` does not exclude `*.spec.ts` | Playwright `test()` runs under Vitest |
| i18n key drift | `theme.popArt` vs `theme.options.*` | `web/src/i18n/index.test.ts` fails |
| Semantic token drift | Raw `green`, `red-500` vs `bg-status-success/20` | Golden tests fail |
| Stale component tests | `HistoryItem.test.tsx` vs current BEM implementation | Class name mismatches |
| Missing icon assertions | `DangerZone.test.tsx` expects literal icon node | Component no longer renders it |
| Prompt label changes | Combined labels vs `Paper Context & References` / `Target Figure Brief` | `GD-001-M-empty-prompt.test.tsx` fails |

### Warnings

- `BatchProgressPanel.test.tsx`: jsdom navigation warnings (component triggers
  link navigation that jsdom does not implement)
- `BatchProgressPanel.test.tsx`, `ImageUpload.test.tsx`: `act(...)` warnings
  (state transitions not consistently awaited)
- `test-ui.spec.ts`: Uses `waitForTimeout()` and screenshots -- timing-based,
  not suitable for unit test suite

## Playwright Configuration

- **Config:** `web/playwright.config.ts`
- **Browser:** Chromium only
- **Reports:** HTML reports, screenshots on failure
- **IMPORTANT:** Config only matches `test-ui.spec.ts`; specs under
  `web/e2e/` are NOT included in the default Playwright run unless
  the configuration changes
