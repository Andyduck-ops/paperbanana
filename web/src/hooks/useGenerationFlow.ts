import { useCallback, useEffect, useRef } from 'react';
import { useAppStore, useGenerationStore } from '../stores';
import { useGenerate, useBatchGeneration, useRefine } from './index';
import { useToast } from './useToast';
import { useLanguage } from './useLanguage';
import { useHistory } from './useHistory';
import { imageSourceToFile } from '../lib/imageUtils';
import type { Artifact } from '../components/ArtifactPreview';

export interface GenerateOptions {
  content?: string;
  visualIntent?: string;
  visualizerNode?: string;
  numCandidates?: number;
  config?: {
    aspectRatio?: string;
    criticRounds?: number;
    retrievalMode?: string;
    pipelineMode?: string;
    queryModel?: string;
    genModel?: string;
  };
}

export function useGenerationFlow() {
  const { t } = useLanguage();
  const { addToast } = useToast();
  const { restoreSession: restoreHistorySession } = useHistory();

  // Store actions
  const setMainTab = useAppStore((state) => state.setMainTab);
  const setExamplePrompt = useAppStore((state) => state.setExamplePrompt);
  const setSelectedSessionId = useGenerationStore((state) => state.setSelectedSessionId);
  const setSelectedBatchCandidateId = useGenerationStore((state) => state.setSelectedBatchCandidateId);
  const clearSelectedBatchCandidateId = useGenerationStore((state) => state.clearSelectedBatchCandidateId);
  // Note: setRefineSeedImageData used via handleSelectSession
  const clearRefineSeedImageData = useGenerationStore((state) => state.clearRefineSeedImageData);
  const setPendingHistoryContext = useGenerationStore((state) => state.setPendingHistoryContext);
  const setExportArtifact = useGenerationStore((state) => state.setExportArtifact);
  const openModal = useAppStore((state) => state.openModal);

  // Generation hooks
  const { generate, reset, restore, isGenerating, result, stages, error, cancel } = useGenerate();
  const { startBatch, resetBatch, restoreBatch, isGenerating: isBatchGenerating, result: batchResult } = useBatchGeneration();
  const { refine, reset: resetRefine, restore: restoreRefine, isRefining, result: refineResult, error: refineError } = useRefine({
    onSuccess: () => addToast("Image refined successfully", "success"),
    onError: (err) => addToast(err.message || "Refinement failed", "error"),
  });

  const lastRecordedHistoryId = useRef<string | null>(null);

  // Handle example loading
  useEffect(() => {
    const handleLoadExample = (event: CustomEvent<{ prompt: string }>) => {
      setExamplePrompt(event.detail.prompt);
      setMainTab("generate");
      reset();
      resetBatch();
      resetRefine();
      clearSelectedBatchCandidateId();
      clearRefineSeedImageData();
    };

    window.addEventListener('workspace:loadExample', handleLoadExample as EventListener);
    return () => {
      window.removeEventListener('workspace:loadExample', handleLoadExample as EventListener);
    };
  }, [reset, resetBatch, resetRefine, setExamplePrompt, setMainTab, clearSelectedBatchCandidateId, clearRefineSeedImageData]);

  const handleGenerate = useCallback(async (prompt: string, options?: GenerateOptions) => {
    setMainTab("generate");
    clearRefineSeedImageData();
    resetRefine();

    setPendingHistoryContext({
      prompt,
      mode: options?.numCandidates && options.numCandidates > 1 ? "batch" : "generate",
    });

    if (options?.numCandidates && options.numCandidates > 1) {
      reset();
      clearSelectedBatchCandidateId();
      await startBatch(prompt, options.numCandidates, {
        content: options.content,
        visualIntent: options.visualIntent,
        visualizerNode: options.visualizerNode,
        config: options.config as any,
      });
    } else {
      resetBatch();
      clearSelectedBatchCandidateId();
      await generate(prompt, {
        content: options?.content,
        visualIntent: options?.visualIntent,
        visualizerNode: options?.visualizerNode,
        config: options?.config as any,
      });
    }
  }, [generate, startBatch, reset, resetBatch, resetRefine, setMainTab, setPendingHistoryContext, clearSelectedBatchCandidateId, clearRefineSeedImageData]);

  const handleRefine = useCallback(async (request: {
    imageData: string;
    instructions: string;
    resolution: "2K" | "4K";
    enableIteration?: boolean;
    maxIterations?: number;
  }) => {
    setPendingHistoryContext({
      prompt: request.instructions || "Refinement task",
      mode: "refine",
    });
    await refine({
      image: {
        file: await imageSourceToFile(request.imageData, "refine-input.png"),
        previewUrl: request.imageData,
      },
      instructions: request.instructions,
      resolution: request.resolution,
      enable_iteration: request.enableIteration,
      max_iterations: request.enableIteration ? request.maxIterations : 1,
    });
  }, [refine, setPendingHistoryContext]);

  const handleExport = useCallback((artifact: Artifact) => {
    setExportArtifact(artifact);
    openModal('export');
  }, [setExportArtifact, openModal]);

  // Handle history session restoration
  const handleSelectSession = useCallback(async (sessionId: string) => {
    setSelectedSessionId(sessionId);
    addToast(t('history.restore') + '...', "info");

    const restored = await restoreHistorySession(sessionId);
    if (!restored) {
      addToast(t('history.restoreFailed') || "Failed to restore session", "error");
      return;
    }

    setPendingHistoryContext(null);

    if (restored.mode === "batch" && restored.candidates) {
      reset();
      resetRefine();
      clearRefineSeedImageData();
      setMainTab("generate");
      restoreBatch({
        batchId: restored.batchId || restored.id,
        status: "completed",
        candidates: restored.candidates.map((candidate) => ({
          candidateId: candidate.candidateId,
          status: candidate.status === "completed" ? "completed" : "failed",
          artifacts: candidate.artifacts.map((artifact) => ({
            id: artifact.assetId || `${candidate.sessionId}-${artifact.kind}`,
            kind: artifact.kind,
            mime_type: artifact.mimeType,
            uri: artifact.uri || '',
            data: artifact.data,
            asset_id: artifact.assetId,
            metadata: artifact.projectId ? { project_id: artifact.projectId } : undefined,
          })),
          error: candidate.error,
        })),
        successful: restored.successful || 0,
        failed: restored.failed || 0,
        startedAt: restored.startedAt || new Date().toISOString(),
        completedAt: restored.completedAt,
      });
      setSelectedBatchCandidateId(
        restored.candidates.find((c) => c.status === "completed")?.sessionId ||
          restored.candidates[0]?.sessionId ||
          null
      );
      addToast(`${t('history.restore')}: ${sessionId.slice(0, 8)}...`, "success");
      return;
    }

    reset();
    resetBatch();
    clearSelectedBatchCandidateId();

    if (restored.mode === "refine") {
      setMainTab("refine");
      clearRefineSeedImageData();
      restoreRefine({
        sessionId: restored.id,
        status: restored.status === "completed" ? "completed" : "failed",
        content: restored.prompt,
        image: {
          data: restored.artifacts[0]?.data || "",
          mimeType: restored.artifacts[0]?.mimeType || "image/png",
        },
      });
      addToast(`${t('history.restore')}: ${sessionId.slice(0, 8)}...`, "success");
      return;
    }

    setMainTab("generate");
    resetRefine();
    clearRefineSeedImageData();
    restore({
      sessionId: restored.id,
      artifacts: restored.artifacts.map((a) => ({
        kind: a.kind,
        mimeType: a.mimeType,
        summary: a.summary || a.kind,
        data: a.data,
        assetId: a.assetId,
        projectId: a.projectId,
        uri: a.uri,
      })),
      stages: restored.stages || [],
      error: restored.error,
      resumeMetadata: restored.resumeMetadata,
    });
    addToast(`${t('history.restore')}: ${sessionId.slice(0, 8)}...`, "success");
  }, [addToast, reset, resetBatch, resetRefine, restoreBatch, restore, restoreHistorySession, restoreRefine, t, setMainTab, setSelectedSessionId, setSelectedBatchCandidateId, setPendingHistoryContext, clearSelectedBatchCandidateId, clearRefineSeedImageData]);

  const refineArtifact = refineResult
    ? {
        kind: "image" as const,
        mimeType: refineResult.image.mimeType,
        summary: "Refined image",
        data: refineResult.image.data,
      }
    : null;

  return {
    // State - Generation
    isGenerating,
    stages,
    result,
    error,
    cancel,
    
    // State - Batch
    isBatchGenerating,
    batchProgress: batchResult,
    batchResult,
    batchError: batchResult?.status === 'failed' ? 'Batch failed' : null,
    
    // State - Refine
    isRefining,
    refineResult,
    refineError,
    refineArtifact,
    
    // Other
    lastRecordedHistoryId,
    
    // Actions
    handleGenerate,
    handleRefine,
    handleExport,
    handleSelectSession,
    resetGenerate: reset,
    resetBatch,
    resetRefine,
    restoreGenerate: restore,
    restoreBatch,
    restoreRefine,
  };
}
