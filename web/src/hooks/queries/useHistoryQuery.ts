import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { queryKeys } from '../../lib/queryClient';

export interface HistoryItem {
  id: string;
  sessionId: string;
  prompt: string;
  mode: 'generate' | 'refine' | 'batch';
  status: 'pending' | 'running' | 'completed' | 'failed';
  createdAt: string;
  completedAt?: string;
  artifacts?: Array<{
    kind: string;
    mimeType: string;
    data?: string;
    assetId?: string;
    projectId?: string;
    uri?: string;
  }>;
  stages?: Array<{
    id: string;
    name: string;
    status: string;
    progress?: number;
    message?: string;
  }>;
  batchId?: string;
  candidates?: Array<{
    sessionId: string;
    candidateId: number;
    status: string;
    artifacts: Array<{
      kind: string;
      mimeType: string;
      data?: string;
      assetId?: string;
      projectId?: string;
      uri?: string;
    }>;
    error?: string;
  }>;
  successful?: number;
  failed?: number;
  error?: string;
  resumeMetadata?: unknown;
}

interface HistoryResponse {
  items: HistoryItem[];
  total: number;
  page: number;
  limit: number;
}

interface HistoryFilters {
  page?: number;
  limit?: number;
  mode?: 'generate' | 'refine' | 'batch';
  status?: string;
}

const fetchHistory = async (filters: HistoryFilters = {}): Promise<HistoryResponse> => {
  const params = new URLSearchParams();
  if (filters.page !== undefined) params.set('page', filters.page.toString());
  if (filters.limit !== undefined) params.set('limit', filters.limit.toString());
  if (filters.mode) params.set('mode', filters.mode);
  if (filters.status) params.set('status', filters.status);
  
  const response = await fetch(`/api/history?${params}`);
  if (!response.ok) {
    throw new Error('Failed to fetch history');
  }
  return response.json();
};

const fetchHistoryItem = async (id: string): Promise<HistoryItem> => {
  const response = await fetch(`/api/history/${id}`);
  if (!response.ok) {
    throw new Error('Failed to fetch history item');
  }
  return response.json();
};

const restoreSession = async (id: string): Promise<HistoryItem> => {
  const response = await fetch(`/api/history/${id}/restore`, {
    method: 'POST',
  });
  if (!response.ok) {
    throw new Error('Failed to restore session');
  }
  return response.json();
};

const deleteHistoryItem = async (id: string): Promise<void> => {
  const response = await fetch(`/api/history/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error('Failed to delete history item');
  }
};

const clearHistory = async (): Promise<void> => {
  const response = await fetch('/api/history', {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error('Failed to clear history');
  }
};

// Hooks
export function useHistoryQuery(filters: HistoryFilters = {}) {
  return useQuery({
    queryKey: queryKeys.history(filters.page, filters.limit),
    queryFn: () => fetchHistory(filters),
  });
}

export function useHistoryItemQuery(id: string) {
  return useQuery({
    queryKey: ['history', 'item', id],
    queryFn: () => fetchHistoryItem(id),
    enabled: !!id,
  });
}

export function useRestoreSessionMutation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: restoreSession,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['history'] });
    },
  });
}

export function useDeleteHistoryItemMutation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: deleteHistoryItem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['history'] });
    },
  });
}

export function useClearHistoryMutation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: clearHistory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['history'] });
    },
  });
}

// Compatibility hook with legacy useHistory interface
export function useHistoryCompat() {
  const { data, isLoading, error, refetch } = useHistoryQuery({ limit: 100 });
  const restoreMutation = useRestoreSessionMutation();
  const deleteMutation = useDeleteHistoryItemMutation();
  
  const restoreSession = useCallback(async (id: string) => {
    try {
      const result = await restoreMutation.mutateAsync(id);
      return {
        id: result.id,
        mode: result.mode,
        status: result.status,
        prompt: result.prompt,
        artifacts: result.artifacts || [],
        stages: result.stages || [],
        candidates: result.candidates,
        batchId: result.batchId,
        successful: result.successful,
        failed: result.failed,
        startedAt: result.createdAt,
        completedAt: result.completedAt,
        error: result.error,
        resumeMetadata: result.resumeMetadata,
      };
    } catch {
      return null;
    }
  }, [restoreMutation]);

  const refresh = useCallback(async () => {
    await refetch();
  }, [refetch]);
  
  return {
    items: data?.items || [],
    count: data?.items?.length || 0,
    isLoading,
    error: error?.message || null,
    refresh,
    restoreSession,
    deleteItem: deleteMutation.mutateAsync,
  };
}
