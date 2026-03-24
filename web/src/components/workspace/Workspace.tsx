import { useState, useCallback } from 'react';
import { useLanguage } from '../../hooks';
import { ModeSwitcher, type WorkspaceMode } from './ModeSwitcher';
import { EmptyState } from './EmptyState';
import { CandidateGrid, type Candidate } from './CandidateGrid';
import { ResultArea } from './ResultArea';
import type { Artifact } from '../ArtifactPreview';
import type { StageState } from '../../hooks/useGenerate';

export interface WorkspaceProps {
  // Mode state
  mode: WorkspaceMode;
  onModeChange: (mode: WorkspaceMode) => void;

  // Generation state
  isGenerating: boolean;
  stages: StageState[];
  result: {
    sessionId: string;
    artifacts: Artifact[];
  } | null;
  error: string | null;

  // Batch state
  isBatchGenerating?: boolean;
  batchCandidates?: Candidate[];
  batchProgress?: {
    batchId: string;
    total: number;
    completed: number;
    failed: number;
  } | null;

  // Refinement state
  refineResult?: Artifact | null;
  isRefining?: boolean;

  // Input components
  generateInput: React.ReactNode;
  refineInput: React.ReactNode;

  // Actions
  onExport?: (artifact: Artifact) => void;
  onCopy?: (artifact: Artifact) => void;
  onNewGeneration?: () => void;
  onSelectCandidate?: (candidateId: string) => void;
  onRefineCandidate?: (candidateId: string) => void;
  onDeleteCandidate?: (candidateId: string) => void;
}

/**
 * Main Workspace Component
 *
 * Implements the core workspace layout with:
 * - Mode switching between Generation and Refinement
 * - Empty state with capability showcase
 * - Multi-candidate comparison view
 * - Result area with hierarchical display
 *
 * Design principles:
 * - Main stage is visually dominant
 * - Secondary elements are subdued
 * - Clear visual hierarchy between current result and alternatives
 */
