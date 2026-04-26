import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '../lib/api';

export interface Version {
  id: string;
  visualization_id: string;
  version: number;
  created_at: string;
  artifacts: Array<{
    id: string;
    kind: string;
    mime_type: string;
    data?: string;
    summary?: string;
  }>;
}

export interface VersionsState {
  versions: Version[];
  isLoading: boolean;
  error: string | null;
}

export function useVersions(projectId?: string, vizId?: string) {
  const [state, setState] = useState<VersionsState>({
    versions: [],
    isLoading: false,
    error: null,
  });

  const fetchVersions = useCallback(async () => {
    if (!projectId || !vizId) return;
    
    setState((prev) => ({ ...prev, isLoading: true, error: null }));
    try {
      const response = await apiClient.getVisualizationHistory(projectId, vizId);
      setState({
        versions: response.versions,
        isLoading: false,
        error: null,
      });
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load versions',
      }));
    }
  }, [projectId, vizId]);

  const restoreVersion = useCallback(async (versionId: string) => {
    if (!projectId || !vizId) return null;
    
    try {
      const response = await apiClient.restoreVersion(projectId, vizId, versionId);
      return response;
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to restore version',
      }));
      return null;
    }
  }, [projectId, vizId]);

  useEffect(() => {
    fetchVersions();
  }, [fetchVersions]);

  return {
    ...state,
    restoreVersion,
    refresh: fetchVersions,
  };
}
