import { useCallback, useRef, useState } from 'react';

import { reduceBatchStreamEvent } from '../lib/batchProgress';
import type { BatchProgress } from '../types/batch';
import type { GenerateRequest } from '../types/api';

interface StartBatchOptions {
  visualizerNode?: string;
  config?: Pick<
    GenerateRequest,
    'aspect_ratio' | 'critic_rounds' | 'retrieval_mode' | 'pipeline_mode' | 'query_model' | 'gen_model'
  >;
}

interface UseBatchGenerationResult {
  isGenerating: boolean;
  progress: BatchProgress | null;
  result: BatchProgress | null;
  error: string | null;
  startBatch: (
    prompt: string,
    numCandidates: number,
    options?: StartBatchOptions
  ) => Promise<void>;
  resetBatch: () => void;
}

function createInitialState() {
  return {
    isGenerating: false,
    progress: null as BatchProgress | null,
    result: null as BatchProgress | null,
    error: null as string | null,
  };
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException
    ? error.name === 'AbortError'
    : error instanceof Error && error.name === 'AbortError';
}

export function useBatchGeneration(): UseBatchGenerationResult {
  const [state, setState] = useState(createInitialState);
  const abortControllerRef = useRef<AbortController | null>(null);
  const requestIdRef = useRef(0);

  const cancelActiveRequest = useCallback(() => {
    abortControllerRef.current?.abort();
    abortControllerRef.current = null;
    requestIdRef.current += 1;
  }, []);

  const resetBatch = useCallback(() => {
    cancelActiveRequest();
    setState(createInitialState());
  }, [cancelActiveRequest]);

  const startBatch = useCallback(
    async (
      prompt: string,
      numCandidates: number,
      options?: StartBatchOptions
    ) => {
      cancelActiveRequest();

      const controller = new AbortController();
      abortControllerRef.current = controller;
      requestIdRef.current += 1;
      const requestId = requestIdRef.current;

      setState({
        isGenerating: true,
        progress: null,
        result: null,
        error: null,
      });

      try {
        const response = await fetch('/api/v1/generate/batch', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            prompt,
            num_candidates: numCandidates,
            visualizer_node: options?.visualizerNode,
            aspect_ratio: options?.config?.aspect_ratio,
            critic_rounds: options?.config?.critic_rounds,
            retrieval_mode: options?.config?.retrieval_mode,
            pipeline_mode: options?.config?.pipeline_mode,
            query_model: options?.config?.query_model,
            gen_model: options?.config?.gen_model,
          }),
          signal: controller.signal,
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }

        const reader = response.body?.getReader();
        if (!reader) {
          throw new Error('No response body');
        }

        const decoder = new TextDecoder();
        let buffer = '';
        let currentProgress: BatchProgress | null = null;
        let latestResult: BatchProgress | null = null;

        while (true) {
          const { done, value } = await reader.read();
          if (done || controller.signal.aborted || requestId !== requestIdRef.current) {
            break;
          }

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split('\n');
          buffer = lines.pop() || '';

          for (const line of lines) {
            if (
              controller.signal.aborted ||
              requestId !== requestIdRef.current ||
              !line.startsWith('data:')
            ) {
              continue;
            }

            try {
              const data = JSON.parse(line.slice(5).trim());
              currentProgress = reduceBatchStreamEvent(
                currentProgress,
                data,
                numCandidates
              );

              if (!currentProgress) {
                continue;
              }

              if (data.results) {
                latestResult = currentProgress;
              }

              setState((previous) => {
                if (requestId !== requestIdRef.current) {
                  return previous;
                }

                return {
                  ...previous,
                  progress: currentProgress,
                  result: latestResult,
                };
              });
            } catch {
              // Skip unparseable lines to preserve current stream semantics.
            }
          }
        }

        if (requestId !== requestIdRef.current || controller.signal.aborted) {
          return;
        }

        if (abortControllerRef.current === controller) {
          abortControllerRef.current = null;
        }

        setState((previous) => ({
          ...previous,
          isGenerating: false,
          progress: currentProgress,
          result: latestResult,
        }));
      } catch (error) {
        if (requestId !== requestIdRef.current || isAbortError(error)) {
          return;
        }

        if (abortControllerRef.current === controller) {
          abortControllerRef.current = null;
        }

        setState((previous) => ({
          ...previous,
          isGenerating: false,
          error:
            error instanceof Error ? error.message : 'Batch generation failed',
        }));
      }
    },
    [cancelActiveRequest]
  );

  return {
    isGenerating: state.isGenerating,
    progress: state.progress,
    result: state.result,
    error: state.error,
    startBatch,
    resetBatch,
  };
}