export function Workspace({
  mode,
  onModeChange,
  isGenerating,
  stages,
  result,
  error,
  isBatchGenerating,
  batchCandidates,
  batchProgress,
  refineResult,
  isRefining,
  generateInput,
  refineInput,
  onExport,
  onCopy,
  onNewGeneration,
  onSelectCandidate,
  onRefineCandidate,
  onDeleteCandidate,
}: WorkspaceProps) {
  const { t } = useLanguage();
  const [selectedCandidateId, setSelectedCandidateId] = useState<string | null>(null);

  // Determine if we should show empty state
  const showEmptyState = !isGenerating &&
    !result &&
    !refineResult &&
    !isBatchGenerating &&
    !batchProgress &&
    !error;

  // Determine if we should show result area
  const showResultArea = result || refineResult;

  // Determine if we should show candidate grid
  const showCandidateGrid = batchCandidates && batchCandidates.length > 0;

  // Handle candidate selection
  const handleSelectCandidate = useCallback((candidateId: string) => {
    setSelectedCandidateId(candidateId);
    onSelectCandidate?.(candidateId);
  }, [onSelectCandidate]);

  // Handle empty state action
  const handleEmptyStateAction = useCallback((action: 'generate' | 'example') => {
    if (action === 'generate') {
      onModeChange('generate');
    }
    // Examples would be handled by the parent component
  }, [onModeChange]);

  return (
    <div className="workspace-container h-full flex flex-col">
      {/* Mode Switcher - Lightweight, always visible */}
      <div className="workspace-mode-bar flex items-center justify-between px-6 py-3 border-b border-border/50 bg-background/50 backdrop-blur-sm">
        <ModeSwitcher
          mode={mode}
          onChange={onModeChange}
          disabled={isGenerating || isRefining || isBatchGenerating}
        />
        {(isGenerating || isRefining || isBatchGenerating) && (
          <span className="text-xs text-muted-foreground animate-pulse">
            {isGenerating ? t('generate.generating') :
             isRefining ? t('refine.refining') :
             t('generate.batchMode')}
          </span>
        )}
      </div>

      {/* Main Workspace Area */}
      <div className="workspace-main flex-1 overflow-auto p-6">
        {/* Error Display */}
        {error && (
          <div className="workspace-error mb-6 rounded-2xl border border-red-500/25 bg-red-500/10 p-4 text-red-700">
            <div className="flex items-center gap-2">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="font-medium">{error}</span>
            </div>
          </div>
        )}

        {/* Empty State */}
        {showEmptyState && (
          <EmptyState
            mode={mode}
            onAction={handleEmptyStateAction}
          />
        )}

        {/* Input Area - Only show when not generating and no results */}
        {!showEmptyState && !showResultArea && !isGenerating && !isBatchGenerating && !refineResult && (
          <div className="workspace-input-area max-w-4xl mx-auto">
            {mode === 'generate' ? generateInput : refineInput}
          </div>
        )}

        {/* Progress Display */}
        {(isGenerating || isRefining) && stages.length > 0 && (
          <div className="workspace-progress max-w-2xl mx-auto">
            <div className="space-y-3">
              {stages.map((stage) => (
                <div
                  key={stage.stage}
                  className={`flex items-center gap-3 p-3 rounded-xl border transition-all ${
                    stage.status === 'running'
                      ? 'border-primary/30 bg-primary/5'
                      : stage.status === 'complete'
                      ? 'border-status-success/30 bg-status-success/5'
                      : stage.status === 'error'
                      ? 'border-status-error/30 bg-status-error/5'
                      : 'border-border/50 bg-muted/30'
                  }`}
                >
                  <div className={`w-8 h-8 rounded-full flex items-center justify-center ${
                    stage.status === 'running'
                      ? 'bg-primary/20 text-primary'
                      : stage.status === 'complete'
                      ? 'bg-status-success/20 text-status-success'
                      : stage.status === 'error'
                      ? 'bg-status-error/20 text-status-error'
                      : 'bg-muted text-muted-foreground'
                  }`}>
                    {stage.status === 'complete' ? (
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    ) : stage.status === 'running' ? (
                      <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
                    ) : stage.status === 'error' ? (
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    ) : (
                      <span className="text-xs">{stage.stage[0].toUpperCase()}</span>
                    )}
                  </div>
                  <div className="flex-1">
                    <div className="font-medium text-foreground">{stage.agent}</div>
                    {stage.summary && (
                      <div className="text-xs text-muted-foreground">{stage.summary}</div>
                    )}
                  </div>
                  <div className="text-xs text-muted-foreground capitalize">{stage.status}</div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Batch Progress */}
        {isBatchGenerating && batchProgress && (
          <div className="workspace-batch-progress max-w-2xl mx-auto">
            <div className="rounded-2xl border border-border/70 bg-card/80 p-6 shadow-lg">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-foreground">{t('generate.batchProgress')}</h3>
                <span className="text-sm text-muted-foreground">
                  {batchProgress.completed} / {batchProgress.total}
                </span>
              </div>
              <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                <div
                  className="h-full bg-primary transition-all duration-300"
                  style={{ width: `${(batchProgress.completed / batchProgress.total) * 100}%` }}
                />
              </div>
              <div className="mt-4 flex gap-4 text-sm">
                <div className="flex items-center gap-1">
                  <div className="w-2 h-2 rounded-full bg-status-success" />
                  <span className="text-muted-foreground">{batchProgress.completed} {t('generate.successful')}</span>
                </div>
                {batchProgress.failed > 0 && (
                  <div className="flex items-center gap-1">
                    <div className="w-2 h-2 rounded-full bg-status-error" />
                    <span className="text-muted-foreground">{batchProgress.failed} {t('generate.failed')}</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Result Area */}
        {showResultArea && (
          <div className="workspace-result-area">
            <ResultArea
              mode={mode}
              result={result}
              refineResult={refineResult}
              stages={stages}
              onExport={onExport}
              onCopy={onCopy}
              onNewGeneration={onNewGeneration}
            />
          </div>
        )}

        {/* Candidate Grid - Multi-candidate comparison */}
        {showCandidateGrid && (
          <div className="workspace-candidates mt-6">
            <CandidateGrid
              candidates={batchCandidates}
              selectedId={selectedCandidateId}
              onSelect={handleSelectCandidate}
              onRefine={onRefineCandidate}
              onDelete={onDeleteCandidate}
              onExport={onExport}
            />
          </div>
        )}
      </div>
    </div>
  );
}

export default Workspace;
