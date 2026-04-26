import { useState } from 'react';
import { Workspace, type WorkspaceMode, type Candidate } from '../components/workspace';
import { GeneratePanel, RefinePanel, type Artifact } from '../components';
import type { GenerateOptions as PanelGenerateOptions } from '../components';
import type { RefineRequest as ApiRefineRequest } from '../types/api';
import { useGenerate, useRefine, useBatchGeneration, useToast } from '../hooks';
import { copyImageToClipboard } from '../lib/clipboard';

// Type for RefinePanel callback
interface PanelRefineRequest {
  imageData: string;
  instructions: string;
  resolution: '2K' | '4K';
}

// Helper to convert base64 data URL to File
function dataUrlToFile(dataUrl: string, filename: string): File {
  const arr = dataUrl.split(',');
  const mime = arr[0].match(/:(.*?);/)?.[1] || 'image/png';
  const bstr = atob(arr[1]);
  let n = bstr.length;
  const u8arr = new Uint8Array(n);
  while (n--) {
    u8arr[n] = bstr.charCodeAt(n);
  }
  return new File([u8arr], filename, { type: mime });
}

/**
 * Workspace Page
 *
 * Demonstrates the new Workspace component with:
 * - Generate/Refine mode switching
 * - Empty state with example prompts
 * - Multi-candidate comparison
 * - Result area with hierarchical display
 */
export function WorkspacePage() {
  const [mode, setMode] = useState<WorkspaceMode>('generate');

  const { addToast } = useToast();

  const {
    isGenerating,
    stages,
    result,
    error,
    generate,
    cancel,
    reset,
  } = useGenerate();

  const {
    isGenerating: isBatchGenerating,
    progress: batchProgress,
    result: batchResult,
    error: batchError,
    startBatch,
    resetBatch,
  } = useBatchGeneration();

  const {
    isRefining,
    result: refineResult,
    refine,
    reset: resetRefine,
  } = useRefine({
    onSuccess: () => {
      addToast('Image refined successfully', 'success');
    },
    onError: (refineError) => {
      addToast(refineError.message || 'Refinement failed', 'error');
    },
  });

  // Convert batch result to candidates for the grid
  // Note: Candidate.id is string, BatchCandidate.candidateId is number
  const batchCandidates: Candidate[] = batchResult?.candidates.map((c) => ({
    id: String(c.candidateId),
    index: c.candidateId,
    status: c.status === 'completed' ? 'completed' : c.status === 'failed' ? 'failed' : 'running',
    artifacts: c.artifacts?.map((a) => ({
      kind: a.kind || 'image',
      mimeType: a.mime_type || 'image/png',
      summary: '',
      data: '',
    })),
    error: c.error,
  })) || [];

  // Handle generate from GeneratePanel
  const handleGenerate = async (prompt: string, options?: PanelGenerateOptions) => {
    const config = options?.config ? {
      aspect_ratio: options.config.aspectRatio,
      critic_rounds: options.config.criticRounds,
      retrieval_mode: options.config.retrievalMode,
      pipeline_mode: options.config.pipelineMode,
      query_model: options.config.queryModel,
      gen_model: options.config.genModel,
    } : undefined;

    if (options?.numCandidates && options.numCandidates > 1) {
      await startBatch(prompt, options.numCandidates, {
        visualizerNode: options.visualizerNode,
        content: options?.content,
        visualIntent: options?.visualIntent,
        config,
      });
    } else {
      await generate(prompt, {
        visualizerNode: options?.visualizerNode,
        content: options?.content,
        visualIntent: options?.visualIntent,
        config,
      });
    }
  };

  // Handle refine from RefinePanel
  const handleRefine = (request: PanelRefineRequest) => {
    const apiRequest: ApiRefineRequest = {
      image: {
        file: dataUrlToFile(request.imageData, 'refine-image.png'),
        previewUrl: request.imageData,
      },
      instructions: request.instructions,
      resolution: request.resolution,
    };
    refine(apiRequest);
  };

  const handleExport = (artifact: Artifact) => {
    // Export modal would be shown here
    void artifact;
  };

  const handleCopy = async (artifact: Artifact) => {
    if (artifact.data) {
      const success = await copyImageToClipboard(artifact.data);
      if (success) {
        addToast('Image copied to clipboard', 'success');
      } else {
        addToast('Failed to copy image', 'error');
      }
    }
  };

  const handleNewGeneration = () => {
    reset();
    resetBatch();
    resetRefine();
  };

  const handleSelectCandidate = (_candidateId: string) => {
  };

  const handleRefineCandidate = (candidateId: string) => {
    setMode('refine');
    void candidateId;
  };

  const handleDeleteCandidate = (_candidateId: string) => {
  };

  return (
    <div className="h-screen flex flex-col">
      <Workspace
        mode={mode}
        onModeChange={setMode}
        isGenerating={isGenerating}
        stages={stages}
        result={result}
        error={error || batchError}
        isBatchGenerating={isBatchGenerating}
        batchCandidates={batchCandidates.length > 0 ? batchCandidates : undefined}
        batchProgress={batchProgress ? {
          batchId: batchProgress.batchId || '',
          total: batchProgress.candidates?.length || 0,
          completed: batchProgress.successful || 0,
          failed: batchProgress.failed || 0,
        } : null}
        batchCompleted={batchProgress?.status === 'completed'}
        refineResult={refineResult?.image ? {
          kind: 'image',
          mimeType: refineResult.image.mimeType,
          summary: 'Refined image',
          data: refineResult.image.data,
        } : null}
        isRefining={isRefining}
        onCancel={cancel}
        generateInput={
          <GeneratePanel
            onGenerate={handleGenerate}
            isGenerating={isGenerating || isBatchGenerating}
            collapsed={isGenerating || isBatchGenerating}
          />
        }
        refineInput={
          <RefinePanel
            onRefine={handleRefine}
            isRefining={isRefining}
          />
        }
        onExport={handleExport}
        onCopy={handleCopy}
        onNewGeneration={handleNewGeneration}
        onSelectCandidate={handleSelectCandidate}
        onRefineCandidate={handleRefineCandidate}
        onDeleteCandidate={handleDeleteCandidate}
      />
    </div>
  );
}

export default WorkspacePage;
