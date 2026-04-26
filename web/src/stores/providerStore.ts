import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

// ============================================================================
// Types (migrated from deprecated ModelConfigContext)
// ============================================================================

export interface ModelInfo {
  id: string;
  name: string;
  max_tokens?: number;
  supports_vision?: boolean;
  enabled: boolean;
}

export interface Provider {
  id: string;
  name: string;
  type: 'openai' | 'anthropic' | 'gemini' | 'openrouter' | 'deepseek' | 'zhipu' | 'moonshot' | 'qwen' | 'grok' | 'xai' | 'custom';
  display_name: string;
  query_model?: string;
  gen_model?: string;
  base_url?: string;
  api_key?: string;
  timeout: number;
  enabled: boolean;
  models: ModelInfo[];
  last_fetched?: string;
  is_system?: boolean;
  is_default?: boolean;
  status: 'configured' | 'no_keys' | 'invalid' | 'pending';
}

/** @deprecated Use Provider instead */
export type Channel = Provider;

export type WorkflowRole = 'image_generation' | 'retrieval_reasoning';

export interface RoleAssignment {
  role: WorkflowRole;
  provider_id: string;
  model_id: string;
  provider_name: string;
  model_name: string;
}

export interface ModelSnapshot {
  timestamp: string;
  provider_assignments: Record<WorkflowRole, RoleAssignment>;
}

export interface ProviderState {
  providers: Provider[];
  role_assignments: Record<WorkflowRole, RoleAssignment | null>;
  snapshots: Record<string, ModelSnapshot>;
  loading: boolean;
  error: string | null;
}

export interface ProviderActions {
  // Provider operations
  setProviders: (providers: Provider[]) => void;
  addProvider: (provider: Provider) => void;
  updateProvider: (id: string, updates: Partial<Provider>) => void;
  deleteProvider: (id: string) => void;
  
  // Model operations
  setProviderModels: (provider_id: string, models: ModelInfo[]) => void;
  toggleModel: (provider_id: string, model_id: string, enabled: boolean) => void;
  addModel: (provider_id: string, model: ModelInfo) => void;
  removeModel: (provider_id: string, model_id: string) => void;
  
  // Role operations
  setRoleAssignment: (assignment: RoleAssignment) => void;
  clearRoleAssignment: (role: WorkflowRole) => void;
  
  // Snapshot operations
  saveSnapshot: (task_id: string, snapshot: ModelSnapshot) => void;
  getSnapshot: (task_id: string) => ModelSnapshot | null;
  
  // Loading & Error
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // Hydration
  hydrateState: (state: Partial<ProviderState>) => void;
  
  // Computed
  getEnabledModels: () => Array<{ provider: Provider; model: ModelInfo }>;
  getProvider: (id: string) => Provider | undefined;
  getRoleAssignment: (role: WorkflowRole) => RoleAssignment | null;
  
  // Reset
  resetProviderState: () => void;
  
  // Backward compatibility aliases
  /** @deprecated Use providers instead */
  channels: Provider[];
  /** @deprecated Use setProviders instead */
  setChannels: (channels: Provider[]) => void;
}

// ============================================================================
// Utility Functions
// ============================================================================

function parseTimeout(timeout: string): number {
  if (!timeout) return 60;
  const match = timeout.match(/^(\d+)(ms|s)?$/);
  if (!match) return 60;
  const value = parseInt(match[1], 10);
  if (match[2] === 'ms') {
    return Math.ceil(value / 1000);
  }
  return value;
}

// ============================================================================
// Initial State
// ============================================================================

const initialState: ProviderState = {
  providers: [],
  role_assignments: {
    image_generation: null,
    retrieval_reasoning: null,
  },
  snapshots: {},
  loading: false,
  error: null,
};

// ============================================================================
// Store
// ============================================================================

