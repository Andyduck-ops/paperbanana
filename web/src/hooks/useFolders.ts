import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '../lib/api';

export interface Folder {
  id: string;
  name: string;
  project_id: string;
  parent_id?: string;
  type: 'folder' | 'visualization';
  created_at: string;
  updated_at?: string;
}

export interface FolderItem {
  id: string;
  name: string;
  type: 'folder' | 'visualization';
  created_at: string;
}

export interface FoldersState {
  folders: Folder[];
  currentFolderId: string | null;
  items: FolderItem[];
  isLoading: boolean;
  error: string | null;
}

export function useFolders(projectId?: string) {
  const [state, setState] = useState<FoldersState>({
    folders: [],
    currentFolderId: null,
    items: [],
    isLoading: false,
    error: null,
  });

  const fetchFolders = useCallback(async () => {
    if (!projectId) return;
    
    setState((prev) => ({ ...prev, isLoading: true, error: null }));
    try {
      const response = await apiClient.listFolderContents(projectId, state.currentFolderId || undefined);
      setState((prev) => ({
        ...prev,
        items: response.items,
        isLoading: false,
        error: null,
      }));
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load folders',
      }));
    }
  }, [projectId, state.currentFolderId]);

  const navigateToFolder = useCallback((folderId: string | null) => {
    setState((prev) => ({ ...prev, currentFolderId: folderId }));
  }, []);

  const createFolder = useCallback(async (name: string, parentId?: string) => {
    if (!projectId) return null;
    
    try {
      const response = await apiClient.createFolder({
        name,
        project_id: projectId,
        parent_id: parentId || state.currentFolderId || undefined,
      });
      await fetchFolders();
      return response;
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to create folder',
      }));
      return null;
    }
  }, [projectId, state.currentFolderId, fetchFolders]);

  const renameFolder = useCallback(async (folderId: string, name: string) => {
    try {
      const response = await apiClient.updateFolder(folderId, { name });
      await fetchFolders();
      return response;
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to rename folder',
      }));
      return null;
    }
  }, [fetchFolders]);

  const deleteFolder = useCallback(async (folderId: string) => {
    try {
      await apiClient.deleteFolder(folderId);
      await fetchFolders();
      return true;
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to delete folder',
      }));
      return false;
    }
  }, [fetchFolders]);

  useEffect(() => {
    fetchFolders();
  }, [fetchFolders]);

  return {
    ...state,
    navigateToFolder,
    createFolder,
    renameFolder,
    deleteFolder,
    refresh: fetchFolders,
  };
}
