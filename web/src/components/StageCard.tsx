import { useLanguage } from '../hooks';

export type StageStatus = 'pending' | 'running' | 'complete' | 'error' | 'not_run';

export interface StageCardProps {
  stage: string;
  agent: string;
  status: StageStatus;
  summary?: string;
  error?: string;
  artifactCount?: number;
  onRetry?: () => void;
  isRetrying?: boolean;
  duration?: number;
  startedAt?: string;
  completedAt?: string;
}

function formatDuration(ms?: number): string {
  if (!ms) return '';
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
}

function StageIcon({ stage, className = 'w-4 h-4' }: { stage: string; className?: string }) {
  const iconProps = { className, fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', strokeWidth: 2 };
  switch (stage) {
    case 'retriever':
      return (
        <svg {...iconProps}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      );
    case 'planner':
      return (
        <svg {...iconProps}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
        </svg>
      );
    case 'stylist':
      return (
        <svg {...iconProps}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
        </svg>
      );
    case 'visualizer':
      return (
        <svg {...iconProps}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
      );
    case 'critic':
      return (
        <svg {...iconProps}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      );
    default:
      return (
        <svg {...iconProps}>
          <circle cx="12" cy="12" r="5" />
        </svg>
      );
  }
}

export function StageCard({
  stage,
  agent,
  status,
  summary,
  error,
  artifactCount,
  onRetry,
  isRetrying,
  duration,
}: StageCardProps) {
  const { t } = useLanguage();

  const statusConfig = {
    pending: {
      container: 'border-border/50 bg-muted/30',
      iconBg: 'bg-muted text-muted-foreground',
      text: 'text-muted-foreground',
    },
    running: {
      container: 'border-primary/30 bg-primary/5',
      iconBg: 'bg-primary/20 text-primary',
      text: 'text-primary',
    },
    complete: {
      container: 'border-status-success/30 bg-status-success/5',
      iconBg: 'bg-status-success/20 text-status-success',
      text: 'text-status-success',
    },
    error: {
      container: 'border-status-error/30 bg-status-error/5',
      iconBg: 'bg-status-error/20 text-status-error',
      text: 'text-status-error',
    },
    not_run: {
      container: 'border-border/30 bg-muted/20',
      iconBg: 'bg-muted/50 text-muted-foreground/50',
      text: 'text-muted-foreground/50',
    },
  };

  const config = statusConfig[status];

  return (
    <div className={`p-4 rounded-xl border transition-all duration-300 ${config.container}`}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 ${config.iconBg} ${status === 'running' ? 'animate-pulse' : ''}`}>
            {status === 'complete' ? (
              <>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                <span className="sr-only">✓</span>
              </>
            ) : status === 'running' ? (
              <>
                <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
                <span className="sr-only">◆</span>
              </>
            ) : status === 'error' ? (
              <>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
                <span className="sr-only">✗</span>
              </>
            ) : (
              <>
                <StageIcon stage={stage} />
                <span className="sr-only">{status === 'not_run' ? '—' : '○'}</span>
              </>
            )}
          </div>
          <div>
            <h4 className="font-medium text-foreground">{agent}</h4>
            <p className="text-sm text-muted-foreground">{stage}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {duration !== undefined && status === 'complete' && (
            <span className="text-xs px-2 py-1 rounded-full bg-background/60 text-muted-foreground font-mono">
              {formatDuration(duration)}
            </span>
          )}
          {artifactCount !== undefined && artifactCount > 0 && (
            <span className="text-xs px-2 py-1 rounded-full bg-background/60 text-muted-foreground">
              {artifactCount} {t('generate.artifacts')}
            </span>
          )}
        </div>
      </div>
      {summary && status === 'complete' && (
        <p className="mt-3 text-sm text-muted-foreground pl-11">{summary}</p>
      )}
      {error && status === 'error' && (
        <div className="mt-3 p-3 rounded-xl bg-status-error/10 border border-status-error/20" role="alert">
          <div className="flex items-start justify-between gap-3">
            <div className="flex-1">
              <div className="flex items-center gap-2 text-status-error font-medium text-sm mb-1">
                <svg className="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span>{t('generate.failed')}</span>
              </div>
              <p className="text-sm text-status-error/80 break-words">{error}</p>
            </div>
            {onRetry && (
              <button
                type="button"
                onClick={onRetry}
                disabled={isRetrying}
                className="px-3 py-1.5 text-xs font-medium text-background bg-status-error hover:bg-status-error/90 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap flex-shrink-0"
                aria-label={t('common.retry')}
              >
                {isRetrying ? t('common.loading') : t('common.retry')}
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
