import React, { createContext, useContext, useReducer, useEffect, useCallback } from 'react';
import { subscribeToConfigChanges } from '../lib/configStream';

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

export interface Channel {
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

export type WorkflowRole = 'image_generation' | 'retrieval_reasoning';

export interface RoleAssignment {
  role: WorkflowRole;
  channel_id: string;
  model_id: string;
  channel_name: string;
  model_name: string;
}

export interface ModelSnapshot {
  timestamp: string;
  channel_assignments: Record<WorkflowRole, RoleAssignment>;
}

export interface ModelConfigState {
  channels: Channel[];
  role_assignments: Record<WorkflowRole, RoleAssignment | null>;
  snapshots: Record<string, ModelSnapshot>;
  loading: boolean;
  error: string | null;
}

// ============================================================================
// Actions
// ============================================================================

type ModelConfigAction =
  | { type: 'SET_CHANNELS'; payload: Channel[] }
  | { type: 'ADD_CHANNEL'; payload: Channel }
  | { type: 'UPDATE_CHANNEL'; payload: { id: string; updates: Partial<Channel> } }
  | { type: 'DELETE_CHANNEL'; payload: string }
  | { type: 'SET_CHANNEL_MODELS'; payload: { channel_id: string; models: ModelInfo[] } }
  | { type: 'TOGGLE_MODEL'; payload: { channel_id: string; model_id: string; enabled: boolean } }
  | { type: 'ADD_MODEL'; payload: { channel_id: string; model: ModelInfo } }
  | { type: 'REMOVE_MODEL'; payload: { channel_id: string; model_id: string } }
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
  channels: [],
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
    case 'SET_CHANNELS':
      return { ...state, channels: action.payload };

    case 'ADD_CHANNEL':
      return {
        ...state,
        channels: [...state.channels, action.payload],
      };

    case 'UPDATE_CHANNEL':
      return {
        ...state,
        channels: state.channels.map((ch) =>
          ch.id === action.payload.id ? { ...ch, ...action.payload.updates } : ch
        ),
      };

    case 'DELETE_CHANNEL':
      return {
        ...state,
        channels: state.channels.filter((ch) => ch.id !== action.payload),
        role_assignments: {
          image_generation:
            state.role_assignments.image_generation?.channel_id === action.payload
              ? null
              : state.role_assignments.image_generation,
          retrieval_reasoning:
            state.role_assignments.retrieval_reasoning?.channel_id === action.payload
              ? null
              : state.role_assignments.retrieval_reasoning,
        },
      };

    case 'SET_CHANNEL_MODELS':
      return {
        ...state,
        channels: state.channels.map((ch) =>
          ch.id === action.payload.channel_id
            ? { ...ch, models: action.payload.models, last_fetched: new Date().toISOString() }
            : ch
        ),
      };

    case 'TOGGLE_MODEL':
      return {
        ...state,
        channels: state.channels.map((ch) =>
          ch.id === action.payload.channel_id
            ? {
                ...ch,
                models: ch.models.map((m) =>
                  m.id === action.payload.model_id ? { ...m, enabled: action.payload.enabled } : m
                ),
              }
            : ch
        ),
      };

    case 'ADD_MODEL':
      return {
        ...state,
        channels: state.channels.map((ch) =>
          ch.id === action.payload.channel_id
            ? { ...ch, models: [...ch.models, action.payload.model] }
            : ch
        ),
      };

