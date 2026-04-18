# State Management

> How state is managed in the PaperBanana frontend.

---

## Overview

The project uses a **hook-and-context model** that is transitioning to **Zustand stores** for global state. There is no Redux, MobX, or other heavy state library. React Query manages server-cached state.

> **[PRE-TAURI]**: State management will remain the same after the Tauri migration. Only the transport layer changes (fetch -> Tauri IPC). The hooks and stores that consume transport data will need their transport calls updated but their state shape will not change.

---

## State Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Zustand Global Stores                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  appStore     │  │ providerStore│  │ generationStore  │  │
│  │  - UI state   │  │ - providers  │  │ - session state  │  │
│  │  - theme      │  │ - channels   │  │ - router         │  │
│  │  - language   │  │ - roles      │  │ - selection      │  │
│  │  - drawers    │  │ - models     │  │                  │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│         │                  │                   │             │
│         └──────────────────┴───────────────────┘             │
│                            │                                 │
│                    React components                          │
│                            │                                 │
│         ┌──────────────────┴───────────────────┐             │
│         │                                      │             │
│  ┌──────────────┐                    ┌──────────────────┐   │
│  │ Custom Hooks │                    │ React Query      │   │
│  │ (local state)│                    │ (server cache)   │   │
│  │              │                    │                  │   │
│  │ useGenerate  │                    │ queryClient      │   │
│  │ useRefine    │                    │ fetch-on-demand  │   │
│  │ useHistory   │                    │ background refetch│   │
│  └──────────────┘                    └──────────────────┘   │
│                            │                                 │
│                    Transport layer                           │
│         ┌──────────────────┴───────────────────┐             │
│         │                                      │             │
│  ┌──────────────┐                    ┌──────────────────┐   │
│  │  REST (api)  │                    │  SSE streaming   │   │
│  └──────────────┘                    └──────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## Decision Framework

When deciding where to put state, use this hierarchy:

### Level 1: Local `useState` (Default)

Use for component-local UI concerns that no other component needs:

```typescript
const [isExpanded, setIsExpanded] = useState(false);
const [inputValue, setInputValue] = useState('');
```

**When**: State is created and consumed entirely within one component.

### Level 2: Custom Hooks (Async Workflows and Derived State)

Use for multi-step async workflows, derived state, and state that spans a few related components:

```typescript
// useGenerate.ts - manages the full generation lifecycle
export function useGenerate() {
  const [state, setState] = useState<GenerateState>(createInitialState);
  // ... SSE streaming, stage tracking, error handling
  return { ...state, generate, cancel, reset, restore };
}
```

**When**:
- Multi-step async process (generate, refine, batch)
- State needs cleanup/reset logic
- Multiple related state fields change together
- State needs to be shared by a small component tree (prop drilling through 2-3 levels is acceptable)

### Level 3: Zustand Stores (Broadly Shared Global State)

Use for state that many unrelated components across the app need:

```typescript
// stores/appStore.ts
export const useAppStore = create<UIState & UIActions>()(
  persist(
    (set) => ({
      theme: 'qi-baishi',
      language: 'en',
      currentPage: 'workspace',
      // ... actions
      setTheme: (theme) => set({ theme }),
    }),
    { name: 'paperbanana-ui' }
  )
);
```

**When**:
- State is needed by components in different feature trees
- State must survive component unmount/remount
- State needs localStorage persistence
- Multiple independent components read and write the same state

### Level 4: React Query (Server-Cached State)

Use for data fetched from the backend that benefits from caching, background refetching, and stale-while-revalidate:

```typescript
// Through useProviders hook or direct useQuery calls
const { data: providers } = useQuery({
  queryKey: ['providers'],
  queryFn: fetchProviders,
});
```

**When**:
- Data originates from the backend
- Multiple components need the same server data
- You want automatic background refetching
- You need loading/error states for server data

---

## Current Store Inventory

### Zustand Stores (in `web/src/stores/`)

