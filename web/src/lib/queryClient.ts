import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      refetchOnWindowFocus: false,
      retry: 2,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    },
    mutations: {
      retry: 1,
    },
  },
});

// Query keys for type-safe cache management
export const queryKeys = {
  providers: ['providers'] as const,
  provider: (id: string) => ['providers', id] as const,
  history: (page?: number, limit?: number) => 
    ['history', { page, limit }] as const,
  projects: ['projects'] as const,
  project: (id: string) => ['projects', id] as const,
  config: ['config'] as const,
  templates: ['templates'] as const,
};
