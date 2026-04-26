import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { queryKeys } from '../../lib/queryClient';

export interface Project {
  id: string;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
  thumbnail?: string;
  itemCount?: number;
}

export interface ProjectFolder {
  id: string;
  name: string;
  parentId?: string;
  projectId: string;
}

export interface ProjectItem {
  id: string;
  name: string;
  type: 'image' | 'document' | 'other';
  folderId?: string;
  projectId: string;
  createdAt: string;
  thumbnail?: string;
  metadata?: Record<string, unknown>;
}

interface ProjectsResponse {
  projects: Project[];
  total: number;
}

// API functions
const fetchProjects = async (): Promise<ProjectsResponse> => {
  const response = await fetch('/api/projects');
  if (!response.ok) {
    throw new Error('Failed to fetch projects');
  }
  return response.json();
};

const fetchProject = async (id: string): Promise<Project> => {
  const response = await fetch(`/api/projects/${id}`);
  if (!response.ok) {
    throw new Error('Failed to fetch project');
  }
  return response.json();
};

const createProject = async (data: Omit<Project, 'id' | 'createdAt' | 'updatedAt'>): Promise<Project> => {
  const response = await fetch('/api/projects', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error('Failed to create project');
  }
  return response.json();
};

const updateProject = async ({ id, ...data }: Partial<Project> & { id: string }): Promise<Project> => {
  const response = await fetch(`/api/projects/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error('Failed to update project');
  }
  return response.json();
};

const deleteProject = async (id: string): Promise<void> => {
  const response = await fetch(`/api/projects/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error('Failed to delete project');
  }
};

const fetchProjectFolders = async (projectId: string): Promise<ProjectFolder[]> => {
  const response = await fetch(`/api/projects/${projectId}/folders`);
  if (!response.ok) {
    throw new Error('Failed to fetch project folders');
  }
  return response.json();
};

const fetchProjectItems = async (projectId: string, folderId?: string): Promise<ProjectItem[]> => {
  const params = folderId ? `?folderId=${folderId}` : '';
  const response = await fetch(`/api/projects/${projectId}/items${params}`);
  if (!response.ok) {
    throw new Error('Failed to fetch project items');
  }
  return response.json();
};

// Hooks
export function useProjectsQuery() {
  return useQuery({
    queryKey: queryKeys.projects,
    queryFn: fetchProjects,
  });
}

export function useProjectQuery(id: string) {
  return useQuery({
    queryKey: queryKeys.project(id),
    queryFn: () => fetchProject(id),
    enabled: !!id,
  });
}

export function useCreateProjectMutation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: createProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.projects });
    },
  });
}

export function useUpdateProjectMutation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: updateProject,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.projects });
      queryClient.invalidateQueries({ queryKey: queryKeys.project(data.id) });
    },
  });
}

export function useDeleteProjectMutation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: deleteProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.projects });
    },
  });
}

export function useProjectFoldersQuery(projectId: string) {
  return useQuery({
    queryKey: ['projects', projectId, 'folders'],
    queryFn: () => fetchProjectFolders(projectId),
    enabled: !!projectId,
  });
}

export function useProjectItemsQuery(projectId: string, folderId?: string) {
  return useQuery({
    queryKey: ['projects', projectId, 'items', folderId],
    queryFn: () => fetchProjectItems(projectId, folderId),
    enabled: !!projectId,
  });
}

// Compatibility hook
export function useProjectsCompat() {
  const { data, isLoading, error, refetch } = useProjectsQuery();
  const createMutation = useCreateProjectMutation();
  const updateMutation = useUpdateProjectMutation();
  const deleteMutation = useDeleteProjectMutation();
  
  const refresh = useCallback(async () => {
    await refetch();
  }, [refetch]);
  
  return {
    projects: data?.projects || [],
    isLoading,
    error: error?.message || null,
    refresh,
    createProject: createMutation.mutateAsync,
    updateProject: updateMutation.mutateAsync,
    deleteProject: deleteMutation.mutateAsync,
  };
}
