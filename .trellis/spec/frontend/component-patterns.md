# Component Patterns

> Design patterns and conventions for React components in PaperBanana.

---

## Overview

This document describes the patterns used for building React components, including input handling, styling, internationalization, testing, and known fragile areas.

---

## Input Handling

### Structured Input

The generate workflow uses structured input even when a combined prompt string is also built. Components like `DualInputPanel` collect structured fields (prompt, visual intent, content) separately, and the hook assembles them into the API request:

```typescript
// DualInputPanel collects structured input
<DualInputPanel
  prompt={prompt}
  onPromptChange={setPrompt}
  visualIntent={visualIntent}
  onVisualIntentChange={setVisualIntent}
  content={content}
  onContentChange={setContent}
/>

// useGenerate assembles the API request
await generate(prompt, {
  visualizerNode,
  content,
  visualIntent,
  config: {
    aspect_ratio,
    critic_rounds,
    retrieval_mode,
    pipeline_mode,
    query_model,
    gen_model,
  },
});
```

**Rule**: When adding new input fields, add them as structured props rather than concatenating into a single string. The API request assembly happens in the hook, not the component.

### Image Upload

Image upload components (e.g., `ImageUpload.tsx`) handle file selection and preview URL generation. They pass `File` objects up to the parent, which converts them to data URLs in the transport layer:

```typescript
// lib/refine.ts converts File to data URL at transport time
async function fileToDataUrl(file: File): Promise<string> {
  return await new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = () => reject(reader.error ?? new Error("Failed to read file"));
    reader.readAsDataURL(file);
  });
}
```

---

## Styling

### Semantic Design Tokens

Prefer semantic design tokens over raw color utility classes. The theme CSS files define custom properties that provide semantic meaning:

```css
/* Good: semantic tokens in theme CSS */
--color-primary: #2563eb;
--bg-status-success: #22c55e;
--bg-status-error: #ef4444;
```

```typescript
// Good: use semantic class names
<div className="bg-status-success" />
<div className="text-primary" />

// Bad: raw color classes break theming
<div className="bg-green-500" />
<div className="text-blue-600" />
```

When writing new components, check the existing theme CSS files in `web/src/themes/` for available tokens. If a needed semantic token does not exist, add it to the base theme and all variant themes.

### CSS Theme System

The project uses multiple CSS theme files loaded via `App.tsx`:

```typescript
import "./themes/base.css";
import "./themes/qi-baishi.css";
import "./themes/pop-anime.css";
import "./themes/rococo.css";
import "./themes/japanese-bw.css";
import "./themes/workspace.css";
```

Theme switching works by toggling a CSS class on the root element. Each theme file overrides the base custom properties.

### Tailwind Usage

- Use Tailwind utility classes for layout and spacing.
- Use CSS custom properties for colors and theme-dependent values.
- Do not hardcode color values in Tailwind classes -- use the theme tokens.
- Complex component styles can use the `@apply` directive in CSS files if needed.

---

## Internationalization (i18n)

### Translation Keys

All user-facing strings must go through i18n. Translation keys are defined in locale files:

```
web/src/i18n/locales/
├── en.json    # English (source of truth)
└── zh.json    # Chinese
```

### Using Translations in Components

Route all translations through `useTranslation` from `react-i18next` or the project's `useLanguage` hook:

```typescript
import { useTranslation } from 'react-i18next';

function GeneratePanel() {
  const { t } = useTranslation();
  return <h2>{t('generate.title')}</h2>;
}
```

**Rules**:
- Never hardcode user-facing strings in components.
- Add new keys to both `en.json` and `zh.json` simultaneously.
- Use dot-delimited paths for keys: `generate.title`, `settings.provider.name`.
- Keep keys descriptive but concise.

### Language Switching

Language switching is managed through the Zustand `appStore`:

```typescript
const { language, setLanguage } = useAppStore();
```

The `i18next` instance is initialized in `web/src/i18n/index.ts` and the active language is synced from the store.

---

## Component Composition

### Feature Components vs Primitives

The component tree distinguishes between feature-level components and primitive components:

| Type | Location | Examples |
|------|----------|---------|
| Feature components | `components/feature-name/` | `workspace/Workspace.tsx`, `history/HistoryPanel.tsx` |
| Domain components | `components/` root | `GeneratePanel.tsx`, `ResultPanel.tsx`, `StageCard.tsx` |
| Layout primitives | `components/layout/` | `MinimalLayout.tsx` |
| Input primitives | `components/input/` | `FloatingInput.tsx` |
| Progress primitives | `components/progress/` | `ProgressLine.tsx` |
| Model config | `components/model/` | `ModelPopover.tsx`, `ChannelManager.tsx` |
| Settings | `components/settings/` | `SettingsDrawer.tsx`, `ProviderForm.tsx` |
| Refine | `components/refine/` | `IterationTimeline.tsx`, `BatchVariantGrid.tsx` |

### Prop Drilling Limit

