# Conventions

> Coding conventions observed in the PaperBanana frontend codebase.

---

## Naming

| Element | Convention | Example |
|---------|-----------|---------|
| React components | `PascalCase` | `GeneratePanel`, `StageCard`, `HistoryPopover` |
| Component filenames | `PascalCase.tsx` | `GeneratePanel.tsx`, `StageCard.tsx` |
| Hooks | `use` + `PascalCase` | `useGenerate()`, `useHistory()`, `useRefine()` |
| Hook filenames | `camelCase` with `use` prefix | `useGenerate.ts`, `useRefine.ts` |
| Helper functions | `camelCase` | `formatDuration()`, `fileToDataUrl()`, `detectMIMEType()` |
| Constants | `SCREAMING_SNAKE_CASE` | `API_BASE`, `REFINE_ENDPOINT` |
| TypeScript interfaces | `PascalCase` | `GenerateState`, `RefineResult`, `ApiError` |
| TypeScript type aliases | `PascalCase` | `StageStatus`, `ArtifactKind`, `SSEEventType` |
| Props interfaces | Component name + `Props` suffix | `GeneratePanelProps`, `StageCardProps` |
| State interfaces | Feature name + `State` suffix | `GenerateState`, `RefineState`, `HistoryState` |
| CSS custom properties | `--category-variant` | `--color-primary`, `--bg-status-success` |
| i18n keys | Dot-delimited paths | `generate.title`, `settings.provider.name` |

---

## Code Style

### Quotes

Mixed quote style is present in the codebase. Follow the prevailing style in the file you are editing:

- `App.tsx` and some newer files use **double quotes**: `import { App } from "./App"`
- `useGenerate.ts` and other hooks use **single quotes**: `import { useState } from 'react'`

**Rule**: Do not mix quote styles within a single file. If a file predominantly uses one style, follow it.

### Explicit Prop Typing

All component props use explicit TypeScript interfaces:

```typescript
// Good: typed props interface
export interface GeneratePanelProps {
  onGenerate: (prompt: string, options?: GenerateOptions) => void;
  isGenerating: boolean;
  disabled?: boolean;
}

export function GeneratePanel({ onGenerate, isGenerating, disabled }: GeneratePanelProps) {
  // ...
}
```

```typescript
// Bad: inline prop types
export function GeneratePanel({ onGenerate, isGenerating }: {
  onGenerate: (prompt: string) => void;
  isGenerating: boolean;
}) {
  // ...
}
```

### Inline Objects Over Abstraction

Prefer inline object construction over premature abstraction. Do not create intermediate types or helper functions unless they are reused in 3+ places:

```typescript
// Good: inline construction
const initialStages: StageState[] = stageOrder.map((stage) => ({
  stage,
  agent: agentNames[stage] || stage,
  status: 'pending' as StageStatus,
}));

// Bad: unnecessary abstraction for single use
function createStageState(stage: string, agent: string): StageState {
  return { stage, agent, status: 'pending' as StageStatus };
}
```

### Section-Divider Comments

Large files use section-divider comments to organize code:

```typescript
// ============================================
// Generate API - Types
// ============================================
```

```typescript
// ============================================================================
// Store Exports
// ============================================================================
```

Use this pattern when a file exceeds ~100 lines or has logically distinct sections.

---

## Import Organization

Imports follow this order within a file:

1. **Style imports** (CSS) -- always first when present
2. **React and third-party** libraries
3. **Local modules** (relative paths)

```typescript
// 1. Style imports
import "./themes/base.css";
import "./themes/workspace.css";

// 2. React and third-party
import { useEffect, useRef, lazy, Suspense } from "react";
import { QueryClientProvider } from "@tanstack/react-query";

// 3. Local modules (relative paths)
import { Layout, Header, Footer } from "./components";
import { useGenerationFlow, useToast } from "./hooks";
import { useAppStore, useGenerationStore } from "./stores";
```

### No Path Aliases

The project does not use TypeScript path aliases (`@/...`). All imports use **relative paths**:

