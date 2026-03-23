import { useLanguage } from '../hooks';

export type StageStatus = 'pending' | 'running' | 'complete' | 'error';

export interface StageCardProps {
  stage: string;
  agent: string;
  status: StageStatus;
  summary?: string;
  error?: string;
  artifactCount?: number;
}

export function StageCard({
  stage,
  agent,
  status,
  summary,
  error,
  artifactCount,
}: StageCardProps) {
  const { t } = useLanguage();

  const statusColors = {
    pending: 'bg-muted text-muted-foreground',
    running: 'bg-primary/20 text-primary animate-pulse',
    complete: 'bg-green-500/20 text-green-600',
    error: 'bg-red-500/20 text-red-600',
  };

  const statusIcons = {
    pending: '○',
    running: '◆',
    complete: '✓',
    error: '✗',
  };

  return (
    <div className={`p-4 rounded-lg border border-border ${statusColors[status]}`}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-lg">{statusIcons[status]}</span>
          <div>
            <h4 className="font-medium">{agent}</h4>
            <p className="text-sm opacity-75">{stage}</p>
          </div>
        </div>
        {artifactCount !== undefined && artifactCount > 0 && (
          <span className="text-xs px-2 py-1 rounded bg-background/50">
            {artifactCount} {t('generate.artifacts')}
          </span>
        )}
      </div>
      {summary && status === 'complete' && (
        <p className="mt-2 text-sm opacity-80">{summary}</p>
      )}
      {error && status === 'error' && (
        <p className="mt-2 text-sm font-medium">{error}</p>
      )}
    </div>
  );
}
