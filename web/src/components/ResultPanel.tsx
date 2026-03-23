import { useLanguage } from '../hooks';
import { ArtifactPreview, type Artifact } from './ArtifactPreview';

export interface ResultPanelProps {
  sessionId: string;
  artifacts: Artifact[];
  onExport?: (artifact: Artifact) => void;
  onCopy?: (artifact: Artifact) => void;
  onNewGeneration?: () => void;
}

export function ResultPanel({
  sessionId,
  artifacts,
  onExport,
  onCopy,
  onNewGeneration,
}: ResultPanelProps) {
  const { t } = useLanguage();

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-heading text-foreground">
          {t('generate.result')}
        </h3>
        <span className="text-xs text-muted-foreground font-mono">
          {sessionId}
        </span>
      </div>

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

      {onNewGeneration && (
        <button
          onClick={onNewGeneration}
          className="w-full px-4 py-2 rounded-lg border border-border text-foreground hover:bg-muted transition-colors"
        >
          {t('generate.new')}
        </button>
      )}
    </div>
  );
}
