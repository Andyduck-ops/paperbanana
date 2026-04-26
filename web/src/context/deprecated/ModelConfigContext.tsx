import React, { createContext, useContext, useReducer, useEffect, useCallback } from 'react';
import { subscribeToConfigChanges } from '../../lib/configStream';

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Parse timeout string like "60s" or "500ms" to seconds
 */
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
// Types
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

export interface ModelConfigState {
  providers: Provider[];
  role_assignments: Record<WorkflowRole, RoleAssignment | null>;
  snapshots: Record<string, ModelSnapshot>;
  loading: boolean;
  error: string | null;
}

/** @deprecated Use ModelConfigState instead */
export type ChannelConfigState = ModelConfigState;

// ============================================================================
// Actions
// ============================================================================

type ModelConfigAction =
  | { type: 'SET_PROVIDERS'; payload: Provider[] }
  | { type: 'ADD_PROVIDER'; payload: Provider }
  | { type: 'UPDATE_PROVIDER'; payload: { id: string; updates: Partial<Provider> } }
  | { type: 'DELETE_PROVIDER'; payload: string }
  | { type: 'SET_PROVIDER_MODELS'; payload: { provider_id: string; models: ModelInfo[] } }
  | { type: 'TOGGLE_MODEL'; payload: { provider_id: string; model_id: string; enabled: boolean } }
  | { type: 'ADD_MODEL'; payload: { provider_id: string; model: ModelInfo } }
  | { type: 'REMOVE_MODEL'; payload: { provider_id: string; model_id: string } }
  | { type: 'SET_ROLE_ASSIGNMENT'; payload: RoleAssignment }
  | { type: 'CLEAR_ROLE_ASSIGNMENT'; payload: WorkflowRole }
  | { type: 'SAVE_SNAPSHOT'; payload: { task_id: string; snapshot: ModelSnapshot } }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'HYDRATE_STATE'; payload: ModelConfigState };

// ============================================================================
// Reducer
// ============================================================================

const initialState: ModelConfigState = {
  providers: [],
  role_assignments: {
    image_generation: null,
    retrieval_reasoning: null,
  },
  snapshots: {},
  loading: false,
  error: null,
};

function modelConfigReducer(state: ModelConfigState, action: ModelConfigAction): ModelConfigState {
  switch (action.type) {
    case 'SET_PROVIDERS':
      return { ...state, providers: action.payload };

    case 'ADD_PROVIDER':
      return {
        ...state,
        providers: [...state.providers, action.payload],
      };

    case 'UPDATE_PROVIDER':
      return {
        ...state,
        providers: state.providers.map((p) =>
          p.id === action.payload.id ? { ...p, ...action.payload.updates } : p
        ),
      };

    case 'DELETE_PROVIDER':
      return {
        ...state,
        providers: state.providers.filter((p) => p.id !== action.payload),
        role_assignments: {
          image_generation:
            state.role_assignments.image_generation?.provider_id === action.payload
              ? null
              : state.role_assignments.image_generation,
          retrieval_reasoning:
            state.role_assignments.retrieval_reasoning?.provider_id === action.payload
              ? null
              : state.role_assignments.retrieval_reasoning,
        },
      };

    case 'SET_PROVIDER_MODELS':
      return {
        ...state,
        providers: state.providers.map((p) =>
          p.id === action.payload.provider_id
            ? { ...p, models: action.payload.models, last_fetched: new Date().toISOString() }
            : p
        ),
      };

    case 'TOGGLE_MODEL':
      return {
        ...state,
        providers: state.providers.map((p) =>
          p.id === action.payload.provider_id
            ? {
                ...p,
                models: p.models.map((m) =>
                  m.id === action.payload.model_id ? { ...m, enabled: action.payload.enabled } : m
                ),
              }
            : p
        ),
      };

    case 'ADD_MODEL':
      return {
        ...state,
        providers: state.providers.map((p) =>
          p.id === action.payload.provider_id
            ? { ...p, models: [...p.models, action.payload.model] }
            : p
        ),
      };

    case 'REMOVE_MODEL':
      return {
        ...state,
        providers: state.providers.map((p) =>
          p.id === action.payload.provider_id
            ? { ...p, models: p.models.filter((m) => m.id !== action.payload.model_id) }
            : p
        ),
      };

    case 'SET_ROLE_ASSIGNMENT':
      return {
        ...state,
        role_assignments: {
          ...state.role_assignments,
          [action.payload.role]: action.payload,
        },
      };

    case 'CLEAR_ROLE_ASSIGNMENT':
      return {
        ...state,
        role_assignments: {
          ...state.role_assignments,
          [action.payload]: null,
        },
      };

    case 'SAVE_SNAPSHOT':
      return {
        ...state,
        snapshots: {
          ...state.snapshots,
          [action.payload.task_id]: action.payload.snapshot,
        },
      };

    case 'SET_LOADING':
      return { ...state, loading: action.payload };

    case 'SET_ERROR':
      return { ...state, error: action.payload };

    case 'HYDRATE_STATE':
      return { ...action.payload };

    default:
      return state;
  }
}