```typescript
// Good: relative imports
import { apiClient } from '../lib/api';
import type { StageStatus } from '../components/StageCard';

// Bad: path aliases (not configured)
import { apiClient } from '@/lib/api';
```

This may change after the Tauri migration if path aliases are introduced.

---

## Error Handling

### API Error Class

All HTTP errors are wrapped in `ApiError` defined in `web/src/lib/api.ts`:

```typescript
export class ApiError extends Error {
  constructor(
    public status: number,
    public statusText: string,
    message: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}
```

The `handleResponse` helper in `api.ts` throws `ApiError` for non-2xx responses. Other transport files (like `refine.ts`) also construct `ApiError` manually when needed.

### Hook Error State

Custom hooks **expose error state** through their return value. They do **not** throw errors to the caller:

```typescript
// Pattern: hook exposes error state
export function useGenerate() {
  const [state, setState] = useState<GenerateState>({
    // ...
    error: null,
  });

  // Error is caught and stored in state
  } catch (err) {
    setState((prev) => ({
      ...prev,
      error: err instanceof Error ? err.message : 'Unknown error',
    }));
  }

  return {
    ...state,  // includes error: string | null
    generate,
    cancel,
    reset,
  };
}
```

```typescript
// Consumer checks error state
const { isGenerating, error, generate } = useGenerate();
// Render error in UI, don't try/catch the hook
```

### ErrorBoundary for Render Failures

`web/src/components/ErrorBoundary.tsx` wraps the app to catch React render errors. It is used in `App.tsx` as a top-level safety net:

```typescript
// App.tsx
<ErrorBoundary>
  <QueryClientProvider client={queryClient}>
    <App />
  </QueryClientProvider>
</ErrorBoundary>
```

### useToast for Action Failures

User-facing action failures (copy to clipboard, export failed) are reported via `useToast`:

```typescript
const { addToast } = useToast();

try {
  await copyImageToClipboard(dataUrl);
  addToast({ message: t('export.copied'), type: 'success' });
} catch {
  addToast({ message: t('export.copyFailed'), type: 'error' });
}
```

### Error Handling Summary

| Error Source | Strategy | Location |
|-------------|----------|----------|
| HTTP API errors | `ApiError` class, hook `error` state | `lib/api.ts`, hooks |
| SSE stream errors | Hook `error` state + `onError` callback | `useGenerate.ts` |
| React render errors | `ErrorBoundary` component | `components/ErrorBoundary.tsx` |
| User action failures | `useToast` notifications | Anywhere with user actions |
| Unknown/unexpected | Fallback message in hook state | hooks |

### Defensive API Response Handling

**Rule**: Never assume an API response field exists or is an array. Always use defensive patterns when accessing fields that cross the backend/frontend boundary:

```typescript
// Bad: implicit assumption that field exists and is an array
const items = response.items.map((item) => transform(item));

// Good: defensive with fallback
const items = (response.items || []).map((item) => transform(item));

// Good: optional chaining with nullish coalescing
const items = response.items?.map((item) => transform(item)) ?? [];

// Good: explicit validation with early return
if (!Array.isArray(response.items)) {
  setState({ items: [], error: 'Invalid response format' });
  return;
}
const items = response.items.map((item) => transform(item));
```

**Rationale**: TypeScript types are erased at runtime. The backend can return any JSON shape — missing fields, `null`, or unexpected structures. Defensive handling at the boundary prevents runtime crashes that TypeScript cannot catch.

**Applies to**: All custom hooks that consume `apiClient` responses, especially `useHistory`, `useGenerate`, `useGenerationFlow`, and any future hooks.

---

## Comments

### Section Dividers

Use section-divider comments in files with distinct logical sections (see Code Style above).

### JSDoc

JSDoc is used **selectively** for non-obvious utilities and public API functions. Do not add JSDoc to every function -- only when the purpose is not immediately clear from the name:

