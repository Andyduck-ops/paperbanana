import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '../lib/api';

export interface Template {
  id: string;
  name: string;
  description?: string;
  category: string;
  thumbnail?: string;
  created_at: string;
}

export interface TemplatesState {
  templates: Template[];
  isLoading: boolean;
  error: string | null;
}

export function useTemplates() {
  const [state, setState] = useState<TemplatesState>({
    templates: [],
    isLoading: false,
    error: null,
  });

  const fetchTemplates = useCallback(async () => {
    setState((prev) => ({ ...prev, isLoading: true, error: null }));
    try {
      const response = await apiClient.listTemplates();
      setState({
        templates: response.templates,
        isLoading: false,
        error: null,
      });
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load templates',
      }));
    }
  }, []);

  const createTemplate = useCallback(async (data: { name: string; description?: string; category: string; content: string }) => {
    try {
      const response = await apiClient.createTemplate(data);
      await fetchTemplates();
      return response;
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to create template',
      }));
      return null;
    }
  }, [fetchTemplates]);

  const updateTemplate = useCallback(async (templateId: string, data: { name?: string; description?: string; category?: string; content?: string }) => {
    try {
      const response = await apiClient.updateTemplate(templateId, data);
      await fetchTemplates();
      return response;
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to update template',
      }));
      return null;
    }
  }, [fetchTemplates]);

  const deleteTemplate = useCallback(async (templateId: string) => {
    try {
      await apiClient.deleteTemplate(templateId);
      await fetchTemplates();
      return true;
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to delete template',
      }));
      return false;
    }
  }, [fetchTemplates]);

  useEffect(() => {
    fetchTemplates();
  }, [fetchTemplates]);

  return {
    ...state,
    createTemplate,
    updateTemplate,
    deleteTemplate,
    refresh: fetchTemplates,
  };
}