// ============================================================================
// Context
// ============================================================================

interface ModelConfigContextValue extends ModelConfigState {
  // Provider operations
  addProvider: (provider: Omit<Provider, 'id' | 'models'>) => Promise<void>;
  updateProvider: (id: string, updates: Partial<Provider>) => Promise<void>;
  deleteProvider: (id: string) => Promise<void>;
  fetchProviderModels: (provider_id: string) => Promise<void>;

  // Model operations
  toggleModel: (provider_id: string, model_id: string, enabled: boolean) => Promise<void>;
  addModel: (provider_id: string, model: Omit<ModelInfo, 'id'>) => Promise<void>;
  removeModel: (provider_id: string, model_id: string) => Promise<void>;

  // Role operations
  assignRole: (role: WorkflowRole, provider_id: string, model_id: string) => Promise<void>;
  clearRole: (role: WorkflowRole) => void;

  // Snapshot operations
  saveSnapshot: (task_id: string) => void;
  getSnapshot: (task_id: string) => ModelSnapshot | null;

  // Utility
  refreshProviders: () => Promise<void>;
  getEnabledModels: () => Array<{ provider: Provider; model: ModelInfo }>;
  getRoleDisplayName: (role: WorkflowRole) => string;

  // Backward compatibility aliases
  /** @deprecated Use providers instead */
  channels: Provider[];
  /** @deprecated Use addProvider instead */
  addChannel: (provider: Omit<Provider, 'id' | 'models'>) => Promise<void>;
  /** @deprecated Use updateProvider instead */
  updateChannel: (id: string, updates: Partial<Provider>) => Promise<void>;
  /** @deprecated Use deleteProvider instead */
  deleteChannel: (id: string) => Promise<void>;
  /** @deprecated Use fetchProviderModels instead */
  fetchChannelModels: (provider_id: string) => Promise<void>;
  /** @deprecated Use refreshProviders instead */
  refreshChannels: () => Promise<void>;
}

const ModelConfigContext = createContext<ModelConfigContextValue | null>(null);

// ============================================================================
// Provider
// ============================================================================

const STORAGE_KEY = 'paperbanana_model_config_v1';

