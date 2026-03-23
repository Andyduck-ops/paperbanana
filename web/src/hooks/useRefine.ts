import { useCallback, useState } from 'react';

import { refineImage } from '../lib/refine';
import type { RefineRequest, RefineResult } from '../types/api';

export interface UseRefineOptions {
  onSuccess?: (result: RefineResult) => void;
  onError?: (error: Error) => void;
}

export interface RefineState {
  isRefining: boolean;
  result: RefineResult | null;
  error: Error | null;
}

function createInitialState(): RefineState {
  return {
    isRefining: false,
    result: null,
    error: null,
  };
}

export function useRefine(options: UseRefineOptions = {}) {
  const { onSuccess, onError } = options;
  const [state, setState] = useState<RefineState>(createInitialState);

  const refine = useCallback(
    async (request: RefineRequest) => {
      setState((previous) => ({
        ...previous,
        isRefining: true,
        error: null,
      }));

      try {
        const result = await refineImage(request);
        setState({
          isRefining: false,
          result,
          error: null,
        });
        onSuccess?.(result);
        return result;
      } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        setState((previous) => ({
          ...previous,
          isRefining: false,
          error,
        }));
        onError?.(error);
        throw error;
      }
    },
    [onError, onSuccess]
  );

  const reset = useCallback(() => {
    setState(createInitialState());
  }, []);

  return {
    isRefining: state.isRefining,
    result: state.result,
    error: state.error,
    refine,
    reset,
  };
}
