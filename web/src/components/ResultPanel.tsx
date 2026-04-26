import { memo } from 'react';
import { useLanguage } from '../hooks';
import { ArtifactPreview, type Artifact } from './ArtifactPreview';

export interface ResultPanelProps {
  sessionId: string;
  artifacts: Artifact[];
  onExport?: (artifact: Artifact) => void;
  onCopy?: (artifact: Artifact) => void;
  onNewGeneration?: () => void;
}

function ResultPanelComponent({
  sessionId,
  artifacts,
  onExport,
  onCopy,
  onNewGeneration,
}: ResultPanelProps) {
  const { t } = useLanguage();

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-status-success/15 flex items-center justify-center">
            <svg className="w-4 h-4 text-status-success" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h3 className="text-lg font-heading font-semibold text-foreground">
            {t('generate.result')}
          </h3>
        </div>
        <span className="text-xs text-muted-foreground font-mono px-2 py-1 rounded-md bg-muted/50">
          {sessionId.slice(0, 8)}
        </span>
      </div>

      {/* Artifacts */}
      <div className="grid gap-4">
        {artifacts.map((artifact, index) => (
          <ArtifactPreview
            key={`${artifact.kind}-${index}`}
            artifact={artifact}
            onExport={onExport}
            onCopy={onCopy}
          />
        ))}
      </div>

      {/* New Generation Button */}
      {onNewGeneration && (
        <button
          onClick={onNewGeneration}
          className="w-full px-4 py-3 rounded-xl border border-border/70 text-foreground hover:bg-muted/60 hover:border-primary/30 active:scale-[0.99] transition-all duration-200 flex items-center justify-center gap-2 text-sm font-medium"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          {t('generate.new')}
        </button>
      )}
    </div>
  );
}

export const ResultPanel = memo(ResultPanelComponent);
ResultPanel.displayName = 'ResultPanel';
