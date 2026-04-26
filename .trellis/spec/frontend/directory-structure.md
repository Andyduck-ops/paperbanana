# Directory Structure

> How frontend code is organized in the PaperBanana `web/src/` directory.

---

## Overview

The frontend follows a feature-based organization with barrel exports, co-located tests, and a clear separation between UI components, stateful hooks, transport logic, and types.

> **[PRE-TAURI]**: The transport layer (`lib/api.ts`, `lib/sse.ts`) currently uses browser `fetch` and `EventSource` against a standalone Go server on `:8080`. After the Tauri migration, these will be replaced with Tauri IPC commands. The directory structure will remain the same, but transport implementations will change.

---

## Directory Layout

```
web/src/
├── App.tsx                         # Main SPA shell and feature coordination
├── main.tsx                        # Frontend bootstrap (StrictMode, QueryClientProvider, i18n)
├── components/                     # UI building blocks
│   ├── index.ts                    # Barrel exports for all components
│   ├── workspace/                  # Workspace feature components
│   │   ├── Workspace.tsx
│   │   ├── CandidateGrid.tsx
│   │   └── ...
│   ├── history/                    # History panel components (V2 - Sliding Panel)
│   │   ├── index.tsx               # Feature barrel export
│   │   ├── HistoryPanel.tsx
│   │   ├── HistoryPopover.tsx
│   │   ├── HistoryItem.tsx
│   │   └── ...
│   ├── model/                      # Model config components (V2 - Unified Model Config)
│   │   ├── ModelIndicator.tsx
│   │   ├── ModelPopover.tsx
│   │   ├── ModelSelector.tsx
│   │   ├── ChannelManager.tsx
│   │   └── RoleMapping.tsx
│   ├── settings/                   # Settings drawer components
│   │   ├── SettingsDrawer.tsx
│   │   └── ProviderForm.tsx
│   ├── refine/                     # Refinement components
│   │   ├── IterationTimeline.tsx
│   │   └── BatchVariantGrid.tsx
│   ├── layout/                     # Layout primitives
│   │   └── MinimalLayout.tsx
│   ├── input/                      # Input primitives
│   │   └── FloatingInput.tsx
│   ├── progress/                   # Progress primitives
│   │   └── ProgressLine.tsx
│   ├── folder/                     # Folder & version management
│   ├── organisms/                  # Atomic-design organisms (some components live here)
│   │   └── generation/
│   │       └── GeneratePanel.tsx   # Re-exported via components/index.ts
│   ├── GeneratePanel.tsx           # Re-export barrel -> organisms/generation/GeneratePanel
│   ├── ProgressPanel.tsx
│   ├── StageCard.tsx
│   ├── ResultPanel.tsx
│   ├── ArtifactPreview.tsx
│   ├── BatchProgressPanel.tsx
│   ├── RefinePanel.tsx
│   ├── ExportModal.tsx
│   ├── DualInputPanel.tsx
│   ├── ImageUpload.tsx
│   ├── Toast.tsx
│   ├── ErrorBoundary.tsx
│   ├── FieldError.tsx
│   ├── WelcomeWizard.tsx
│   ├── ProjectSelector.tsx
│   ├── TemplateSelector.tsx
│   ├── ShortcutsHelpPanel.tsx
│   ├── ConfigPanel.tsx
│   ├── HistorySidebar.tsx          # Legacy (deprecated, use V2 history/)
│   ├── HistoryItem.tsx             # Legacy (deprecated, use V2 history/)
│   ├── theme/                      # Color-scheme controls
│   │   ├── DarkModeToggle.tsx      # Light / Dark / Auto tri-state toggle
│   │   └── DarkModeToggle.test.tsx
│   └── Layout.tsx / Header.tsx / Footer.tsx
├── hooks/                          # Custom React hooks
│   ├── index.ts                    # Barrel exports
│   ├── useGenerate.ts              # Generate lifecycle (SSE streaming)
│   ├── useGenerationFlow.ts        # Generation flow orchestration
│   ├── useGenerationStateMachine.ts # State machine for generation
│   ├── useBatchGeneration.ts       # Batch generation
│   ├── useRefine.ts                # Refinement
│   ├── useHistory.ts               # History management
│   ├── useLanguage.ts              # i18n
│   ├── useProviders.ts             # Provider CRUD
│   ├── useNetworkStatus.ts         # Online/offline detection
│   ├── useFolders.ts               # Folder tree management
│   ├── useVersions.ts              # Version timeline
│   ├── useTemplates.ts             # Template management
│   ├── usePromptTemplates.ts       # Prompt template management
│   ├── useLocalWorkRecords.ts      # Local work record persistence
│   ├── useToast.ts                 # Toast notifications
│   ├── useKeyboardShortcuts.ts     # Keyboard shortcut handler
│   └── useProviderStoreAdapter.ts  # Zustand store adapter (migration)
├── stores/                         # Zustand global state stores
│   ├── index.ts                    # Barrel exports
│   ├── appStore.ts                 # UI state, colorScheme, language, drawers, modals
│   ├── providerStore.ts            # Provider/channel state, role assignments
│   └── generationStore.ts          # Generation state, router, session selection
├── context/                        # React context providers (deprecated)
│   └── deprecated/                 # Migrated to Zustand stores
├── lib/                            # Transport and utility layer
│   ├── api.ts                      # REST client (ApiError, apiClient)
│   ├── sse.ts                      # SSE streaming client (stage events)
│   ├── configStream.ts             # Config SSE subscription
│   ├── batchProgress.ts            # Batch progress reducer
│   ├── refine.ts                   # Refine API client
│   ├── export.ts                   # Export helpers
│   ├── clipboard.ts                # Clipboard utilities
│   ├── imageUtils.ts               # Image processing helpers
│   ├── performance.ts              # Performance monitoring
│   ├── queryClient.ts              # React Query client setup
│   ├── generationStateMachine.ts   # State machine definition
│   ├── refine.test.ts              # Co-located test
│   └── ...                         # Zone.Identifier sidecar files (should not be committed)
├── pages/                          # Page-level composition
│   ├── WorkspacePage.tsx
│   ├── SettingsPage.tsx
│   ├── ProviderEditPage.tsx
│   ├── ProjectsPage.tsx
│   └── VisualizationDetailPage.tsx
├── types/                          # Shared TypeScript types
│   ├── api.ts                      # API request/response contracts
│   └── batch.ts                    # Batch types
├── i18n/                           # Internationalization
│   ├── index.ts                    # i18next setup
│   ├── types.d.ts                  # Type declarations for translation keys
│   └── locales/                    # en.json, zh.json
├── themes/                         # CSS color-scheme system (2 anchors only)
│   ├── tokens.css                  # Tailwind v4 @theme bridge + shared --theme-* tokens
│   ├── claude-light.css            # Claude / Anthropic light anchor (parchment, Source Serif 4)
│   └── linear-dark.css             # Linear dark anchor (#08090a, Inter Variable wght 510)
├── features/                       # Feature modules (evolving structure)
├── utils/                          # Shared utility functions
├── services/                       # Service layer (evolving)
└── test/                           # Test setup and shared suites
    ├── setup.ts                    # Vitest setup (jest-dom matchers)
    ├── golden/                     # Golden UI tests
    └── integration/                # Integration tests (e.g., sse-flow.test.ts)
```