export function ModelConfigProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(modelConfigReducer, initialState);

  // API Operations - Uses /api/v1/providers endpoint
  const refreshProviders = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const response = await fetch('/api/v1/providers');
      if (!response.ok) throw new Error('Failed to fetch providers');
      const data = await response.json();
      // Map provider response to provider format
      const providers: Provider[] = (data.providers || []).map((p: {
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
      dispatch({ type: 'SET_PROVIDERS', payload: providers });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  // Backward compatibility alias
  const refreshChannels = refreshProviders;

  // Load from localStorage on mount, then fetch fresh data from backend
  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        dispatch({ type: 'HYDRATE_STATE', payload: { ...initialState, ...parsed } });
      } catch (e) {
        console.error('Failed to parse stored model config:', e);
      }
    }

    // Always fetch fresh data from backend on mount
    refreshProviders();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    return subscribeToConfigChanges(() => {
      void refreshProviders();
    });
  }, [refreshProviders]);

  // Save to localStorage on state change
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  }, [state]);

  const addProvider = useCallback(async (provider: Omit<Provider, 'id' | 'models'>) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
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
      const newProvider: Provider = {
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
      dispatch({ type: 'ADD_PROVIDER', payload: newProvider });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  // Backward compatibility alias
  const addChannel = addProvider;

  const updateProvider = useCallback(async (id: string, updates: Partial<Provider>) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
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
      dispatch({ type: 'UPDATE_PROVIDER', payload: { id, updates } });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  // Backward compatibility alias
  const updateChannel = updateProvider;

  const deleteProvider = useCallback(async (id: string) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const response = await fetch(`/api/v1/providers/${id}`, { method: 'DELETE' });
      if (!response.ok) throw new Error('Failed to delete provider');
      dispatch({ type: 'DELETE_PROVIDER', payload: id });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  // Backward compatibility alias
  const deleteChannel = deleteProvider;

  const fetchProviderModels = useCallback(async (provider_id: string) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const response = await fetch(`/api/v1/providers/${provider_id}/models`);
      if (!response.ok) throw new Error('Failed to fetch models');
      const data = await response.json();
      const models: ModelInfo[] = (data.models || []).map((m: {
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
      dispatch({
        type: 'SET_PROVIDER_MODELS',
        payload: { provider_id, models },
      });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  // Backward compatibility alias
  const fetchChannelModels = fetchProviderModels;

  const toggleModel = useCallback(async (provider_id: string, model_id: string, enabled: boolean) => {
    try {
      const provider = state.providers.find(p => p.id === provider_id);
      if (!provider) return;

      const updatedModels = provider.models.map(m =>
        m.id === model_id ? { ...m, enabled } : m
      );

      const response = await fetch(`/api/v1/providers/${provider_id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ models: updatedModels }),
      });
      if (!response.ok) throw new Error('Failed to toggle model');
      dispatch({ type: 'TOGGLE_MODEL', payload: { provider_id, model_id, enabled } });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    }
  }, [state.providers]);

  const addModel = useCallback(async (provider_id: string, model: Omit<ModelInfo, 'id'>) => {
    try {
      const provider = state.providers.find(p => p.id === provider_id);
      if (!provider) return;

      const newModel: ModelInfo = {
        ...model,
        id: model.name.toLowerCase().replace(/\s+/g, '-'),
      };
      const updatedModels = [...provider.models, newModel];

      const response = await fetch(`/api/v1/providers/${provider_id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ models: updatedModels }),
      });
      if (!response.ok) throw new Error('Failed to add model');
      dispatch({ type: 'ADD_MODEL', payload: { provider_id, model: newModel } });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    }
  }, [state.providers]);

  const removeModel = useCallback(async (provider_id: string, model_id: string) => {
    try {
      const provider = state.providers.find(p => p.id === provider_id);
      if (!provider) return;

      const updatedModels = provider.models.filter(m => m.id !== model_id);

      const response = await fetch(`/api/v1/providers/${provider_id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ models: updatedModels }),
      });
      if (!response.ok) throw new Error('Failed to remove model');
      dispatch({ type: 'REMOVE_MODEL', payload: { provider_id, model_id } });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    }
  }, [state.providers]);

  const assignRole = useCallback(
    async (role: WorkflowRole, provider_id: string, model_id: string) => {
      const provider = state.providers.find((p) => p.id === provider_id);
      const model = provider?.models.find((m) => m.id === model_id);
      if (!provider || !model) return;

      const assignment: RoleAssignment = {
        role,
        provider_id,
        model_id,
        provider_name: provider.display_name,
        model_name: model.name,
      };

      try {
        // Update provider with the model assignment
        const updateKey = role === 'image_generation' ? 'gen_model' : 'query_model';
        const response = await fetch(`/api/v1/providers/${provider_id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ [updateKey]: model_id }),
        });
        if (!response.ok) throw new Error('Failed to assign role');
        dispatch({ type: 'SET_ROLE_ASSIGNMENT', payload: assignment });
      } catch (error) {
        dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
      }
    },
    [state.providers]
  );

  const clearRole = useCallback((role: WorkflowRole) => {
    dispatch({ type: 'CLEAR_ROLE_ASSIGNMENT', payload: role });
  }, []);

  const saveSnapshot = useCallback(
    (task_id: string) => {
      const snapshot: ModelSnapshot = {
        timestamp: new Date().toISOString(),
        provider_assignments: {
          image_generation: state.role_assignments.image_generation!,
          retrieval_reasoning: state.role_assignments.retrieval_reasoning!,
        },
      };
      dispatch({ type: 'SAVE_SNAPSHOT', payload: { task_id, snapshot } });
    },
    [state.role_assignments]
  );

  const getSnapshot = useCallback(
    (task_id: string) => {
      return state.snapshots[task_id] || null;
    },
    [state.snapshots]
  );

  const getEnabledModels = useCallback(() => {
    const result: Array<{ provider: Provider; model: ModelInfo }> = [];
    for (const provider of state.providers) {
      if (!provider.enabled) continue;
      for (const model of provider.models) {
        if (model.enabled) {
          result.push({ provider, model });
        }
      }
    }
    return result;
  }, [state.providers]);

  const getRoleDisplayName = useCallback((role: WorkflowRole) => {
    const names: Record<WorkflowRole, string> = {
      image_generation: 'Image Generation',
      retrieval_reasoning: 'Retrieval & Reasoning',
    };
    return names[role];
  }, []);

  const value: ModelConfigContextValue = {
    ...state,
    // New Provider-based API
    providers: state.providers,
    addProvider,
    updateProvider,
    deleteProvider,
    fetchProviderModels,
    refreshProviders,
    // Backward compatible aliases
    channels: state.providers,
    addChannel,
    updateChannel,
    deleteChannel,
    fetchChannelModels,
    refreshChannels,
    // Common operations
    toggleModel,
    addModel,
    removeModel,
    assignRole,
    clearRole,
    saveSnapshot,
    getSnapshot,
    getEnabledModels,
    getRoleDisplayName,
  };

  return <ModelConfigContext.Provider value={value}>{children}</ModelConfigContext.Provider>;
}

// ============================================================================
// Hook
// ============================================================================

export function useModelConfig(): ModelConfigContextValue {
  const context = useContext(ModelConfigContext);
  if (!context) {
    throw new Error('useModelConfig must be used within a ModelConfigProvider');
  }
  return context;
}

export function useProvider(providerId: string): Provider | undefined {
  const { providers } = useModelConfig();
  return providers.find((p) => p.id === providerId);
}

/** @deprecated Use useProvider instead */
export function useChannel(channelId: string): Provider | undefined {
  return useProvider(channelId);
}

export function useRoleAssignment(role: WorkflowRole): RoleAssignment | null {
  const { role_assignments } = useModelConfig();
  return role_assignments[role];
}
