import { useCallback, useEffect } from 'react';
import {
  useProviderStore,
  Provider,
  ModelInfo,
  WorkflowRole,
  RoleAssignment,
  fetchProviders,
  createProvider,
  updateProviderAPI,
  deleteProviderAPI,
  fetchProviderModelsAPI,
  assignRoleAPI,
} from '../stores';
import { subscribeToConfigChanges } from '../lib/configStream';

// ============================================================================
// Provider Store Adapter Hook
// 
// This hook provides a bridge between the existing ModelConfigContext API
// and the new Zustand store. It maintains the same interface for easy migration.
// ============================================================================

export interface ProviderStoreAdapter {
  // State
  providers: Provider[];
  role_assignments: Record<WorkflowRole, RoleAssignment | null>;
  loading: boolean;
  error: string | null;
  
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
  getSnapshot: (task_id: string) => { timestamp: string; provider_assignments: Record<WorkflowRole, RoleAssignment> } | null;
  
  // Utility
  refreshProviders: () => Promise<void>;
  getEnabledModels: () => Array<{ provider: Provider; model: ModelInfo }>;
  getRoleDisplayName: (role: WorkflowRole) => string;
  
  // Backward compatibility aliases
  channels: Provider[];
  addChannel: (provider: Omit<Provider, 'id' | 'models'>) => Promise<void>;
  updateChannel: (id: string, updates: Partial<Provider>) => Promise<void>;
  deleteChannel: (id: string) => Promise<void>;
  fetchChannelModels: (provider_id: string) => Promise<void>;
  refreshChannels: () => Promise<void>;
}