---

## Naming Rules

| Category | Rule | Example |
|----------|------|---------|
| Component files | `PascalCase.tsx` | `GeneratePanel.tsx`, `StageCard.tsx` |
| Component directories | `kebab-case/` or `PascalCase/` (feature dirs) | `workspace/`, `history/`, `model/` |
| Hook files | `camelCase` with `use` prefix | `useGenerate.ts`, `useRefine.ts` |
| Hook names | `use` + `PascalCase` noun/verb | `useGenerate()`, `useHistory()` |
| Lib/utility files | `camelCase` | `api.ts`, `sse.ts`, `refine.ts`, `clipboard.ts` |
| Type files | `camelCase` | `api.ts`, `batch.ts` |
| Test files | Mirror production name + `.test.ts` / `.test.tsx` | `useGenerate.test.ts`, `StageCard.test.tsx` |
| Barrel files | `index.ts` or `index.tsx` | `components/index.ts`, `hooks/index.ts` |
| CSS theme files | `kebab-case.css` | `tokens.css`, `claude-light.css`, `linear-dark.css` |
| Constants | `SCREAMING_SNAKE_CASE` | (within modules, not in filenames) |

---

## Module Organization Rules

### Barrel Exports

Every feature directory uses a barrel `index.ts` to re-export its public API:

```typescript
// web/src/hooks/index.ts
export { useGenerate, type StageState, type GenerateResult } from './useGenerate';
export { useHistory, type HistorySession, type HistoryState } from './useHistory';
export { useRefine, type RefineState } from './useRefine';
```

```typescript
// web/src/components/index.ts
export { GeneratePanel, type GeneratePanelProps } from './GeneratePanel';
export { HistoryPanel, type HistoryPanelProps } from './history';
// Legacy exports explicitly marked
export { HistorySidebar } from './HistorySidebar'; // deprecated
```

**Rules**:
- Barrel files export both values and types.
- Mark legacy/deprecated exports with a comment.
- Do not re-export internal helpers through barrels.

### Co-located Tests

Tests live next to the code they test:

```
lib/
├── refine.ts
└── refine.test.ts        # Co-located

components/
├── StageCard.tsx
└── StageCard.test.tsx    # Co-located
```

Shared test infrastructure lives in `web/src/test/`.

### Feature Directories

Feature domains group related components under a single directory:

```
components/history/
├── index.tsx            # Feature barrel
├── HistoryPanel.tsx
├── HistoryPopover.tsx
├── HistoryItem.tsx
├── HistoryItem.test.tsx
└── HistoryPanel.test.tsx
```

---

## Known Structural Issues

1. **Dual location for some components**: `GeneratePanel.tsx` exists as a re-export barrel pointing to `organisms/generation/GeneratePanel.tsx`. Other components may have similar indirection.
2. **Legacy vs V2 history**: Both `HistorySidebar` (legacy) and `history/HistoryPanel` (V2) are exported from `components/index.ts`. The legacy version is deprecated but still exported.
3. **Zone.Identifier sidecar files**: Windows Zone.Identifier files (e.g., `clipboard.ts:Zone.Identifier`) have been committed to `lib/`. These must be removed and `.gitignore` updated.
4. **Context directory deprecated**: `context/` only contains `deprecated/` -- all state has migrated to Zustand stores. The directory can be removed.
5. **`features/` and `services/` directories**: Partially populated evolving structure. Do not add new code here until the pattern is finalized.

---

## Adding a New Feature Module

When adding a new feature (e.g., "annotations"):

1. Create `web/src/components/annotations/` with an `index.tsx` barrel.
2. Add re-exports to `web/src/components/index.ts`.
3. Create hook(s) in `web/src/hooks/` with `use` prefix.
4. Add hook re-exports to `web/src/hooks/index.ts`.
5. Add types to `web/src/types/` or co-locate within the hook if types are hook-specific.
6. Co-locate `.test.tsx` files with each component.
7. Add i18n keys to `web/src/i18n/locales/en.json` and `zh.json`.

---

## Pre-Tauri Migration Notes

The following directories and files will be most affected by the Tauri migration:

| Area | Current | Post-Tauri |
|------|---------|------------|
| `lib/api.ts` | `fetch()` to `:8080` | Tauri `invoke()` commands |
| `lib/sse.ts` | Browser `EventSource` | Tauri event listeners |
| `lib/configStream.ts` | SSE subscription | Tauri event channel |
| `lib/refine.ts` | `fetch()` to `:8080` | Tauri `invoke()` commands |
| `main.tsx` | Browser `createRoot` | Tauri webview window |
| `App.tsx` | SPA routing | Tauri window management |

The component tree, hooks, stores, types, and i18n will remain largely unchanged.