export const useProviderStore = create<ProviderState & ProviderActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      // Provider operations
      setProviders: (providers) => set({ providers }),
      
      addProvider: (provider) =>
        set((state) => ({
          providers: [...state.providers, provider],
        })),
      
      updateProvider: (id, updates) =>
        set((state) => ({
          providers: state.providers.map((p) =>
            p.id === id ? { ...p, ...updates } : p
          ),
        })),
      
      deleteProvider: (id) =>
        set((state) => ({
          providers: state.providers.filter((p) => p.id !== id),
          role_assignments: {
            image_generation:
              state.role_assignments.image_generation?.provider_id === id
                ? null
                : state.role_assignments.image_generation,
            retrieval_reasoning:
              state.role_assignments.retrieval_reasoning?.provider_id === id
                ? null
                : state.role_assignments.retrieval_reasoning,
          },
        })),

      // Model operations
      setProviderModels: (provider_id, models) =>
        set((state) => ({
          providers: state.providers.map((p) =>
            p.id === provider_id
              ? { ...p, models, last_fetched: new Date().toISOString() }
              : p
          ),
        })),
      
      toggleModel: (provider_id, model_id, enabled) =>
        set((state) => ({
          providers: state.providers.map((p) =>
            p.id === provider_id
              ? {
                  ...p,
                  models: p.models.map((m) =>
                    m.id === model_id ? { ...m, enabled } : m
                  ),
                }
              : p
          ),
        })),
      
      addModel: (provider_id, model) =>
        set((state) => ({
          providers: state.providers.map((p) =>
            p.id === provider_id
              ? { ...p, models: [...p.models, model] }
              : p
          ),
        })),
      
      removeModel: (provider_id, model_id) =>
        set((state) => ({
          providers: state.providers.map((p) =>
            p.id === provider_id
              ? { ...p, models: p.models.filter((m) => m.id !== model_id) }
              : p
          ),
        })),

      // Role operations
      setRoleAssignment: (assignment) =>
        set((state) => ({
          role_assignments: {
            ...state.role_assignments,
            [assignment.role]: assignment,
          },
        })),
      
      clearRoleAssignment: (role) =>
        set((state) => ({
          role_assignments: {
            ...state.role_assignments,
            [role]: null,
          },
        })),

      // Snapshot operations
      saveSnapshot: (task_id, snapshot) =>
        set((state) => ({
          snapshots: {
            ...state.snapshots,
            [task_id]: snapshot,
          },
        })),
      
      getSnapshot: (task_id) => {
        return get().snapshots[task_id] || null;
      },

      // Loading & Error
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),

      // Hydration
      hydrateState: (newState) =>
        set((state) => ({
          ...state,
          ...newState,
        })),

      // Computed
      getEnabledModels: () => {
        const result: Array<{ provider: Provider; model: ModelInfo }> = [];
        for (const provider of get().providers) {
          if (!provider.enabled) continue;
          for (const model of provider.models) {
            if (model.enabled) {
              result.push({ provider, model });
            }
          }
        }
        return result;
      },
      
      getProvider: (id) => {
        return get().providers.find((p) => p.id === id);
      },
      
      getRoleAssignment: (role) => {
        return get().role_assignments[role];
      },

      // Reset
      resetProviderState: () => set(initialState),

      // Backward compatibility - alias channels to providers
      get channels() {
        return get().providers;
      },
      setChannels: (channels) => set({ providers: channels }),
    }),
    {
      name: 'paperbanana-provider-store-v1',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        providers: state.providers.map(({ api_key: _key, ...rest }) => rest),
        role_assignments: state.role_assignments,
        snapshots: state.snapshots,
      }),
    }
  )
);

// ============================================================================
// Selectors
// ============================================================================

export const selectProviders = (state: ProviderState & ProviderActions) => state.providers;
export const selectRoleAssignments = (state: ProviderState & ProviderActions) => state.role_assignments;
export const selectLoading = (state: ProviderState & ProviderActions) => state.loading;
export const selectError = (state: ProviderState & ProviderActions) => state.error;
export const selectSnapshots = (state: ProviderState & ProviderActions) => state.snapshots;

// ============================================================================
// Async Actions (API integration)
// ============================================================================

