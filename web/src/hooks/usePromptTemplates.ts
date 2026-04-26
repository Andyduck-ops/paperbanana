import { useState, useEffect, useCallback } from 'react';

export interface PromptTemplate {
  id: string;
  name: string;
  methodContent: string;
  caption: string;
  createdAt: string;
}

export interface PromptTemplatesState {
  templates: PromptTemplate[];
  isLoading: boolean;
}

const STORAGE_KEY = 'prompt-templates';

function loadTemplatesFromStorage(): PromptTemplate[] {
  if (typeof window === 'undefined') return [];
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      return JSON.parse(stored);
    }
  } catch (e) {
    console.error('Failed to load prompt templates:', e);
  }
  return [];
}

function saveTemplatesToStorage(templates: PromptTemplate[]) {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(templates));
  } catch (e) {
    console.error('Failed to save prompt templates:', e);
  }
}

export function usePromptTemplates() {
  const [state, setState] = useState<PromptTemplatesState>({
    templates: [],
    isLoading: true,
  });

  // Load from localStorage on mount
  useEffect(() => {
    const templates = loadTemplatesFromStorage();
    setState({
      templates,
      isLoading: false,
    });
  }, []);

  // Save to localStorage whenever templates change
  useEffect(() => {
    if (!state.isLoading) {
      saveTemplatesToStorage(state.templates);
    }
  }, [state.templates, state.isLoading]);

  const addTemplate = useCallback((name: string, methodContent: string, caption: string) => {
    const newTemplate: PromptTemplate = {
      id: `template-${Date.now()}`,
      name: name.trim() || `Template ${state.templates.length + 1}`,
      methodContent,
      caption,
      createdAt: new Date().toISOString(),
    };
    setState((prev) => ({
      ...prev,
      templates: [...prev.templates, newTemplate],
    }));
    return newTemplate.id;
  }, [state.templates.length]);

  const updateTemplate = useCallback((id: string, updates: Partial<Omit<PromptTemplate, 'id' | 'createdAt'>>) => {
    setState((prev) => ({
      ...prev,
      templates: prev.templates.map((t) =>
        t.id === id ? { ...t, ...updates } : t
      ),
    }));
  }, []);

  const deleteTemplate = useCallback((id: string) => {
    setState((prev) => ({
      ...prev,
      templates: prev.templates.filter((t) => t.id !== id),
    }));
  }, []);

  const getTemplate = useCallback((id: string): PromptTemplate | undefined => {
    return state.templates.find((t) => t.id === id);
  }, [state.templates]);

  return {
    ...state,
    addTemplate,
    updateTemplate,
    deleteTemplate,
    getTemplate,
  };
}