    case 'REMOVE_MODEL':
      return {
        ...state,
        channels: state.channels.map((ch) =>
          ch.id === action.payload.channel_id
            ? { ...ch, models: ch.models.filter((m) => m.id !== action.payload.model_id) }
            : ch
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
  // Channel operations
  addChannel: (channel: Omit<Channel, 'id' | 'models'>) => Promise<void>;
  updateChannel: (id: string, updates: Partial<Channel>) => Promise<void>;
  deleteChannel: (id: string) => Promise<void>;
  fetchChannelModels: (channel_id: string) => Promise<void>;

  // Model operations
  toggleModel: (channel_id: string, model_id: string, enabled: boolean) => Promise<void>;
  addModel: (channel_id: string, model: Omit<ModelInfo, 'id'>) => Promise<void>;
  removeModel: (channel_id: string, model_id: string) => Promise<void>;

  // Role operations
  assignRole: (role: WorkflowRole, channel_id: string, model_id: string) => Promise<void>;
  clearRole: (role: WorkflowRole) => void;

  // Snapshot operations
  saveSnapshot: (task_id: string) => void;
  getSnapshot: (task_id: string) => ModelSnapshot | null;

  // Utility
  refreshChannels: () => Promise<void>;
  getEnabledModels: () => Array<{ channel: Channel; model: ModelInfo }>;
  getRoleDisplayName: (role: WorkflowRole) => string;
}

const ModelConfigContext = createContext<ModelConfigContextValue | null>(null);

// ============================================================================
// Provider
// ============================================================================

const STORAGE_KEY = 'paperbanana_model_config_v1';

export function ModelConfigProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(modelConfigReducer, initialState);

  // API Operations - Uses /api/v1/providers endpoint (Channel is an alias for Provider)
  const refreshChannels = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const response = await fetch('/api/v1/providers');
      if (!response.ok) throw new Error('Failed to fetch providers');
      const data = await response.json();
      // Map provider response to channel format
      const channels: Channel[] = (data.providers || []).map((p: {
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
        type: p.type as Channel['type'],
        display_name: p.display_name,
        query_model: p.query_model,
        gen_model: p.gen_model,
        base_url: p.base_url,
        timeout: typeof p.timeout === 'string' ? parseTimeout(p.timeout) : 60,
        enabled: p.enabled,
        models: (p.models || []).map(m => ({ ...m, enabled: m.enabled ?? true })),
        is_system: p.is_system,
        is_default: p.is_default,
        status: p.status as Channel['status'],
      }));
      dispatch({ type: 'SET_CHANNELS', payload: channels });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

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
    refreshChannels();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    return subscribeToConfigChanges(() => {
      void refreshChannels();
    });
  }, [refreshChannels]);

  // Save to localStorage on state change
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  }, [state]);

  const addChannel = useCallback(async (channel: Omit<Channel, 'id' | 'models'>) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      // Map channel to provider format
      const providerData = {
        type: channel.type,
        name: channel.name,
        display_name: channel.display_name,
        api_host: channel.base_url,
        api_key: channel.api_key,
        timeout_ms: channel.timeout * 1000,
        enabled: channel.enabled,
      };
      const response = await fetch('/api/v1/providers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(providerData),
      });
      if (!response.ok) throw new Error('Failed to add provider');
      const data = await response.json();
      // Map response back to channel format
      const newChannel: Channel = {
        id: data.provider.id,
        name: data.provider.name,
        type: data.provider.type as Channel['type'],
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
      dispatch({ type: 'ADD_CHANNEL', payload: newChannel });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const updateChannel = useCallback(async (id: string, updates: Partial<Channel>) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      // Map channel updates to provider format
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
      dispatch({ type: 'UPDATE_CHANNEL', payload: { id, updates } });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const deleteChannel = useCallback(async (id: string) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const response = await fetch(`/api/v1/providers/${id}`, { method: 'DELETE' });
      if (!response.ok) throw new Error('Failed to delete provider');
      dispatch({ type: 'DELETE_CHANNEL', payload: id });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
      throw error;
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const fetchChannelModels = useCallback(async (channel_id: string) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const response = await fetch(`/api/v1/providers/${channel_id}/models`);
      if (!response.ok) throw new Error('Failed to fetch models');
      const data = await response.json();
      // Map domain models to our ModelInfo format
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
        type: 'SET_CHANNEL_MODELS',
        payload: { channel_id, models },
      });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const toggleModel = useCallback(async (channel_id: string, model_id: string, enabled: boolean) => {
    try {
      // Get current channel state
      const channel = state.channels.find(c => c.id === channel_id);
      if (!channel) return;

      // Update models array with the toggle
      const updatedModels = channel.models.map(m =>
        m.id === model_id ? { ...m, enabled } : m
      );

      const response = await fetch(`/api/v1/providers/${channel_id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ models: updatedModels }),
      });
      if (!response.ok) throw new Error('Failed to toggle model');
      dispatch({ type: 'TOGGLE_MODEL', payload: { channel_id, model_id, enabled } });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    }
  }, [state.channels]);

  const addModel = useCallback(async (channel_id: string, model: Omit<ModelInfo, 'id'>) => {
    try {
      const channel = state.channels.find(c => c.id === channel_id);
      if (!channel) return;

      // Add model to existing models
      const newModel: ModelInfo = {
        ...model,
        id: model.name.toLowerCase().replace(/\s+/g, '-'),
      };
      const updatedModels = [...channel.models, newModel];

      const response = await fetch(`/api/v1/providers/${channel_id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ models: updatedModels }),
      });
      if (!response.ok) throw new Error('Failed to add model');
      dispatch({ type: 'ADD_MODEL', payload: { channel_id, model: newModel } });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    }
  }, [state.channels]);

  const removeModel = useCallback(async (channel_id: string, model_id: string) => {
    try {
      const channel = state.channels.find(c => c.id === channel_id);
      if (!channel) return;

      // Remove model from array
      const updatedModels = channel.models.filter(m => m.id !== model_id);

      const response = await fetch(`/api/v1/providers/${channel_id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ models: updatedModels }),
      });
      if (!response.ok) throw new Error('Failed to remove model');
      dispatch({ type: 'REMOVE_MODEL', payload: { channel_id, model_id } });
    } catch (error) {
      dispatch({ type: 'SET_ERROR', payload: (error as Error).message });
    }
  }, [state.channels]);

  const assignRole = useCallback(
    async (role: WorkflowRole, channel_id: string, model_id: string) => {
      const channel = state.channels.find((c) => c.id === channel_id);
      const model = channel?.models.find((m) => m.id === model_id);
      if (!channel || !model) return;

      const assignment: RoleAssignment = {
        role,
        channel_id,
        model_id,
        channel_name: channel.display_name,
        model_name: model.name,
      };

      try {
        // Update provider with the model assignment
        const updateKey = role === 'image_generation' ? 'gen_model' : 'query_model';
        const response = await fetch(`/api/v1/providers/${channel_id}`, {
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
    [state.channels]
  );

  const clearRole = useCallback((role: WorkflowRole) => {
    dispatch({ type: 'CLEAR_ROLE_ASSIGNMENT', payload: role });
  }, []);

  const saveSnapshot = useCallback(
    (task_id: string) => {
      const snapshot: ModelSnapshot = {
        timestamp: new Date().toISOString(),
        channel_assignments: {
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
    const result: Array<{ channel: Channel; model: ModelInfo }> = [];
    for (const channel of state.channels) {
      if (!channel.enabled) continue;
      for (const model of channel.models) {
        if (model.enabled) {
          result.push({ channel, model });
        }
      }
    }
    return result;
  }, [state.channels]);

  const getRoleDisplayName = useCallback((role: WorkflowRole) => {
    const names: Record<WorkflowRole, string> = {
      image_generation: 'Image Generation',
      retrieval_reasoning: 'Retrieval & Reasoning',
    };
    return names[role];
  }, []);

  const value: ModelConfigContextValue = {
    ...state,
    addChannel,
    updateChannel,
    deleteChannel,
    fetchChannelModels,
    toggleModel,
    addModel,
    removeModel,
    assignRole,
    clearRole,
    saveSnapshot,
    getSnapshot,
    refreshChannels,
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

export function useChannel(channelId: string): Channel | undefined {
  const { channels } = useModelConfig();
  return channels.find((c) => c.id === channelId);
}

export function useRoleAssignment(role: WorkflowRole): RoleAssignment | null {
  const { role_assignments } = useModelConfig();
  return role_assignments[role];
}