```typescript
/**
 * API Types - Unified API type definitions
 * Keep in sync with backend DTOs
 */
```

```typescript
// No JSDoc needed -- the name is self-documenting
function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
}
```

### Inline Comments

Use inline comments sparingly, primarily for:
- Explaining *why* a non-obvious decision was made
- Marking compatibility/fallback logic
- Noting edge cases

```typescript
// Legacy compatibility fields
type?: string;
format?: string;
```

```typescript
// Ignore network errors, abort locally anyway
} catch {
  // silently fail retry
}
```

---

## Function Design

### Hook Return Shape

Custom hooks return **composite state + actions objects**. State fields and action functions are spread together:

```typescript
export function useGenerate() {
  // ...
  return {
    // State
    ...state,           // isGenerating, stages, result, error, ...
    // Actions
    generate,
    cancel,
    reset,
    restore,
    // Helpers
    formatDuration,
  };
}
```

**Rule**: Always destructure at the call site to make usage explicit:

```typescript
const { isGenerating, error, generate } = useGenerate();
```

### Component API Design

Component APIs use typed props objects. Colocate the props interface with the component:

```typescript
export interface StageCardProps {
  stage: string;
  agent: string;
  status: StageStatus;
  summary?: string;
  error?: string;
  artifactCount?: number;
}

export function StageCard({ stage, agent, status, summary, error }: StageCardProps) {
  // ...
}
```

Export props types alongside the component through barrel files.

---

## Module Design

### Barrel Files

Every feature directory uses an `index.ts` barrel file to re-export its public API. See [Directory Structure](./directory-structure.md#barrel-exports) for details.

### Co-locate Tests

Test files live next to the source they test. See [Testing](./testing.md) for full details.

### Re-export Indirection

Some components use a re-export barrel pattern:

```typescript
// web/src/components/GeneratePanel.tsx
// Re-export from atomic design location for backward compatibility
export {
  GeneratePanel,
  type GeneratePanelProps,
  type GenerateOptions,
} from './organisms/generation/GeneratePanel';
```

This pattern exists for backward compatibility during a refactoring. When creating new components, do not use re-export indirection -- place the component directly in the feature directory.

---

## Risky Inconsistencies

These are known inconsistencies that have been observed in the codebase. Be aware of them and avoid making them worse:

### 1. Quote Style Drift

Single quotes and double quotes are mixed across files. This is a cosmetic issue but makes the codebase inconsistent. Follow the prevailing style in each file.

### 2. Testing Contract Drift

Some tests assert on literal CSS class names (e.g., `expect(el).toHaveClass('bg-green-500')`) while others use semantic design tokens (e.g., `expect(el).toHaveClass('bg-status-success')`). When writing tests, prefer semantic token assertions.

### 3. ESLint Declared But Not Installed

The `package.json` declares `"lint": "eslint ."` but ESLint is not in `devDependencies`. The lint script will fail. Do not rely on `npm run lint` until ESLint is properly installed and configured.

### 4. Vitest Does Not Exclude Playwright Specs

The `vitest.config.ts` excludes `**/*.spec.ts` files, but if a `.spec.ts` file is imported by a Vitest test, it can cause conflicts. Keep Vitest (`.test.ts`) and Playwright (`.spec.ts`) tests strictly separate.

### 5. Legacy vs V2 History Exports

Both `HistorySidebar` (legacy) and `HistoryPanel` (V2) are exported from `components/index.ts`. Importing the wrong one is easy. Always import from the `history/` feature barrel:

```typescript
// Good: import V2 from feature barrel
import { HistoryPanel } from './history';

// Bad: accidentally importing legacy
import { HistorySidebar } from './HistorySidebar';
```

### 6. Zone.Identifier Sidecar Files

Windows Zone.Identifier files (e.g., `clipboard.ts:Zone.Identifier`, `export.ts:Zone.Identifier`) have been committed to `web/src/lib/`. These are Windows metadata files and should not be versioned. Add `*:Zone.Identifier` to `.gitignore` and remove the committed files.
