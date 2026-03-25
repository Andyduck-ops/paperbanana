import { useLanguage } from '../hooks';

export interface Artifact {
  kind: string;
  mimeType: string;
  summary: string;
  data?: string;
  assetId?: string;
  projectId?: string;
}

export interface ArtifactPreviewProps {
  artifact: Artifact;
  onExport?: (artifact: Artifact) => void;
  onCopy?: (artifact: Artifact) => void;
}

/**
 * getArtifactImageUrl constructs the image URL for an artifact.
 * Priority: base64 data > asset API with projectId > asset API without projectId
 */
export function getArtifactImageUrl(artifact: Artifact): string | null {
  if (artifact.data) {
    return `data:${artifact.mimeType};base64,${artifact.data}`;
  }
  if (artifact.assetId) {
    if (artifact.projectId) {
      return `/api/v1/assets/${artifact.projectId}/${artifact.assetId}/download`;
    }
    // Fallback: use assetId as-is (deprecated, won't work without projectId)
    return `/api/v1/assets/${artifact.assetId}`;
  }
  return null;
}

export function ArtifactPreview({ artifact, onExport, onCopy }: ArtifactPreviewProps) {
  const { t } = useLanguage();

  const isImage = artifact.mimeType.startsWith('image/');
  const imageUrl = getArtifactImageUrl(artifact);

  return (
    <div className="border border-border rounded-lg overflow-hidden bg-background">
      {isImage && imageUrl && (
        <div className="aspect-video bg-muted flex items-center justify-center">
          <img
            src={imageUrl}
            alt={artifact.summary}
            className="max-w-full max-h-full object-contain"
          />
        </div>
      )}
      <div className="p-4">
        <div className="flex items-center justify-between mb-2">
          <span className="text-xs px-2 py-1 rounded bg-muted text-muted-foreground">
            {artifact.kind}
          </span>
          <div className="flex gap-2">
            {onCopy && artifact.data && (
              <button
                onClick={() => onCopy(artifact)}
                className="text-xs text-muted-foreground hover:text-primary"
              >
                {t('export.copy') || 'Copy'}
              </button>
            )}
            {onExport && (
              <button
                onClick={() => onExport(artifact)}
                className="text-xs text-primary hover:underline"
              >
                {t('export.title')}
              </button>
            )}
          </div>
        </div>
        <p className="text-sm text-foreground">{artifact.summary}</p>
      </div>
    </div>
  );
}