export function useProviderStoreAdapter(): ProviderStoreAdapter {
  // Get state from Zustand store
  const providers = useProviderStore((state) => state.providers);
  const role_assignments = useProviderStore((state) => state.role_assignments);
  const loading = useProviderStore((state) => state.loading);
  const error = useProviderStore((state) => state.error);
  const snapshots = useProviderStore((state) => state.snapshots);
  
  // Get actions from Zustand store
  const setProviders = useProviderStore((state) => state.setProviders);
  const addProviderToStore = useProviderStore((state) => state.addProvider);
  const updateProviderInStore = useProviderStore((state) => state.updateProvider);
  const deleteProviderFromStore = useProviderStore((state) => state.deleteProvider);
  const setProviderModels = useProviderStore((state) => state.setProviderModels);
  const toggleModelInStore = useProviderStore((state) => state.toggleModel);
  const addModelToStore = useProviderStore((state) => state.addModel);
  const removeModelFromStore = useProviderStore((state) => state.removeModel);
  const setRoleAssignment = useProviderStore((state) => state.setRoleAssignment);
  const clearRoleAssignment = useProviderStore((state) => state.clearRoleAssignment);
  const saveSnapshotToStore = useProviderStore((state) => state.saveSnapshot);
  const setLoading = useProviderStore((state) => state.setLoading);
  const setError = useProviderStore((state) => state.setError);

  // Provider operations
  const refreshProviders = useCallback(async () => {
    setLoading(true);
    try {
      const providers = await fetchProviders();
      setProviders(providers);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }, [setProviders, setLoading, setError]);

  const addProvider = useCallback(async (provider: Omit<Provider, 'id' | 'models'>) => {
    setLoading(true);
    try {
      const newProvider = await createProvider(provider);
      addProviderToStore(newProvider);
    } catch (err) {
      setError((err as Error).message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [addProviderToStore, setLoading, setError]);

  const updateProvider = useCallback(async (id: string, updates: Partial<Provider>) => {
    setLoading(true);
    try {
      await updateProviderAPI(id, updates);
      updateProviderInStore(id, updates);
    } catch (err) {
      setError((err as Error).message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [updateProviderInStore, setLoading, setError]);

  const deleteProvider = useCallback(async (id: string) => {
    setLoading(true);
    try {
      await deleteProviderAPI(id);
      deleteProviderFromStore(id);
    } catch (err) {
      setError((err as Error).message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [deleteProviderFromStore, setLoading, setError]);

  const fetchProviderModels = useCallback(async (provider_id: string) => {
    setLoading(true);
    try {
      const models = await fetchProviderModelsAPI(provider_id);
      setProviderModels(provider_id, models);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }, [setProviderModels, setLoading, setError]);

  // Model operations
  const toggleModel = useCallback(async (provider_id: string, model_id: string, enabled: boolean) => {
    try {
      const provider = providers.find(p => p.id === provider_id);
      if (!provider) return;

      const updatedModels = provider.models.map(m =>
        m.id === model_id ? { ...m, enabled } : m
      );

      await updateProviderAPI(provider_id, { models: updatedModels });
      toggleModelInStore(provider_id, model_id, enabled);
    } catch (err) {
      setError((err as Error).message);
    }
  }, [providers, toggleModelInStore, setError]);

  const addModel = useCallback(async (provider_id: string, model: Omit<ModelInfo, 'id'>) => {
    try {
      const provider = providers.find(p => p.id === provider_id);
      if (!provider) return;

      const newModel: ModelInfo = {
        ...model,
        id: model.name.toLowerCase().replace(/\s+/g, '-'),
      };
      const updatedModels = [...provider.models, newModel];

      await updateProviderAPI(provider_id, { models: updatedModels });
      addModelToStore(provider_id, newModel);
    } catch (err) {
      setError((err as Error).message);
    }
  }, [providers, addModelToStore, setError]);

  const removeModel = useCallback(async (provider_id: string, model_id: string) => {
    try {
      const provider = providers.find(p => p.id === provider_id);
      if (!provider) return;

      const updatedModels = provider.models.filter(m => m.id !== model_id);

      await updateProviderAPI(provider_id, { models: updatedModels });
      removeModelFromStore(provider_id, model_id);
    } catch (err) {
      setError((err as Error).message);
    }
  }, [providers, removeModelFromStore, setError]);

  // Role operations
  const assignRole = useCallback(async (role: WorkflowRole, provider_id: string, model_id: string) => {
    const provider = providers.find((p) => p.id === provider_id);
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
      await assignRoleAPI(role, provider_id, model_id);
      setRoleAssignment(assignment);
    } catch (err) {
      setError((err as Error).message);
    }
  }, [providers, setRoleAssignment, setError]);

  const clearRole = useCallback((role: WorkflowRole) => {
    clearRoleAssignment(role);
  }, [clearRoleAssignment]);

  // Snapshot operations
  const saveSnapshot = useCallback((task_id: string) => {
    const snapshot = {
      timestamp: new Date().toISOString(),
      provider_assignments: {
        image_generation: role_assignments.image_generation!,
        retrieval_reasoning: role_assignments.retrieval_reasoning!,
      },
    };
    saveSnapshotToStore(task_id, snapshot);
  }, [role_assignments, saveSnapshotToStore]);

  const getSnapshot = useCallback((task_id: string) => {
    return snapshots[task_id] || null;
  }, [snapshots]);

  // Utility
  const getEnabledModels = useCallback(() => {
    const result: Array<{ provider: Provider; model: ModelInfo }> = [];
    for (const provider of providers) {
      if (!provider.enabled) continue;
      for (const model of provider.models) {
        if (model.enabled) {
          result.push({ provider, model });
        }
      }
    }
    return result;
  }, [providers]);

  const getRoleDisplayName = useCallback((role: WorkflowRole) => {
    const names: Record<WorkflowRole, string> = {
      image_generation: 'Image Generation',
      retrieval_reasoning: 'Retrieval & Reasoning',
    };
    return names[role];
  }, []);

  return {
    // State
    providers,
    role_assignments,
    loading,
    error,
    
    // Provider operations
    addProvider,
    updateProvider,
    deleteProvider,
    fetchProviderModels,
    
    // Model operations
    toggleModel,
    addModel,
    removeModel,
    
    // Role operations
    assignRole,
    clearRole,
    
    // Snapshot operations
    saveSnapshot,
    getSnapshot,
    
    // Utility
    refreshProviders,
    getEnabledModels,
    getRoleDisplayName,
    
    // Backward compatibility aliases
    channels: providers,
    addChannel: addProvider,
    updateChannel: updateProvider,
    deleteChannel: deleteProvider,
    fetchChannelModels: fetchProviderModels,
    refreshChannels: refreshProviders,
  };
}

// ============================================================================
// Hook for initializing provider state from backend
// ============================================================================

export function useProviderStoreInit() {
  const refreshProviders = useProviderStore((state) => state.setProviders);
  const setLoading = useProviderStore((state) => state.setLoading);
  const setError = useProviderStore((state) => state.setError);

  useEffect(() => {
    // Load from localStorage first (Zustand persist handles this)
    // Then fetch fresh data from backend
    const loadProviders = async () => {
      setLoading(true);
      try {
        const providers = await fetchProviders();
        refreshProviders(providers);
      } catch (err) {
        setError((err as Error).message);
      } finally {
        setLoading(false);
      }
    };

    loadProviders();

    // Subscribe to config changes
    const unsubscribe = subscribeToConfigChanges(() => {
      loadProviders();
    });

    return unsubscribe;
  }, [refreshProviders, setLoading, setError]);
}