| Store | File | Responsibility |
|-------|------|---------------|
| `useAppStore` | `appStore.ts` | UI state: theme, language, page, drawers, modals |
| `useProviderStore` | `providerStore.ts` | Provider/channel state, role assignments, model snapshots |
| `useGenerationStore` | `generationStore.ts` | Generation state, router, session selection, export state |

### Custom Hooks with Significant State (in `web/src/hooks/`)

| Hook | File | State Shape |
|------|------|-------------|
| `useGenerate` | `useGenerate.ts` | `GenerateState` (isGenerating, stages, result, error, resumeMetadata) |
| `useRefine` | `useRefine.ts` | `RefineState` (isRefining, result, error) |
| `useHistory` | `useHistory.ts` | `HistoryState` (sessions, loading, error) |
| `useGenerationFlow` | `useGenerationFlow.ts` | Orchestrates generate/refine/restore |
| `useGenerationStateMachine` | `useGenerationStateMachine.ts` | State machine for generation lifecycle |
| `useBatchGeneration` | `useBatchGeneration.ts` | Batch generation progress |

---

## Persistence Strategy

| Data | Storage | Mechanism |
|------|---------|-----------|
| Theme preference | localStorage | Zustand `persist` middleware |
| Language preference | localStorage | Zustand `persist` middleware |
| Model/provider config | localStorage | Zustand `persist` middleware |
| Generation state | Memory only | Lost on page refresh (intentional) |
| History | Backend | Fetched via API on demand |
| Local work records | localStorage | `useLocalWorkRecords` hook |

### Config State Flow

The config state has a three-phase load sequence that is a **fragile area**:

```
1. localStorage (instant)          -> optimistic UI render
2. Backend fetch via API           -> authoritative values overwrite
3. SSE configStream subscription   -> live updates push changes
```

If the backend is unreachable, the UI falls back to localStorage values. If the backend returns different values than localStorage, the UI flickers as the authoritative values overwrite the optimistic ones.

**Mitigation**: When working on config-related components, always test the three-phase load order. Consider showing a loading state until the backend fetch completes.

---

## Forbidden Patterns

| Pattern | Reason |
|---------|--------|
| Redux / MobX / Vuex | Project uses Zustand for global state. Adding another state library increases bundle size and cognitive load. |
| Prop drilling beyond 3 levels | If you are passing a prop through 3+ intermediate components, lift the state to a Zustand store or React context. |
| Storing server data in Zustand | Use React Query for server-cached data. Zustand is for client-only state. |
| Duplicating state across hook + store | Do not keep the same data in both a custom hook and a Zustand store. Pick one. |
| `useEffect` for derived state | Use `useMemo` for computed values. `useEffect` is for side effects, not data derivation. |
| Direct `localStorage` access outside stores/hooks | Centralize persistence through Zustand persist middleware or dedicated hooks. |

---

## Migration Status

The project is migrating from a pure hook-and-context model to Zustand stores:

| Area | Status | Notes |
|------|--------|-------|
| UI state (theme, language, page) | **Migrated** | `appStore.ts` with persist |
| Provider/channel state | **Migrated** | `providerStore.ts` |
| Generation state | **Migrated** | `generationStore.ts` |
| `ModelConfigContext` | **Deprecated** | Replaced by `providerStore.ts`, files in `context/deprecated/` |
| Adapter hooks | **Bridge** | `useAppStoreAdapter.ts`, `useProviderStoreAdapter.ts` wrap store access for gradual migration |
| Generate/refine hooks | **Not migrated** | Still use `useState` internally. No plan to move these to Zustand -- they are workflow-scoped, not global. |

**When adding new global state**: Put it in the appropriate Zustand store. Do not create new React contexts or hook-based global state.

---

## Pre-Tauri Considerations

After the Tauri migration, the transport layer changes but the state architecture stays the same:

| Current | Post-Tauri |
|---------|------------|
| `fetch('/api/v1/...')` | `invoke('api_command', { ... })` |
| `EventSource('/api/v1/sse')` | `listen('sse-event', handler)` |
| `fetch` error -> `ApiError` | Tauri command error -> typed error |

The Zustand stores, custom hooks, and React Query layer will only need their data-fetching functions updated. The state shapes and component APIs remain unchanged.
