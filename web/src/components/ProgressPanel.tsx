import { StageCard } from './StageCard';
import { EvolutionTimeline, type EvolutionStage } from './EvolutionTimeline';
import { useLanguage } from '../hooks';
import type { StageState, ResumeMetadata } from '../hooks/useGenerate';

export interface ProgressPanelProps {
  stages: StageState[];
  isVisible?: boolean;
  pipelineMode?: 'full' | 'planner-critic';
  isGenerating?: boolean;
  onCancel?: () => void;
  estimatedTime?: number;
  resumeMetadata?: ResumeMetadata;
}

function formatEstimatedTime(seconds: number): string {
  if (seconds < 60) {
    return `~${Math.round(seconds)}s`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = Math.round(seconds % 60);
  if (remainingSeconds === 0) {
    return `~${minutes}m`;
  }
  return `~${minutes}m ${remainingSeconds}s`;
}

export function ProgressPanel({ stages, isVisible = true, pipelineMode = 'full', isGenerating = false, onCancel, estimatedTime }: ProgressPanelProps) {
  const { t } = useLanguage();

  if (!isVisible || stages.length === 0) {
    return null;
  }

  const completedCount = stages.filter(s => s.status === 'complete').length;
  const totalCount = stages.length;
  const progressPercent = totalCount > 0 ? (completedCount / totalCount) * 100 : 0;

  const buildEvolutionStages = (): EvolutionStage[] => {
    return stages.map((stage) => ({
      name: `${getStageLabel(stage.stage)}: ${stage.agent}`,
      status: stage.status,
      description: stage.summary || stage.error,
    }));
  };

  const getStageLabel = (stage: string): string => {
    return stage.charAt(0).toUpperCase() + stage.slice(1);
  };

  return (
    <div className="space-y-5">
      {/* Header with progress bar */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
              <svg className="w-4 h-4 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <h3 className="text-lg font-heading font-semibold text-foreground">
              {t('generate.progress')}
            </h3>
          </div>
          <div className="flex items-center gap-3">
            {estimatedTime !== undefined && estimatedTime > 0 && (
              <span className="text-xs text-muted-foreground font-mono">
                Est. {formatEstimatedTime(estimatedTime)}
              </span>
            )}
            <span className="text-sm font-medium text-muted-foreground px-2 py-1 rounded-md bg-muted/50">
              {completedCount}/{totalCount}
            </span>
            {isGenerating && onCancel && (
              <button
                type="button"
                onClick={onCancel}
                className="px-3 py-1.5 text-xs font-medium text-status-error hover:text-status-error/80 hover:bg-status-error/10 rounded-lg transition-colors border border-status-error/20"
                aria-label="Cancel"
              >
                Cancel
              </button>
            )}
          </div>
        </div>

        {/* Progress bar */}
        <div className="w-full h-1.5 bg-muted rounded-full overflow-hidden">
          <div
            className="h-full bg-primary rounded-full transition-all duration-500 ease-out"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
      </div>

      {/* Stage cards */}
      <div className="space-y-2.5">
        {stages.map((stage) => (
          <StageCard
            key={stage.stage}
            {...stage}
          />
        ))}
      </div>

      <EvolutionTimeline
        stages={buildEvolutionStages()}
        pipelineMode={pipelineMode}
      />
    </div>
  );
}
