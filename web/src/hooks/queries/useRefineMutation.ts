import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback, useState } from 'react';
// import { queryKeys } from '../../lib/queryClient';

export interface RefineImage {
  file: File;
  previewUrl: string;
}

export interface RefineOptions {
  image: RefineImage;
  instructions: string;
  resolution: '2K' | '4K';
  enable_iteration?: boolean;
  max_iterations?: number;
}

export interface RefineResult {
  sessionId: string;
  status: 'completed' | 'failed';
  image: {
    data: string;
    mimeType: string;
  };
  iterations?: Array<{
    iteration: number;
    status: string;
    feedback?: string;
    image?: {
      data: string;
      mimeType: string;
    };
  }>;
}



const refineImage = async (options: RefineOptions): Promise<RefineResult> => {
  const formData = new FormData();
  formData.append('image', options.image.file);
  formData.append('instructions', options.instructions);
  formData.append('resolution', options.resolution);
  
  if (options.enable_iteration) {
    formData.append('enable_iteration', 'true');
    formData.append('max_iterations', (options.max_iterations || 3).toString());
  }

  const response = await fetch('/api/refine', {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || 'Refinement failed');
  }

  return response.json();
};

export function useRefineMutation() {
  const queryClient = useQueryClient();
  const [isRefining, setIsRefining] = useState(false);

  const mutation = useMutation({
    mutationFn: async (options: RefineOptions) => {
      setIsRefining(true);
      try {
        return await refineImage(options);
      } finally {
        setIsRefining(false);
      }
    },
    onSuccess: () => {
      // Invalidate history after successful refinement
      queryClient.invalidateQueries({ queryKey: ['history'] });
    },
  });

  const reset = useCallback(() => {
    mutation.reset();
  }, [mutation]);

  return {
    isRefining: mutation.isPending || isRefining,
    result: mutation.data || null,
    error: mutation.error?.message || null,
    refine: mutation.mutateAsync,
    reset,
  };
}

// Compatibility hook with callbacks
export function useRefineCompat(options?: {
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}) {
  const mutation = useRefineMutation();
  
  const refine = useCallback(async (request: {
    image: RefineImage;
    instructions: string;
    resolution: '2K' | '4K';
    enable_iteration?: boolean;
    max_iterations?: number;
  }) => {
    try {
      const result = await mutation.refine(request);
      options?.onSuccess?.();
      return result;
    } catch (error) {
      if (error instanceof Error) {
        options?.onError?.(error);
      }
      throw error;
    }
  }, [mutation, options]);

  return {
    isRefining: mutation.isRefining,
    result: mutation.result,
    error: mutation.error,
    refine,
    reset: mutation.reset,
  };
}
