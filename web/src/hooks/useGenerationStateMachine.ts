/**
 * useGenerationStateMachine
 * 
 * A hook that uses the generation state machine to manage
 * all generation-related state in one place.
 * 
 * This eliminates the "tacit knowledge" problem by:
 * 1. Centralizing all state transitions
 * 2. Making state changes explicit and testable
 * 3. Eliminating distributed state across multiple hooks
 */

import { useReducer, useCallback, useEffect, useRef } from 'react';
import {

  generationReducer,
  createInitialState,
  GenerationSelectors,
} from '../lib/generationStateMachine';
import { useAppStore } from '../stores';
import { useToast } from './useToast';
import { useGenerate, useBatchGeneration, useRefine } from './index';
import { imageSourceToFile } from '../lib/imageUtils';

export interface UseGenerationStateMachineOptions {
  onGenerateSuccess?: () => void;
  onRefineSuccess?: () => void;
  onError?: (error: string) => void;
}

export function useGenerationStateMachine(options: UseGenerationStateMachineOptions = {}) {
  const { onGenerateSuccess, onRefineSuccess, onError } = options;

  const { addToast } = useToast();
  
  // UI state from app store
  const setMainTab = useAppStore((state) => state.setMainTab);
  const setExamplePrompt = useAppStore((state) => state.setExamplePrompt);
  
  // Generation state machine
  const [state, dispatch] = useReducer(generationReducer, null, createInitialState);
  
  // Underlying hooks - their state is synced to the machine
  const generate = useGenerate();
  const batch = useBatchGeneration();
  const refine = useRefine({
    onSuccess: () => {
      addToast("Image refined successfully", "success");
      dispatch({ type: 'COMPLETE_REFINE', sessionId: '' });
      onRefineSuccess?.();
    },
    onError: (err) => {
      const errorMsg = err.message || "Refinement failed";
      addToast(errorMsg, "error");
      dispatch({ type: 'FAIL', error: errorMsg });
      onError?.(errorMsg);
    },
  });
  
  // Sync external hook states to machine (one-way sync)
  const prevGenerateRef = useRef(generate.isGenerating);
  const prevBatchRef = useRef(batch.isGenerating);
  const prevRefineRef = useRef(refine.isRefining);
  
  useEffect(() => {
    // Detect completion transitions
    if (prevGenerateRef.current && !generate.isGenerating && state.mode === 'single') {
      if (generate.error) {
        dispatch({ type: 'FAIL', error: generate.error });
        onError?.(generate.error);
      } else {
        dispatch({ type: 'COMPLETE_SINGLE', sessionId: generate.result?.sessionId || '' });
        onGenerateSuccess?.();
      }
    }
    prevGenerateRef.current = generate.isGenerating;
  }, [generate.isGenerating, generate.error, generate.result, state.mode, onGenerateSuccess, onError]);
  
  useEffect(() => {
    if (prevBatchRef.current && !batch.isGenerating && state.mode === 'batch') {
      if (batch.error) {
        dispatch({ type: 'FAIL', error: batch.error });
        onError?.(batch.error);
      } else {
        dispatch({ type: 'COMPLETE_BATCH', batchId: batch.result?.batchId || '' });
        onGenerateSuccess?.();
      }
    }
    prevBatchRef.current = batch.isGenerating;
  }, [batch.isGenerating, batch.error, batch.result, state.mode, onGenerateSuccess, onError]);
  
  useEffect(() => {
    prevRefineRef.current = refine.isRefining;
  }, [refine.isRefining]);
  
  // Actions - explicit state transitions
  const startGeneration = useCallback(async (
    prompt: string,
    options?: {
      content?: string;
      visualIntent?: string;
      visualizerNode?: string;
      numCandidates?: number;
      config?: any;
    }
  ) => {
    if (!GenerationSelectors.canStartNew(state)) {
      addToast("Please wait for current operation to complete", "info");
      return;
    }
    
    setMainTab("generate");
    
    if (options?.numCandidates && options.numCandidates > 1) {
      // Batch generation
      dispatch({ 
        type: 'START_BATCH', 
        prompt, 
        batchId: `batch-${Date.now()}`,
        count: options.numCandidates 
      });
      
      await batch.startBatch(prompt, options.numCandidates, {
        content: options.content,
        visualIntent: options.visualIntent,
        visualizerNode: options.visualizerNode,
        config: options.config,
      });
    } else {
      // Single generation
      const sessionId = `gen-${Date.now()}`;
      dispatch({ type: 'START_SINGLE', prompt, sessionId });
      
      await generate.generate(prompt, {
        content: options?.content,
        visualIntent: options?.visualIntent,
        visualizerNode: options?.visualizerNode,
        config: options?.config,
      });
    }
  }, [state, setMainTab, generate.generate, batch.startBatch, addToast]);
  
  const startRefine = useCallback(async (params: {
    imageData: string;
    instructions: string;
    resolution: "2K" | "4K";
    enableIteration?: boolean;
    maxIterations?: number;
  }) => {
    if (!GenerationSelectors.canStartNew(state)) {
      addToast("Please wait for current operation to complete", "info");
      return;
    }
    
    dispatch({ 
      type: 'START_REFINE', 
      imageData: params.imageData, 
      instructions: params.instructions 
    });
    
    await refine.refine({
      image: {
        file: await imageSourceToFile(params.imageData, "refine-input.png"),
        previewUrl: params.imageData,
      },
      instructions: params.instructions,
      resolution: params.resolution,
      enable_iteration: params.enableIteration,
      max_iterations: params.enableIteration ? params.maxIterations : 1,
    });
  }, [state, refine.refine, addToast]);
  
  const cancel = useCallback(() => {
    generate.cancel?.();
    dispatch({ type: 'CANCEL' });
  }, [generate.cancel]);
  
  const resetForNew = useCallback(() => {
    generate.reset();
    batch.resetBatch();
    refine.reset();
    dispatch({ type: 'RESET_FOR_NEW' });
  }, [generate.reset, batch.resetBatch, refine.reset]);
  
  const resetAll = useCallback(() => {
    generate.reset();
    batch.resetBatch();
    refine.reset();
    dispatch({ type: 'RESET_ALL' });
  }, [generate.reset, batch.resetBatch, refine.reset]);
  
  const selectCandidate = useCallback((candidateId: string) => {
    dispatch({ type: 'SELECT_CANDIDATE', candidateId });
  }, []);
  
  const loadExample = useCallback((prompt: string) => {
    resetForNew();
    setExamplePrompt(prompt);
    setMainTab("generate");
  }, [resetForNew, setExamplePrompt, setMainTab]);
  
  const clearError = useCallback(() => {
    dispatch({ type: 'CLEAR_ERROR' });
  }, []);
  
  return {
    // State (from machine)
    mode: state.mode,
    isGenerating: state.isGenerating,
    isBatchGenerating: state.isBatchGenerating,
    isRefining: state.isRefining,
    isActive: GenerationSelectors.isActive(state),
    canStartNew: GenerationSelectors.canStartNew(state),
    currentSessionId: state.currentSessionId,
    currentBatchId: state.currentBatchId,
    selectedCandidateId: state.selectedCandidateId,
    pendingPrompt: state.pendingPrompt,
    refineSeedImage: state.refineSeedImage,
    error: state.error,
    
    // Derived data from underlying hooks
    stages: generate.stages,
    result: generate.result,
    batchResult: batch.result,
    batchProgress: batch.progress,
    refineResult: refine.result,
    
    // Actions
    startGeneration,
    startRefine,
    cancel,
    resetForNew,
    resetAll,
    selectCandidate,
    loadExample,
    clearError,
    
    // For testing/debugging
    _dispatch: dispatch,
    _state: state,
  };
}
