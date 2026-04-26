import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { queryKeys } from '../../lib/queryClient';
import { Provider } from '../../stores';

// API functions
const fetchProviders = async (): Promise<Provider[]> => {
  const response = await fetch('/api/providers');
  if (!response.ok) {
    throw new Error('Failed to fetch providers');
  }
  return response.json();
};

const fetchProvider = async (id: string): Promise<Provider> => {
  const response = await fetch(`/api/providers/${id}`);
  if (!response.ok) {
    throw new Error('Failed to fetch provider');
  }
  return response.json();
};

const createProvider = async (provider: Omit<Provider, 'id'>): Promise<Provider> => {
  const response = await fetch('/api/providers', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(provider),
  });
  if (!response.ok) {
    throw new Error('Failed to create provider');
  }
  return response.json();
};

const updateProvider = async ({ id, ...data }: Provider): Promise<Provider> => {
  const response = await fetch(`/api/providers/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error('Failed to update provider');
  }
  return response.json();
};

const deleteProvider = async (id: string): Promise<void> => {
  const response = await fetch(`/api/providers/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error('Failed to delete provider');
  }
};

// Hooks
export function useProvidersQuery() {
  return useQuery({
    queryKey: queryKeys.providers,
    queryFn: fetchProviders,
  });
}

export function useProviderQuery(id: string) {
  return useQuery({
    queryKey: queryKeys.provider(id),
    queryFn: () => fetchProvider(id),
    enabled: !!id,
  });
}

export function useCreateProviderMutation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: createProvider,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.providers });
    },
  });
}

export function useUpdateProviderMutation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: updateProvider,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.providers });
      queryClient.invalidateQueries({ queryKey: queryKeys.provider(data.id) });
    },
  });
}

export function useDeleteProviderMutation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: deleteProvider,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.providers });
    },
  });
}

// Compatibility hook with legacy useProviders interface
export function useProvidersCompat() {
  const { data: providers = [], isLoading, error, refetch } = useProvidersQuery();
  
  const refresh = useCallback(async () => {
    await refetch();
  }, [refetch]);
  
  return {
    providers,
    isLoading,
    error: error?.message || null,
    refresh,
  };
}
