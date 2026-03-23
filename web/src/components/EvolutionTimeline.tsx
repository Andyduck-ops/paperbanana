import { useState } from 'react';
import { useLanguage } from '../hooks';

export interface EvolutionStage {
  name: string;
  status: 'pending' | 'running' | 'complete' | 'error';
  description?: string;
  image?: string;
  suggestions?: string;
}

export interface EvolutionTimelineProps {
  stages: EvolutionStage[];
  currentStage?: number;
  pipelineMode?: 'full' | 'planner-critic';
}

export function EvolutionTimeline({
  stages,
  currentStage: _currentStage = -1,
  pipelineMode: _pipelineMode = 'full'
}: EvolutionTimelineProps) {
  // currentStage and pipelineMode are reserved for future filtering functionality
  void _currentStage;
  void _pipelineMode;
  const { t } = useLanguage();
  const [expandedStage, setExpandedStage] = useState<number | null>(null);

  const getStageIcon = (status: EvolutionStage['status']) => {
    switch (status) {
      case 'complete': return '✅';
      case 'running': return '🔄';
      case 'error': return '❌';
      default: return '○';
    }
  };

  if (stages.length === 0) {
    return null;
  }

  return (
    <div className="evolution-timeline space-y-2">
      <h4 className="text-sm font-medium text-foreground mb-3">
        {t('generate.evolutionTimeline')}
      </h4>

      {stages.map((stage, index) => (
        <div key={index} className="stage-item">
          <button
            onClick={() => setExpandedStage(expandedStage === index ? null : index)}
            className="w-full flex items-center gap-2 p-2 rounded border border-border hover:bg-muted/50 text-left"
          >
            <span className="text-lg">{getStageIcon(stage.status)}</span>
            <span className="flex-1 text-sm font-medium text-foreground">
              {stage.name}
            </span>
            <span className="text-xs text-muted-foreground">
              {stage.status === 'running' ? t('generate.running') : ''}
            </span>
          </button>

          {expandedStage === index && stage.description && (
            <div className="mt-1 p-3 bg-muted/30 rounded text-sm text-muted-foreground">
              {stage.description}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
