import { useState, useCallback, memo } from 'react';
import { useLanguage } from '../../hooks';
import { ModeSwitcher, type WorkspaceMode } from './ModeSwitcher';
import { CandidateGrid, type Candidate } from './CandidateGrid';
import { ResultArea } from './ResultArea';
import type { Artifact } from '../ArtifactPreview';
import type { StageState } from '../../hooks/useGenerate';
import { downloadBatchArchive } from '../../types/batch';

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

  // Cancel action
  onCancel?: () => void;

  // Batch state
  isBatchGenerating?: boolean;
  batchCandidates?: Candidate[];
  batchProgress?: {
    batchId: string;
    total: number;
    completed: number;
    failed: number;
  } | null;
  batchCompleted?: boolean;

  // Refinement state
  refineResult?: Artifact | null;
  isRefining?: boolean;

  // Input components
  hero?: React.ReactNode;
  generateInput: React.ReactNode;
  refineInput: React.ReactNode;

  // Actions
  onExport?: (artifact: Artifact) => void;
  onCopy?: (artifact: Artifact) => void;
  onNewGeneration?: () => void;
  onSelectCandidate?: (candidateId: string) => void;
  onRefineCandidate?: (candidateId: string) => void;
  onDeleteCandidate?: (candidateId: string) => void;
  onRefineResult?: () => void;
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
  batchCompleted,
  refineResult,
  isRefining,
  hero,
  generateInput,
  refineInput,
  onExport,
  onCopy,
  onNewGeneration,
  onSelectCandidate,
  onRefineCandidate,
  onDeleteCandidate,
  onCancel,
}: WorkspaceProps) {
  const { t } = useLanguage();
  const [selectedCandidateId, setSelectedCandidateId] = useState<string | null>(null);
  const [selectedCandidateIds, setSelectedCandidateIds] = useState<string[]>([]);
  const [isDownloading, setIsDownloading] = useState(false);
  const [downloadError, setDownloadError] = useState<string | null>(null);
  const [enableMultiSelect, setEnableMultiSelect] = useState(false);

  const handleBatchDownload = useCallback(async () => {
    if (!batchProgress?.batchId) return;
    setIsDownloading(true);
    setDownloadError(null);
    try {
      await downloadBatchArchive(batchProgress.batchId);
    } catch (err) {
      setDownloadError(err instanceof Error ? err.message : 'Download failed');
    } finally {
      setIsDownloading(false);
    }
  }, [batchProgress?.batchId]);

  // Determine if we should show result area
  const showResultArea = result || refineResult;

  // Determine if we should show candidate grid
  const showCandidateGrid = batchCandidates && batchCandidates.length > 0;

  // Determine if we should show input area
  const showInputArea = !isGenerating && !isBatchGenerating && !refineResult && !showResultArea;

  // Handle candidate selection
  const handleSelectCandidate = useCallback((candidateId: string) => {
    setSelectedCandidateId(candidateId);
    onSelectCandidate?.(candidateId);
  }, [onSelectCandidate]);

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
          <div className="workspace-error mb-6 rounded-2xl border border-status-error/25 bg-status-error/10 p-4">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-status-error/15 flex items-center justify-center flex-shrink-0">
                <svg className="w-4 h-4 text-status-error" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <span className="font-medium text-status-error">{error}</span>
            </div>
          </div>
        )}

        {/* Hero Slot - visible with the primary input flow */}
        {showInputArea && hero && (
          <div className="workspace-hero-slot mb-6">
            {hero}
          </div>
        )}

        {/* Input Area - Always visible at top when no results/generation in progress */}
        {showInputArea && (
          <div className="workspace-input-area max-w-3xl mx-auto">
            {mode === 'generate' ? generateInput : refineInput}
          </div>
        )}

        {/* Progress Display */}
        {(isGenerating || isRefining) && stages.length > 0 && (
          <div className="workspace-progress max-w-2xl mx-auto">
            {/* Header with progress bar */}
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-primary/15 flex items-center justify-center">
                  <svg className="w-4 h-4 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-medium text-foreground">
                    {isRefining ? t('refine.refining') : t('generate.generating')}
                  </h3>
                  <div className="w-32 h-1.5 bg-muted rounded-full overflow-hidden mt-1">
                    <div
                      className="h-full bg-primary rounded-full transition-all duration-500"
                      style={{
                        width: `${(stages.filter(s => s.status === 'complete').length / stages.length) * 100}%`
                      }}
                    />
                  </div>
                </div>
              </div>
              {/* Cancel Button */}
              {isGenerating && onCancel && (
                <button
                  onClick={onCancel}
                  className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-status-error hover:text-status-error/80 hover:bg-status-error/10 rounded-lg transition-colors border border-status-error/20 hover:border-status-error/40"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  {t('common.cancel')}
                </button>
              )}
            </div>

            <div className="space-y-3">
              {stages.map((stage) => (
                <div
                  key={stage.stage}
                  className={`flex items-center gap-3 p-3 rounded-xl border transition-all ${
                    stage.status === 'running'
                      ? 'border-primary/30 bg-primary/5 shadow-sm'
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
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-foreground">{stage.agent}</span>
                      {stage.duration !== undefined && stage.status === 'complete' && (
                        <span className="text-xs font-mono text-muted-foreground bg-muted/60 px-1.5 py-0.5 rounded">
                          {stage.duration < 1000
                            ? `${stage.duration}ms`
                            : stage.duration < 60000
                              ? `${(stage.duration / 1000).toFixed(1)}s`
                              : `${Math.floor(stage.duration / 60000)}m ${Math.floor((stage.duration % 60000) / 1000)}s`}
                        </span>
                      )}
                    </div>
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
        {(isBatchGenerating || batchCompleted) && batchProgress && (
          <div className="workspace-batch-progress max-w-2xl mx-auto">
            <div className={`rounded-2xl border p-6 shadow-lg ${
              batchCompleted
                ? 'border-status-success/70 bg-status-success/10'
                : 'border-border/70 bg-card/80'
            }`}>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-foreground">
                  {batchCompleted ? t('generate.batchComplete') : t('generate.batchProgress')}
                </h3>
                <span className="text-sm text-muted-foreground">
                  {batchProgress.completed} / {batchProgress.total}
                </span>
              </div>
              <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                <div
                  className={`h-full transition-all duration-300 ${
                    batchCompleted ? 'bg-status-success' : 'bg-primary'
                  }`}
                  style={{ width: `${(batchProgress.completed / batchProgress.total) * 100}%` }}
                />
              </div>
              <div className="mt-4 flex items-center justify-between">
                <div className="flex gap-4 text-sm">
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
                {batchCompleted && (
                  <button
                    onClick={handleBatchDownload}
                    disabled={isDownloading}
                    className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-primary hover:bg-primary/90 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isDownloading ? (
                      <>
                        <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                        {t('generate.downloading') || 'Downloading...'}
                      </>
                    ) : (
                      <>
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                        </svg>
                        {t('generate.downloadAll') || 'Download All'}
                      </>
                    )}
                  </button>
                )}
              </div>
              {downloadError && (
                <div className="mt-3 p-2 rounded bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm">
                  {downloadError}
                </div>
              )}
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
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-4">
                <label className="flex items-center gap-2 text-sm text-muted-foreground cursor-pointer">
                  <input
                    type="checkbox"
                    checked={enableMultiSelect}
                    onChange={(e) => {
                      setEnableMultiSelect(e.target.checked);
                      if (!e.target.checked) {
                        setSelectedCandidateIds([]);
                      }
                    }}
                    className="rounded border-border"
                  />
                  {t('generate.enableComparison') || 'Enable comparison mode'}
                </label>
                {enableMultiSelect && selectedCandidateIds.length > 0 && (
                  <span className="text-sm text-muted-foreground">
                    {selectedCandidateIds.length} {t('generate.selected') || 'selected'}
                  </span>
                )}
              </div>
            </div>
            <CandidateGrid
              candidates={batchCandidates}
              selectedId={selectedCandidateId}
              selectedIds={selectedCandidateIds}
              onSelect={handleSelectCandidate}
              onMultiSelect={enableMultiSelect ? setSelectedCandidateIds : undefined}
              multiSelect={enableMultiSelect}
              maxSelection={4}
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

const MemoizedWorkspace = memo(Workspace);
MemoizedWorkspace.displayName = 'Workspace';
export default MemoizedWorkspace;
