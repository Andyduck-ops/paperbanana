import { useState } from 'react';
import { useLanguage } from '../../hooks';
import type { Version } from '../../hooks/useVersions';

export interface VersionTimelineProps {
  versions: Version[];
  isLoading: boolean;
  onRestore: (versionId: string) => void;
  onPreview?: (version: Version) => void;
}

export function VersionTimeline({
  versions,
  isLoading,
  onRestore,
  onPreview,
}: VersionTimelineProps) {
  const { t } = useLanguage();
  const [expandedVersion, setExpandedVersion] = useState<string | null>(null);

  if (isLoading) {
    return (
      <div className="version-timeline-loading">
        <div className="w-6 h-6 border-2 border-primary/25 border-t-primary rounded-full animate-spin" />
        <span>{t('version.loading') || 'Loading versions...'}</span>
      </div>
    );
  }

  if (versions.length === 0) {
    return (
      <div className="version-timeline-empty">
        <svg className="w-12 h-12 text-muted-foreground/50" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p>{t('version.noVersions') || 'No versions yet'}</p>
      </div>
    );
  }

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString();
  };

  return (
    <div className="version-timeline">
      <div className="version-timeline-header">
        <h3 className="version-timeline-title">{t('version.title') || 'Version History'}</h3>
        <span className="version-timeline-count">{versions.length} {t('version.versions') || 'versions'}</span>
      </div>

      <div className="version-timeline-list">
        {versions.map((version, index) => (
          <div
            key={version.id}
            className={`version-timeline-item ${index === 0 ? 'version-timeline-item--latest' : ''}`}
          >
            <div className="version-timeline-marker">
              <div className="version-timeline-dot" />
              {index < versions.length - 1 && <div className="version-timeline-line" />}
            </div>

            <div className="version-timeline-content">
              <div
                className="version-timeline-card"
                onClick={() => setExpandedVersion(expandedVersion === version.id ? null : version.id)}
              >
                <div className="version-timeline-card-header">
                  <div className="version-timeline-info">
                    <span className="version-timeline-number">
                      {t('version.v') || 'v'}{version.version}
                      {index === 0 && (
                        <span className="version-timeline-badge">
                          {t('version.latest') || 'Latest'}
                        </span>
                      )}
                    </span>
                    <span className="version-timeline-date">{formatDate(version.created_at)}</span>
                  </div>
                  <div className="version-timeline-actions">
                    {onPreview && version.artifacts.length > 0 && (
                      <button
                        className="version-timeline-btn version-timeline-btn--secondary"
                        onClick={(e) => {
                          e.stopPropagation();
                          onPreview(version);
                        }}
                      >
                        {t('version.preview') || 'Preview'}
                      </button>
                    )}
                    <button
                      className="version-timeline-btn version-timeline-btn--primary"
                      onClick={(e) => {
                        e.stopPropagation();
                        onRestore(version.id);
                      }}
                    >
                      {t('version.restore') || 'Restore'}
                    </button>
                  </div>
                </div>

                {expandedVersion === version.id && version.artifacts.length > 0 && (
                  <div className="version-timeline-artifacts">
                    <p className="version-timeline-artifacts-title">
                      {t('version.artifacts') || 'Artifacts'}
                    </p>
                    <div className="version-timeline-artifacts-list">
                      {version.artifacts.map((artifact) => (
                        <div key={artifact.id} className="version-timeline-artifact">
                          <span className="version-timeline-artifact-kind">{artifact.kind}</span>
                          <span className="version-timeline-artifact-mime">{artifact.mime_type}</span>
                          {artifact.summary && (
                            <span className="version-timeline-artifact-summary">{artifact.summary}</span>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {version.artifacts.length > 0 && (
                  <div className="version-timeline-expand">
                    <svg
                      className={`w-4 h-4 transition-transform ${expandedVersion === version.id ? 'rotate-180' : ''}`}
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={2}
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                    </svg>
                  </div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
