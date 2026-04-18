# Testing

> Testing conventions, patterns, and known issues for the PaperBanana frontend.

---

## Overview

The frontend test suite uses Vitest as the test runner with jsdom environment, @testing-library/react for component testing, and Playwright for E2E tests. Tests are co-located with source files and organized by type.

---

## Framework

| Tool | Version | Purpose |
|------|---------|---------|
| Vitest | 3.2.4 | Unit/integration test runner |
| jsdom | 26.x | DOM environment for Vitest |
| @testing-library/react | 16.3.2 | Component rendering and querying |
| @testing-library/jest-dom | 6.6.3 | DOM matchers (toBeVisible, toHaveTextContent, etc.) |
| @testing-library/user-event | 14.6.1 | User interaction simulation |
| Playwright | 1.58.x | End-to-end browser testing |

---

## Test Configuration

### Vitest Configuration

Defined in `web/vitest.config.ts`:

```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test/setup.ts',
    exclude: [
      '**/node_modules/**',
      '**/dist/**',
      '**/e2e/**',
      '**/*.spec.ts',        // Exclude Playwright specs
      '**/test-ui.spec.ts',
    ],
  },
});
```

### Test Setup

`web/src/test/setup.ts` configures `@testing-library/jest-dom` matchers and any global test utilities.

### Playwright Configuration

Playwright tests live in `web/e2e/` and use `*.spec.ts` naming. They are excluded from Vitest by the `**/*.spec.ts` glob in `vitest.config.ts`.

---

## Test Organization

### Co-located Tests (Default)

Unit and integration tests live next to the source they test, mirroring the production filename:

```
web/src/
├── components/
│   ├── StageCard.tsx
│   ├── StageCard.test.tsx          # Component test
│   ├── FieldError.tsx
│   ├── FieldError.test.tsx         # Component test
│   ├── workspace/
│   │   ├── CandidateGrid.tsx
│   │   └── CandidateGrid.test.tsx  # Feature component test
│   └── history/
│       └── HistoryItem.test.tsx    # Feature component test
├── hooks/
│   └── useGenerate.test.ts         # Hook test (if present)
└── lib/
    ├── refine.ts
    └── refine.test.ts              # Utility test
```

### Golden Tests

Golden UI tests validate rendering output against expected snapshots or semantic structure:

```
web/src/test/golden/
```

Golden tests should assert on **semantic structure** (roles, labels, text content) rather than literal class names or DOM structure, which are brittle.

### Integration Tests

Integration tests cover cross-cutting flows like SSE streaming:

```
web/src/test/integration/
└── sse-flow.test.ts
```

### E2E Tests

End-to-end tests cover complete user flows through a real browser:

```
web/e2e/
└── p0-p1-core-flows.spec.ts
```

---

## Test Patterns

### Component Tests

Use `render` from `@testing-library/react` and `userEvent` for interactions:

```typescript
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { StageCard } from './StageCard';

describe('StageCard', () => {
  it('renders stage name and agent', () => {
    render(
      <StageCard
        stage="retriever"
        agent="Retriever"
        status="complete"
      />
    );
    expect(screen.getByText('Retriever')).toBeVisible();
  });

  it('shows error message when status is error', () => {
    render(
      <StageCard
        stage="visualizer"
        agent="Visualizer"
        status="error"
        error="Connection timed out"
      />
    );
    expect(screen.getByText('Connection timed out')).toBeVisible();
  });
});
```

### Hook Tests

Use `renderHook` from `@testing-library/react` and `waitFor` for async state changes:

```typescript
import { renderHook, act, waitFor } from '@testing-library/react';
import { useRefine } from './useRefine';

describe('useRefine', () => {
  it('initializes with idle state', () => {
    const { result } = renderHook(() => useRefine());
    expect(result.current.isRefining).toBe(false);
    expect(result.current.result).toBeNull();
    expect(result.current.error).toBeNull();
  });
});
```

### Async Assertions

Use `waitFor` for assertions on state that changes asynchronously:

```typescript
await waitFor(() => {
  expect(result.current.isGenerating).toBe(false);
});
```

Do not use `setTimeout` or manual delays in tests.

---

## Mocking

### Mocking Hooks

Use `vi.mock` to mock custom hooks in component tests:

```typescript
vi.mock('../hooks/useGenerate', () => ({
  useGenerate: () => ({
    isGenerating: false,
    stages: [],
    result: null,
    error: null,
    generate: vi.fn(),
    cancel: vi.fn(),
    reset: vi.fn(),
    restore: vi.fn(),
  }),
}));
```

### Mocking Translations

Mock `react-i18next` to avoid loading locale files in tests:

```typescript
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,  // Return the key itself
    i18n: { language: 'en' },
  }),
}));
```

### Mocking Network Requests

Mock `fetch` or `apiClient` for transport-layer tests:

```typescript
vi.mock('../lib/api', () => ({
  apiClient: {
    generate: vi.fn().mockResolvedValue({
      session_id: 'test-session',
      generated_artifacts: [],
    }),
  },
  ApiError: class extends Error {
    status: number;
    statusText: string;
    constructor(status: number, statusText: string, message: string) {
      super(message);
      this.status = status;
      this.statusText = statusText;
    }
  },
}));
```

### Mocking SSE

For SSE-based tests, mock the `streamGenerate` function:

```typescript
vi.mock('../lib/sse', () => ({
  streamGenerate: vi.fn(),
}));
```

