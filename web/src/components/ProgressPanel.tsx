import { StageCard, type StageStatus } from './StageCard';
import { EvolutionTimeline, type EvolutionStage } from './EvolutionTimeline';
import { useLanguage } from '../hooks';

export interface StageState {
  stage: string;
  agent: string;
  status: StageStatus;
  summary?: string;
  error?: string;
  artifactCount?: number;
}

export interface ProgressPanelProps {
  stages: StageState[];
  isVisible?: boolean;
  pipelineMode?: 'full' | 'planner-critic';
}

export function ProgressPanel({ stages, isVisible = true, pipelineMode = 'full' }: ProgressPanelProps) {
  const { t } = useLanguage();

  if (!isVisible || stages.length === 0) {
    return null;
  }

  const completedCount = stages.filter(s => s.status === 'complete').length;
  const totalCount = stages.length;

  const buildEvolutionStages = (): EvolutionStage[] => {
    return stages.map((stage) => ({
      name: `${getStageEmoji(stage.stage)} ${stage.agent}`,
      status: stage.status,
      description: stage.summary || stage.error,
    }));
  };

  const getStageEmoji = (stage: string): string => {
    switch (stage) {
      case 'retriever': return '🔍';
      case 'planner': return '📋';
      case 'visualizer': return '🎨';
      case 'stylist': return '✨';
      case 'critic': return '🔍';
      default: return '○';
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-heading text-foreground">
          {t('generate.progress')}
        </h3>
        <span className="text-sm text-muted-foreground">
          {completedCount}/{totalCount}
        </span>
      </div>
      <div className="space-y-2">
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
