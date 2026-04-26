import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback, useRef, useState } from 'react';
// import { queryKeys } from '../../lib/queryClient';
// Types defined locally


interface GenerationStage {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress?: number;
  message?: string;
}

export interface GenerateOptions {
  content?: string;
  visualIntent?: string;
  visualizerNode?: string;
  config?: {
    aspect_ratio?: string;
    critic_rounds?: number;
    retrieval_mode?: string;
    pipeline_mode?: string;
    query_model?: string;
    gen_model?: string;
  };
}

interface GenerateRequest {
  prompt: string;
  options?: GenerateOptions;
}

interface GenerateResponse {
  sessionId: string;
  artifacts: Array<{
    kind: string;
    mimeType: string;
    data?: string;
    assetId?: string;
    projectId?: string;
    uri?: string;
  }>;
  stages: GenerationStage[];
}



const generateImage = async (
  request: GenerateRequest,
  onStage?: (stage: GenerationStage) => void
): Promise<GenerateResponse> => {
  return new Promise((resolve, reject) => {
    const eventSource = new EventSource('/api/generate?' + new URLSearchParams({
      prompt: request.prompt,
      ...(request.options?.config?.aspect_ratio && { aspect_ratio: request.options.config.aspect_ratio }),
      ...(request.options?.config?.critic_rounds && { critic_rounds: request.options.config.critic_rounds.toString() }),
      ...(request.options?.config?.retrieval_mode && { retrieval_mode: request.options.config.retrieval_mode }),
      ...(request.options?.config?.pipeline_mode && { pipeline_mode: request.options.config.pipeline_mode }),
      ...(request.options?.config?.query_model && { query_model: request.options.config.query_model }),
      ...(request.options?.config?.gen_model && { gen_model: request.options.config.gen_model }),
    }));

    let result: GenerateResponse | null = null;
    const stages: GenerationStage[] = [];

    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      switch (data.type) {
        case 'stage':
          const stage: GenerationStage = {
            id: data.id,
            name: data.name,
            status: data.status,
            progress: data.progress,
            message: data.message,
          };
          stages.push(stage);
          onStage?.(stage);
          break;
        case 'complete':
          result = {
            sessionId: data.sessionId,
            artifacts: data.artifacts,
            stages: data.stages || stages,
          };
          eventSource.close();
          resolve(result);
          break;
        case 'error':
          eventSource.close();
          reject(new Error(data.message || 'Generation failed'));
          break;
      }
    };

    eventSource.onerror = () => {
      eventSource.close();
      reject(new Error('Connection error'));
    };
  });
};

export function useGenerateMutation() {
  const queryClient = useQueryClient();
  const abortControllerRef = useRef<AbortController | null>(null);
  const [stages, setStages] = useState<GenerationStage[]>([]);

  const mutation = useMutation({
    mutationFn: async (request: GenerateRequest) => {
      abortControllerRef.current = new AbortController();
      setStages([]);
      
      try {
        const result = await generateImage(request, (stage) => {
          setStages((prev) => [...prev, stage]);
        });
        return result;
      } finally {
        abortControllerRef.current = null;
      }
    },
    onSuccess: () => {
      // Invalidate history after successful generation
      queryClient.invalidateQueries({ queryKey: ['history'] });
    },
  });

  const cancel = useCallback(() => {
    abortControllerRef.current?.abort();
    mutation.reset();
    setStages([]);
  }, [mutation]);

  const reset = useCallback(() => {
    mutation.reset();
    setStages([]);
  }, [mutation]);

  return {
    isGenerating: mutation.isPending,
    stages,
    result: mutation.data || null,
    error: mutation.error?.message || null,
    generate: mutation.mutateAsync,
    cancel,
    reset,
  };
}

// Compatibility hook
export function useGenerateCompat() {
  const mutation = useGenerateMutation();
  
  const generate = useCallback(async (
    prompt: string, 
    options?: GenerateOptions
  ) => {
    return mutation.generate({ prompt, options });
  }, [mutation.generate]);

  return {
    ...mutation,
    generate,
  };
}
