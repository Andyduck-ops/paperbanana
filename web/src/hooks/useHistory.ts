import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '../lib/api';

export interface HistorySession {
  id: string;
  projectId: string;
  createdAt: string;
  status: string;
  prompt?: string;
  thumbnailUrl?: string;
}

export interface HistoryState {
  sessions: HistorySession[];
  isLoading: boolean;
  error: string | null;
}

export function useHistory(projectId?: string) {
  const [state, setState] = useState<HistoryState>({
    sessions: [],
    isLoading: false,
    error: null,
  });

  const fetchHistory = useCallback(async () => {
    setState((prev) => ({ ...prev, isLoading: true, error: null }));
    try {
      const response = await apiClient.listHistory(projectId);
      setState({
        sessions: response.sessions.map((s) => ({
          id: s.id,
          projectId: s.project_id,
          createdAt: s.created_at,
          status: s.status,
        })),
        isLoading: false,
        error: null,
      });
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load history',
      }));
    }
  }, [projectId]);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  return {
    ...state,
    refresh: fetchHistory,
  };
}