For integration tests that exercise the SSE flow, use the real `streamGenerate` with a test server.

### Do Not Mock Basic React Rendering

In golden tests and component structure tests, do not mock basic React rendering. Test the actual component output:

```typescript
// Good: golden test renders the real component
render(<GeneratePanel onGenerate={vi.fn()} isGenerating={false} />);
expect(screen.getByRole('button', { name: /generate/i })).toBeVisible();

// Bad: mocking React internals
vi.mock('react', () => ({ ... }));
```

---

## Assertion Conventions

### Semantic Queries

Prefer semantic queries over structural queries:

```typescript
// Good: semantic query
screen.getByRole('button', { name: /generate/i });
screen.getByLabelText(/prompt/i);
screen.getByText(/retriever/i);

// Bad: structural query
screen.getByTestId('generate-button');
container.querySelector('.btn-primary');
```

### Class Assertions

When asserting on CSS classes, prefer semantic design tokens over raw utility classes:

```typescript
// Good: semantic token (stable across theme changes)
expect(el).toHaveClass('bg-status-success');

// Fragile: raw utility class (changes with theme)
expect(el).toHaveClass('bg-green-500');
```

### Accessibility Assertions

Include accessibility checks in component tests where applicable:

```typescript
expect(screen.getByRole('button')).toHaveAttribute('aria-label');
expect(screen.getByRole('dialog')).toHaveAttribute('aria-modal', 'true');
```

---

## Run Commands

| Command | Purpose |
|---------|---------|
| `cd web && npm run test:run` | Run all Vitest tests once |
| `cd web && npm run test` | Run Vitest in watch mode |
| `cd web && npx vitest --run path/to/test.tsx` | Run a single test file |
| `cd web && npx playwright test` | Run Playwright E2E tests |
| `cd web && npx playwright test --ui` | Run Playwright with UI |

---

## Known Issues

### 1. Failing Tests

There are approximately 31 failing tests in the suite. The primary causes are:

- **Playwright/Vitest conflict**: Some `.spec.ts` files are incorrectly discovered by Vitest despite the exclusion pattern.
- **Stale golden tests**: Golden tests reference outdated component structure or class names.
- **Semantic token drift**: Tests assert on raw color classes (`bg-green-500`) that have been replaced with semantic tokens (`bg-status-success`).

When fixing failing tests, prioritize:
1. Updating assertion methods (semantic tokens over raw classes)
2. Fixing test setup (mock alignment with current hook APIs)
3. Updating golden test snapshots

### 2. ESLint Not Available

`npm run lint` will fail because ESLint is declared in `package.json` scripts but not installed as a dependency. Do not rely on lint for test validation. Use `tsc` for type checking instead:

```bash
cd web && npx tsc --noEmit
```

### 3. Test Isolation

Some tests share state through module-level mocks. Ensure each test resets mocks in `afterEach`:

```typescript
afterEach(() => {
  vi.restoreAllMocks();
});
```

### 4. Playwright Spec Discovery

Vitest's `exclude` pattern filters `**/*.spec.ts` but does not handle all edge cases. If you see Vitest trying to run Playwright specs, verify the `vitest.config.ts` exclude list and ensure no `.spec.ts` file is imported by a `.test.ts` file.

### 5. Zone.Identifier Files

Windows Zone.Identifier sidecar files (e.g., `refine.ts:Zone.Identifier`) are present in the `lib/` directory. These are not test files and should not be executed. If Vitest attempts to run them, the exclude pattern should catch them, but it is better to remove them from version control entirely.

### 6. E2E Screenshots Can Catch Runtime Errors

**Case study (2026-04-18)**: The HistoryPanel showed a raw JavaScript error (`Cannot read properties of undefined (reading 'map')`) in E2E screenshots. The error was invisible in unit tests because the mock data always included the expected field. Only when running against the real backend with edge-case responses did the UI crash.

**Takeaway**: E2E tests with `fullPage: true` screenshots are valuable for discovering cross-layer bugs that unit tests miss. Review screenshots manually after significant changes, especially for panels, dialogs, and error states that may render outside the initial viewport.

---

## Test Coverage Guidelines

### What to Test

| Priority | What | How |
|----------|------|-----|
| P0 | Hook state transitions | `renderHook` + `waitFor` |
| P0 | Component rendering (happy path) | `render` + semantic queries |
| P0 | Error state rendering | `render` with error props |
| P1 | User interactions (click, type) | `userEvent` + assertions |
| P1 | Async workflows (generate, refine) | Mocked transport + `waitFor` |
| P2 | Accessibility (ARIA, keyboard nav) | `toHaveAttribute`, `userEvent.keyboard` |
| P2 | Edge cases (empty state, long text) | `render` with boundary props |

### What Not to Test

- React internals (state updates, re-renders)
- Third-party library behavior (i18next, React Query)
- CSS implementation details (exact pixel values)
- Network layer (test the mock, not the real fetch)

---

## Adding a New Test

Checklist:

1. Create `.test.tsx` (component) or `.test.ts` (utility) co-located with the source file.
2. Use `describe`/`it` blocks for organization.
3. Import from `@testing-library/react` -- do not use Enzyme.
4. Use `vi.mock` for hooks and transport layer.
5. Use `vi.restoreAllMocks()` in `afterEach`.
6. Assert with semantic queries and design tokens.
7. Run the test: `cd web && npx vitest --run path/to/your.test.tsx`