If you are passing a prop through 3+ levels of components, lift the state to a Zustand store or use composition:

```typescript
// Bad: excessive prop drilling
<Workspace>
  <ResultArea artifact={artifact} onExport={onExport} onRefine={onRefine}>
    <ArtifactPreview artifact={artifact} onExport={onExport} onRefine={onRefine}>
      <ArtifactActions onExport={onExport} onRefine={onRefine} />
    </ArtifactPreview>
  </ResultArea>
</Workspace>

// Better: use Zustand store for shared state
// The inner components read from the store directly
<Workspace>
  <ResultArea />
    <ArtifactPreview />
      <ArtifactActions />  {/* reads export/refine from store */}
```

### Relative Imports

Use relative imports. Import from nearby helpers before shared abstractions:

```typescript
// Good: relative import, nearby first
import { formatDuration } from '../hooks/useGenerate';
import { apiClient } from '../lib/api';

// Bad: reaching across feature boundaries
import { formatDuration } from '../../workspace/hooks/formatDuration';
```

---

## Co-located Tests

Test files live next to the component they test:

```
components/
├── StageCard.tsx
├── StageCard.test.tsx       # Co-located test
├── FieldError.tsx
└── FieldError.test.tsx      # Co-located test
```

See [Testing](./testing.md) for full testing conventions.

---

## Fragile Areas

These are areas of the component tree where extra care is needed during changes:

### 1. Duplicate Config Surfaces

Model configuration can be accessed through two surfaces:

- **Channel Manager** (`components/model/ChannelManager.tsx`) -- the main settings view
- **Model Popover** (`components/model/ModelPopover.tsx`) -- quick access popover

Both surfaces read from the same Zustand `providerStore` but present different editing interfaces. Changes to one must be consistent with the other. When modifying provider/channel editing logic, test both surfaces.

### 2. Legacy vs V2 History Components

Two sets of history components exist:

- **V2**: `components/history/HistoryPanel.tsx`, `HistoryPopover.tsx`, `HistoryItem.tsx`
- **Legacy** (deprecated): `components/HistorySidebar.tsx`, `components/HistoryItem.tsx`

The barrel file exports both with explicit deprecation markers:

```typescript
// components/index.ts
// History components (V2 - Sliding Panel)
export { HistoryPanel, HistoryPopover, HistoryItem } from './history';
// Legacy History components (deprecated, use V2 above)
export { HistorySidebar } from './HistorySidebar';
export { HistoryItem as HistoryItemLegacy } from './HistoryItem';
```

**Rule**: New code must use V2 components only. Do not import legacy components. Remove legacy components when the migration is complete.

### 3. GeneratePanel Re-export Indirection

`components/GeneratePanel.tsx` is a re-export barrel that points to `components/organisms/generation/GeneratePanel.tsx`:

```typescript
// components/GeneratePanel.tsx
export { GeneratePanel, type GeneratePanelProps, type GenerateOptions }
  from './organisms/generation/GeneratePanel';
```

This exists for backward compatibility. When editing GeneratePanel, edit the file at `organisms/generation/GeneratePanel.tsx`, not the re-export barrel.

### 4. SSE Event Type Coupling

SSE event types in `lib/sse.ts` must match the backend constants defined in `internal/domain/agent/events.go`:

```typescript
// lib/sse.ts - Must match backend EventType constants
export type SSEEventType =
  | 'run_started'
  | 'stage_started'
  | 'stage_completed'
  | 'stage_failed'
  | 'run_completed'
  | 'run_failed'
  | 'run_canceled'
  | 'result'
  | 'error'
  | 'resume_start'
  | 'batch_start'
  | 'candidate_start'
  | 'candidate_complete'
  | 'batch_complete';
```

If the backend adds a new event type, this union must be updated. This is a cross-layer concern -- see the [Cross-Layer Thinking Guide](../guides/cross-layer-thinking-guide.md).

### 5. Artifact Type Drift

The `Artifact` type in `types/api.ts` has both canonical fields and legacy compatibility fields:

```typescript
export interface Artifact {
  id: string;
  kind: ArtifactKind;
  mime_type: string;
  data?: string;      // Base64 encoded (current)
  asset_id?: string;
  summary?: string;
  // Legacy compatibility fields
  type?: string;
  format?: string;
  width?: number;
  height?: number;
}
```

When consuming artifacts, prefer the canonical fields (`kind`, `mime_type`, `data`). Legacy fields exist for backward compatibility with older backend responses.

---

## Adding a New Component

Checklist for adding a new component:

1. Place it in the appropriate feature directory or `components/` root.
2. Define a typed props interface with the `Props` suffix.
3. Export from the feature barrel and `components/index.ts`.
4. Add co-located `.test.tsx` file.
5. Use semantic design tokens for colors.
6. Route all user-facing strings through i18n.
7. Use relative imports, nearby helpers first.
8. If the component needs global state, read from Zustand stores -- do not create new contexts.