export async function fetchProviders(): Promise<Provider[]> {
  const response = await fetch('/api/v1/providers');
  if (!response.ok) throw new Error('Failed to fetch providers');
  const data = await response.json();
  
  return (data.providers || []).map((p: {
    id: string;
    type: string;
    name: string;
    display_name: string;
    query_model?: string;
    gen_model?: string;
    base_url?: string;
    timeout?: string;
    enabled: boolean;
    status: string;
    is_system?: boolean;
    is_default?: boolean;
    models?: ModelInfo[];
  }) => ({
    id: p.id,
    name: p.name,
    type: p.type as Provider['type'],
    display_name: p.display_name,
    query_model: p.query_model,
    gen_model: p.gen_model,
    base_url: p.base_url,
    timeout: typeof p.timeout === 'string' ? parseTimeout(p.timeout) : 60,
    enabled: p.enabled,
    models: (p.models || []).map(m => ({ ...m, enabled: m.enabled ?? true })),
    is_system: p.is_system,
    is_default: p.is_default,
    status: p.status as Provider['status'],
  }));
}

export async function createProvider(provider: Omit<Provider, 'id' | 'models'>): Promise<Provider> {
  const providerData = {
    type: provider.type,
    name: provider.name,
    display_name: provider.display_name,
    api_host: provider.base_url,
    api_key: provider.api_key,
    timeout_ms: provider.timeout * 1000,
    enabled: provider.enabled,
  };
  
  const response = await fetch('/api/v1/providers', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(providerData),
  });
  
  if (!response.ok) throw new Error('Failed to add provider');
  const data = await response.json();
  
  return {
    id: data.provider.id,
    name: data.provider.name,
    type: data.provider.type as Provider['type'],
    display_name: data.provider.display_name,
    query_model: data.provider.query_model,
    gen_model: data.provider.gen_model,
    base_url: data.provider.api_host,
    timeout: data.provider.timeout_ms / 1000,
    enabled: data.provider.enabled,
    models: data.provider.models || [],
    is_system: data.provider.is_system,
    is_default: data.provider.is_default,
    status: data.provider.status || 'configured',
  };
}

export async function updateProviderAPI(id: string, updates: Partial<Provider>): Promise<void> {
  const providerUpdates: Record<string, unknown> = {};
  if (updates.display_name !== undefined) providerUpdates.display_name = updates.display_name;
  if (updates.base_url !== undefined) providerUpdates.api_host = updates.base_url;
  if (updates.timeout !== undefined) providerUpdates.timeout_ms = updates.timeout * 1000;
  if (updates.enabled !== undefined) providerUpdates.enabled = updates.enabled;
  if (updates.models !== undefined) providerUpdates.models = updates.models;

  const response = await fetch(`/api/v1/providers/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(providerUpdates),
  });
  
  if (!response.ok) throw new Error('Failed to update provider');
}

export async function deleteProviderAPI(id: string): Promise<void> {
  const response = await fetch(`/api/v1/providers/${id}`, { method: 'DELETE' });
  if (!response.ok) throw new Error('Failed to delete provider');
}

export async function fetchProviderModelsAPI(provider_id: string): Promise<ModelInfo[]> {
  const response = await fetch(`/api/v1/providers/${provider_id}/models`);
  if (!response.ok) throw new Error('Failed to fetch models');
  const data = await response.json();
  
  return (data.models || []).map((m: {
    id: string;
    name: string;
    max_tokens?: number;
    supports_vision?: boolean;
  }) => ({
    id: m.id,
    name: m.name,
    max_tokens: m.max_tokens,
    supports_vision: m.supports_vision,
    enabled: true,
  }));
}

export async function assignRoleAPI(
  role: WorkflowRole, 
  provider_id: string, 
  model_id: string
): Promise<void> {
  const updateKey = role === 'image_generation' ? 'gen_model' : 'query_model';
  const response = await fetch(`/api/v1/providers/${provider_id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ [updateKey]: model_id }),
  });
  if (!response.ok) throw new Error('Failed to assign role');
}
