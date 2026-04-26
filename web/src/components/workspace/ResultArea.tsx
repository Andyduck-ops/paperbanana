import { useState } from 'react';
import { useLanguage } from '../../hooks';
import type { Artifact } from '../ArtifactPreview';
import type { StageState } from '../../hooks/useGenerate';

export interface ResultAreaProps {
  mode: 'generate' | 'refine';
  result: {
    sessionId: string;
    artifacts: Artifact[];
  } | null;
  refineResult?: Artifact | null;
  stages?: StageState[];
  onExport?: (artifact: Artifact) => void;
  onCopy?: (artifact: Artifact) => void;
  onNewGeneration?: () => void;
}

/**
 * Result Area Component
 *
 * Displays generation/refinement results with:
 * - Current preferred result (visually dominant)
 * - Alternative candidates (if any)
 * - Next-step actions
 * - Clear visual hierarchy
 *
 * Design principles:
 * - Distinguish between current result and alternatives
 * - Actions are clearly available but not intrusive
 * - Stage completion summary visible
 * - Celebration for successful completion
 */
export function ResultArea({
  mode,
  result,
  refineResult,
  stages,
  onExport,
  onCopy,
  onNewGeneration,
}: ResultAreaProps) {
  const { t } = useLanguage();
  const [selectedArtifactIndex, setSelectedArtifactIndex] = useState(0);

  // Determine which result to display
  const displayResult = mode === 'refine' && refineResult
    ? { sessionId: 'refine-result', artifacts: [refineResult] }
    : result;

  if (!displayResult) return null;

  const { sessionId, artifacts } = displayResult;
  const hasMultipleArtifacts = artifacts.length > 1;
  const selectedArtifact = artifacts[selectedArtifactIndex] || artifacts[0];

  // Check if all stages completed (for celebration)
  const allStagesComplete = stages && stages.length > 0 && stages.every(s => s.status === 'complete');

  return (
    <div className="result-area space-y-6">
      {/* Header with session info */}
      <div className="result-area__header flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-status-success/15 flex items-center justify-center">
            <svg className="w-5 h-5 text-status-success" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <div className="flex items-center gap-2">
            <h2 className="text-xl font-heading font-semibold text-foreground">
              {mode === 'refine' ? t('refine.title') : t('generate.result')}
            </h2>
            {allStagesComplete && (
              <span className="inline-flex items-center gap-1 text-xs px-2.5 py-1 rounded-full bg-status-success/10 text-status-success font-medium">
                <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                {t('generate.completed')}
              </span>
            )}
          </div>
        </div>

        <span className="text-xs text-muted-foreground font-mono px-2 py-1 rounded-md bg-muted/50">
          {sessionId.slice(0, 8)}
        </span>
      </div>

      {/* Stage summary (if available) */}
      {stages && stages.length > 0 && (
        <div className="result-area__stages flex flex-wrap gap-2">
          {stages.map((stage) => (
            <div
              key={stage.stage}
              className={`
                inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium
                ${stage.status === 'complete'
                  ? 'bg-status-success/10 text-status-success'
                  : stage.status === 'error'
                  ? 'bg-status-error/10 text-status-error'
                  : stage.status === 'running'
                  ? 'bg-status-running/10 text-status-running'
                  : 'bg-muted text-muted-foreground'}
              `}
            >
              <span>{stage.agent}</span>
              {stage.artifactCount && stage.artifactCount > 0 && (
                <span className="opacity-70">({stage.artifactCount})</span>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Main result display */}
      <div className="result-area__main">
        <div className="rounded-2xl border border-border/70 bg-card/80 overflow-hidden shadow-lg">
          {/* Main artifact */}
          <div className="relative bg-muted/30">
            {selectedArtifact?.data ? (
              <img
                src={selectedArtifact.data}
                alt={selectedArtifact.summary || t('accessibility.generatedResult')}
                className="w-full h-auto max-h-[60vh] object-contain mx-auto"
              />
            ) : (
              <div className="aspect-video flex items-center justify-center text-muted-foreground">
                <div className="text-center">
                  <svg className="w-12 h-12 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  <p>{mode === 'refine' ? t('accessibility.refinedImage') : t('accessibility.generatedResult')}</p>
                </div>
              </div>
            )}

            {/* Artifact info overlay */}
            {selectedArtifact?.summary && (
              <div className="absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-black/60 to-transparent">
                <p className="text-white text-sm">{selectedArtifact.summary}</p>
              </div>
            )}
          </div>

          {/* Artifact selector (if multiple) */}
          {hasMultipleArtifacts && (
            <div className="p-4 border-t border-border/50">
              <p className="text-sm text-muted-foreground mb-2">
                {t('generate.artifacts')}: {artifacts.length}
              </p>
              <div className="flex gap-2 overflow-x-auto">
                {artifacts.map((artifact, index) => (
                  <button
                    key={index}
                    onClick={() => setSelectedArtifactIndex(index)}
                    aria-pressed={selectedArtifactIndex === index}
                    className={`
                      flex-shrink-0 w-20 h-20 rounded-lg border overflow-hidden
                      transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
                      ${selectedArtifactIndex === index
                        ? 'border-primary ring-2 ring-primary/30'
                        : 'border-border/50 hover:border-primary/30'
                      }
                    `}
                  >
                    {artifact.data ? (
                      <img
                        src={artifact.data}
                        alt={t('accessibility.artifact', { number: index + 1 })}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-xs text-muted-foreground">
                        #{index + 1}
                      </div>
                    )}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="result-area__actions flex flex-wrap gap-3">
        {selectedArtifact && (
          <>
            <button
              onClick={() => onExport?.(selectedArtifact)}
              className="
                inline-flex items-center gap-2 px-5 py-2.5 rounded-xl
                bg-primary text-background font-medium text-sm
                hover:opacity-90 active:scale-[0.98] transition-all duration-200
                shadow-lg shadow-primary/20
                focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
              "
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              {t('export.download')}
            </button>

            <button
              onClick={() => onCopy?.(selectedArtifact)}
              className="
                inline-flex items-center gap-2 px-5 py-2.5 rounded-xl
                border border-border/70 bg-background text-foreground font-medium text-sm
                hover:bg-muted/60 hover:border-primary/30 active:scale-[0.98] transition-all duration-200
                focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
              "
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
              {t('export.copy')}
            </button>

            {mode === 'generate' && (
              <button
                onClick={() => {}}
                className="
                  inline-flex items-center gap-2 px-5 py-2.5 rounded-xl
                  border border-border/70 bg-background text-foreground font-medium text-sm
                  hover:bg-muted/60 hover:border-primary/30 active:scale-[0.98] transition-all duration-200
                  focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
                "
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
                {t('refine.title')}
              </button>
            )}
          </>
        )}

        <div className="flex-1"></div>

        <button
          onClick={onNewGeneration}
          className="
            inline-flex items-center gap-2 px-4 py-2.5 rounded-xl
            text-muted-foreground hover:text-foreground font-medium text-sm
            hover:bg-muted/40 active:scale-[0.98] transition-all duration-200
            focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
          "
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          {t('generate.new')}
        </button>
      </div>
    </div>
  );
}

export default ResultArea;
