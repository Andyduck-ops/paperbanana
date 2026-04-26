import { useState, memo } from 'react';
import { useLanguage } from '../../hooks';
import type { Artifact } from '../ArtifactPreview';

export interface Candidate {
  id: string;
  index: number;
  status: 'pending' | 'running' | 'completed' | 'failed';
  artifacts?: Artifact[];
  error?: string;
  selected?: boolean;
  metadata?: {
    model?: string;
    seed?: number;
    duration?: number;
  };
}

export interface CandidateGridProps {
  candidates: Candidate[];
  selectedId?: string | null;
  selectedIds?: string[];
  onSelect?: (candidateId: string) => void;
  onMultiSelect?: (selectedIds: string[]) => void;
  multiSelect?: boolean;
  maxSelection?: number;
  onRefine?: (candidateId: string) => void;
  onDelete?: (candidateId: string) => void;
  onExport?: (artifact: Artifact) => void;
  viewMode?: 'grid' | 'list';
}

/**
 * Candidate Grid Component
 *
 * Displays multiple generation candidates for comparison.
 * Supports:
 * - Grid/List view toggle
 * - Quick candidate selection
 * - Multi-select mode for comparison (max 4 candidates)
 * - Status indicators
 * - Quick actions per candidate
 *
 * Design principles:
 * - Quick scanning layout
 * - Clear visual hierarchy
 * - Status at a glance
 * - Easy selection and branching
 */
export function CandidateGrid({
  candidates,
  selectedId,
  selectedIds = [],
  onSelect,
  onMultiSelect,
  multiSelect = false,
  maxSelection = 4,
  onRefine,
  onDelete,
  onExport,
  viewMode = 'grid',
}: CandidateGridProps) {
  const { t } = useLanguage();
  const [internalSelectedIds, setInternalSelectedIds] = useState<string[]>([]);
  const [showComparison, setShowComparison] = useState(false);

  // Use controlled or uncontrolled selection
  const currentSelectedIds = onMultiSelect ? selectedIds : internalSelectedIds;

  const handleToggleSelection = (candidateId: string) => {
    if (!multiSelect) {
      onSelect?.(candidateId);
      return;
    }

    const isSelected = currentSelectedIds.includes(candidateId);
    let newSelection: string[];

    if (isSelected) {
      newSelection = currentSelectedIds.filter(id => id !== candidateId);
    } else {
      if (currentSelectedIds.length >= maxSelection) {
        return; // Max selection reached
      }
      newSelection = [...currentSelectedIds, candidateId];
    }

    if (onMultiSelect) {
      onMultiSelect(newSelection);
    } else {
      setInternalSelectedIds(newSelection);
    }
  };

  const handleCompare = () => {
    if (currentSelectedIds.length >= 2) {
      setShowComparison(true);
    }
  };

  const handleCloseComparison = () => {
    setShowComparison(false);
  };

  const getStatusIcon = (status: Candidate['status']) => {
    switch (status) {
      case 'completed':
        return (
          <svg className="w-4 h-4 text-status-success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        );
      case 'running':
        return (
          <div className="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />
        );
      case 'failed':
        return (
          <svg className="w-4 h-4 text-status-error" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        );
      default:
        return (
          <div className="w-4 h-4 rounded-full border-2 border-muted-foreground/30" />
        );
    }
  };

  const getStatusLabel = (status: Candidate['status']) => {
    switch (status) {
      case 'completed':
        return t('generate.statusComplete');
      case 'running':
        return t('generate.statusRunning');
      case 'failed':
        return t('generate.statusError');
      default:
        return t('generate.statusPending');
    }
  };

  const selectedCandidates = candidates.filter(c => currentSelectedIds.includes(c.id));

  return (
    <div className="candidate-grid">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <h3 className="text-lg font-medium text-foreground">
            {t('generate.batchComplete')}
          </h3>
          <span className="text-sm text-muted-foreground">
            ({candidates.filter(c => c.status === 'completed').length} {t('generate.successful')})
          </span>
        </div>

        <div className="flex items-center gap-2">
          {/* Multi-select compare button */}
          {multiSelect && currentSelectedIds.length >= 2 && (
            <button
              type="button"
              onClick={handleCompare}
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:opacity-90 transition-opacity"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              {t('generate.compare') || 'Compare'} ({currentSelectedIds.length})
            </button>
          )}

          {/* View mode toggle */}
          <div className="flex items-center gap-1 p-1 rounded-lg bg-muted/50">
            <button
              type="button"
              onClick={() => {}}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === 'grid' ? 'bg-background shadow-sm' : 'hover:bg-muted'
              }`}
              aria-label="Grid view"
            >
              <svg className="w-4 h-4 text-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
            </button>
            <button
              type="button"
              onClick={() => {}}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === 'list' ? 'bg-background shadow-sm' : 'hover:bg-muted'
              }`}
              aria-label="List view"
            >
              <svg className="w-4 h-4 text-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {/* Selection hint */}
      {multiSelect && (
        <div className="mb-4 text-sm text-muted-foreground">
          {currentSelectedIds.length >= maxSelection ? (
            <span className="text-amber-500">{t('generate.maxSelectionReached') || `Maximum ${maxSelection} candidates can be selected`}</span>
          ) : (
            <span>{t('generate.selectToCompare') || `Select 2-${maxSelection} candidates to compare`}</span>
          )}
        </div>
      )}

      {/* Candidates */}
      <div className={`gap-4 ${
        viewMode === 'grid'
          ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'
          : 'flex flex-col'
      }`}>
        {candidates.map((candidate) => (
          <CandidateCard
            key={candidate.id}
            candidate={candidate}
            isSelected={multiSelect ? currentSelectedIds.includes(candidate.id) : selectedId === candidate.id}
            isMultiSelectMode={multiSelect}
            onSelect={() => handleToggleSelection(candidate.id)}
            onRefine={() => onRefine?.(candidate.id)}
            onDelete={() => onDelete?.(candidate.id)}
            onExport={onExport}
            viewMode={viewMode}
            statusIcon={getStatusIcon(candidate.status)}
            statusLabel={getStatusLabel(candidate.status)}
          />
        ))}
      </div>

      {/* Comparison Modal */}
      {showComparison && (
        <ComparisonModal
          candidates={selectedCandidates}
          onClose={handleCloseComparison}
        />
      )}
    </div>
  );
}

interface CandidateCardProps {
  candidate: Candidate;
  isSelected: boolean;
  isMultiSelectMode: boolean;
  onSelect: () => void;
  onRefine: () => void;
  onDelete: () => void;
  onExport?: (artifact: Artifact) => void;
  viewMode: 'grid' | 'list';
  statusIcon: React.ReactNode;
  statusLabel: string;
}

function CandidateCard({
  candidate,
  isSelected,
  isMultiSelectMode,
  onSelect,
  onRefine,
  onDelete,
  onExport,
  viewMode,
  statusIcon,
  statusLabel,
}: CandidateCardProps) {
  const { t } = useLanguage();
  const hasArtifacts = candidate.artifacts && candidate.artifacts.length > 0;
  const firstArtifact = hasArtifacts ? candidate.artifacts![0] : null;

  if (viewMode === 'list') {
    return (
      <div
        role="button"
        tabIndex={0}
        onClick={onSelect}
        onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(); } }}
        aria-pressed={isSelected}
        className={`
          candidate-card candidate-card--list
          flex items-center gap-4 p-4 rounded-xl border cursor-pointer
          transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
          ${isSelected
            ? 'border-primary bg-primary/5 shadow-sm'
            : 'border-border/50 bg-card/50 hover:border-primary/30 hover:bg-primary/5'
          }
        `}
      >
        {/* Checkbox for multi-select */}
        {isMultiSelectMode && (
          <button
            type="button"
            aria-label={isSelected ? 'Deselect' : 'Select'}
            className={`w-5 h-5 rounded border-2 flex items-center justify-center flex-shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 ${
              isSelected
                ? 'bg-primary border-primary'
                : 'border-muted-foreground/30 bg-background'
            }`}
            onClick={(e) => e.stopPropagation()}
          >
            {isSelected && (
              <svg className="w-3 h-3 text-primary-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
              </svg>
            )}
          </button>
        )}

        {/* Thumbnail or status */}
        <div className="w-16 h-16 rounded-lg bg-muted flex items-center justify-center flex-shrink-0 overflow-hidden"
        >
          {firstArtifact?.data ? (
            <img
              src={firstArtifact.data}
              alt={`Candidate ${candidate.index + 1}`}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="text-muted-foreground">{statusIcon}</div>
          )}
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-foreground">
              {t('generate.candidate', { number: candidate.index + 1 })}
            </span>
            <span className={`
              text-xs px-2 py-0.5 rounded-full
              ${candidate.status === 'completed' ? 'bg-status-success/10 text-status-success' :
                candidate.status === 'running' ? 'bg-status-running/10 text-status-running' :
                candidate.status === 'failed' ? 'bg-status-error/10 text-status-error' :
                'bg-muted text-muted-foreground'}
            `}>
              {statusLabel}
            </span>
          </div>
          {candidate.error && (
            <p className="text-xs text-status-error mt-1 truncate">{candidate.error}</p>
          )}
          {candidate.metadata && (
            <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground"
            >
              {candidate.metadata.model && (
                <span>{candidate.metadata.model}</span>
              )}
              {candidate.metadata.duration && (
                <span>{Math.round(candidate.metadata.duration / 1000)}s</span>
              )}
            </div>
          )}
        </div>

        {/* Actions */}
        {candidate.status === 'completed' && hasArtifacts && (
          <div className="flex items-center gap-1"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              onClick={onRefine}
              className="p-2 rounded-lg hover:bg-muted transition-colors"
              title={t('refine.title')}
            >
              <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
              </svg>
            </button>
            {firstArtifact && (
              <button
                onClick={() => onExport?.(firstArtifact)}
                className="p-2 rounded-lg hover:bg-muted transition-colors"
                title={t('export.title')}
              >
                <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
              </button>
            )}
            <button
              onClick={onDelete}
              className="p-2 rounded-lg hover:bg-muted transition-colors"
              title={t('common.delete')}
            >
              <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        )}
      </div>
    );
  }

  // Grid view
  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onSelect}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(); } }}
      aria-pressed={isSelected}
      className={`
        candidate-card candidate-card--grid
        rounded-xl border overflow-hidden cursor-pointer
        transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
        ${isSelected
          ? 'border-primary bg-primary/5 shadow-sm ring-2 ring-primary/20'
          : 'border-border/50 bg-card/50 hover:border-primary/30 hover:bg-primary/5'
        }
      `}
    >
      {/* Image or placeholder */}
      <div className="aspect-video bg-muted flex items-center justify-center relative"
      >
        {firstArtifact?.data ? (
          <img
            src={firstArtifact.data}
            alt={`Candidate ${candidate.index + 1}`}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="flex flex-col items-center gap-2 text-muted-foreground">
            {statusIcon}
            <span className="text-xs">{statusLabel}</span>
          </div>
        )}

        {/* Checkbox for multi-select */}
        {isMultiSelectMode && (
          <button
            type="button"
            aria-label={isSelected ? 'Deselect' : 'Select'}
            className={`absolute top-2 right-2 w-6 h-6 rounded-md flex items-center justify-center transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 ${
              isSelected
                ? 'bg-primary text-primary-foreground'
                : 'bg-background/80 text-muted-foreground border border-border/50'
            }`}
            onClick={(e) => e.stopPropagation()}
          >
            {isSelected ? (
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
              </svg>
            ) : (
              <div className="w-3 h-3" />
            )}
          </button>
        )}

        {/* Single selection indicator (when not in multi-select mode) */}
        {!isMultiSelectMode && isSelected && (
          <div className="absolute top-2 right-2 w-6 h-6 rounded-full bg-primary text-primary-foreground flex items-center justify-center"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
        )}

        {/* Status badge */}
        <div className={`
          absolute top-2 left-2 px-2 py-0.5 rounded-full text-xs font-medium
          ${candidate.status === 'completed' ? 'bg-status-success/90 text-white' :
            candidate.status === 'running' ? 'bg-status-running/90 text-white' :
            candidate.status === 'failed' ? 'bg-status-error/90 text-white' :
            'bg-muted-foreground/50 text-white'}
        `}
        >
          #{candidate.index + 1}
        </div>
      </div>

      {/* Info and actions */}
      <div className="p-3"
      >
        {candidate.error && (
          <p className="text-xs text-status-error mb-2 line-clamp-2">{candidate.error}</p>
        )}

        {candidate.status === 'completed' && hasArtifacts && (
          <div
            className="flex items-center justify-between"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center gap-1"
            >
              <button
                onClick={onRefine}
                className="p-1.5 rounded-lg hover:bg-muted transition-colors"
                title={t('refine.title')}
              >
                <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
              </button>
              {firstArtifact && (
                <button
                  onClick={() => onExport?.(firstArtifact)}
                  className="p-1.5 rounded-lg hover:bg-muted transition-colors"
                  title={t('export.title')}
                >
                  <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                  </svg>
                </button>
              )}
            </div>

            <button
              onClick={onDelete}
              className="p-1.5 rounded-lg hover:bg-muted transition-colors"
              title={t('common.delete')}
            >
              <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

// ============================================
// Comparison Modal Component
// ============================================

interface ComparisonModalProps {
  candidates: Candidate[];
  onClose: () => void;
}

function ComparisonModal({ candidates, onClose }: ComparisonModalProps) {
  const { t } = useLanguage();
  const [selectedIndex, setSelectedIndex] = useState(0);

  const completedCandidates = candidates.filter(c => c.status === 'completed' && c.artifacts && c.artifacts.length > 0);

  if (completedCandidates.length === 0) {
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
        <div className="bg-background rounded-2xl shadow-2xl max-w-md w-full p-6">
          <div className="text-center">
            <p className="text-muted-foreground">{t('generate.noCompletedCandidates') || 'No completed candidates to compare'}</p>
            <button
              onClick={onClose}
              className="mt-4 px-4 py-2 bg-primary text-primary-foreground rounded-lg"
            >
              {t('common.close')}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
      <div className="bg-background rounded-2xl shadow-2xl max-w-7xl w-full max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h3 className="text-lg font-semibold">
            {t('generate.compareCandidates') || 'Compare Candidates'}
          </h3>
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-muted transition-colors"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Comparison Grid */}
        <div className="flex-1 overflow-auto p-4">
          {completedCandidates.length === 2 ? (
            // Side-by-side comparison for 2 candidates
            <div className="grid grid-cols-2 gap-4 h-full">
              {completedCandidates.map((candidate) => (
                <ComparisonCard key={candidate.id} candidate={candidate} />
              ))}
            </div>
          ) : (
            // Gallery view for more candidates
            <div className="space-y-4">
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {completedCandidates.map((candidate, index) => (
                  <button
                    key={candidate.id}
                    onClick={() => setSelectedIndex(index)}
                    className={`relative rounded-xl border overflow-hidden transition-all ${
                      selectedIndex === index
                        ? 'ring-2 ring-primary border-primary'
                        : 'border-border hover:border-primary/50'
                    }`}
                  >
                    <div className="aspect-video bg-muted">
                      {candidate.artifacts?.[0]?.data && (
                        <img
                          src={candidate.artifacts[0].data}
                          alt={`Candidate ${candidate.index + 1}`}
                          className="w-full h-full object-cover"
                        />
                      )}
                    </div>
                    <div className="p-2 text-center text-sm font-medium">
                      #{candidate.index + 1}
                    </div>
                  </button>
                ))}
              </div>
              {/* Large preview of selected */}
              {completedCandidates[selectedIndex] && (
                <div className="rounded-xl border border-border overflow-hidden">
                  <div className="aspect-video bg-muted relative">
                    {completedCandidates[selectedIndex].artifacts?.[0]?.data && (
                      <img
                        src={completedCandidates[selectedIndex].artifacts![0].data}
                        alt={`Candidate ${completedCandidates[selectedIndex].index + 1}`}
                        className="w-full h-full object-contain"
                      />
                    )}
                  </div>
                  <div className="p-4 bg-muted/30">
                    <p className="font-medium">{t('generate.candidate', { number: completedCandidates[selectedIndex].index + 1 })}</p>
                    {completedCandidates[selectedIndex].metadata && (
                      <div className="flex gap-4 mt-2 text-sm text-muted-foreground">
                        {completedCandidates[selectedIndex].metadata!.model && (
                          <span>Model: {completedCandidates[selectedIndex].metadata!.model}</span>
                        )}
                        {completedCandidates[selectedIndex].metadata!.duration && (
                          <span>Duration: {Math.round(completedCandidates[selectedIndex].metadata!.duration! / 1000)}s</span>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-border flex justify-between items-center">
          <span className="text-sm text-muted-foreground">
            {completedCandidates.length} {t('generate.candidatesSelected') || 'candidates selected'}
          </span>
          <button
            onClick={onClose}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
          >
            {t('common.close')}
          </button>
        </div>
      </div>
    </div>
  );
}

function ComparisonCard({ candidate }: { candidate: Candidate }) {
  const { t } = useLanguage();
  const artifact = candidate.artifacts?.[0];

  return (
    <div className="flex flex-col h-full rounded-xl border border-border overflow-hidden bg-card">
      <div className="flex-1 bg-muted relative min-h-0">
        {artifact?.data ? (
          <img
            src={artifact.data}
            alt={`Candidate ${candidate.index + 1}`}
            className="w-full h-full object-contain"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-muted-foreground">
            {t('generate.noImage') || 'No image'}
          </div>
        )}
      </div>
      <div className="p-4 bg-muted/30 border-t border-border">
        <p className="font-medium text-lg">{t('generate.candidate', { number: candidate.index + 1 })}</p>
        {candidate.metadata && (
          <div className="flex flex-col gap-1 mt-2 text-sm text-muted-foreground">
            {candidate.metadata.model && (
              <span>Model: {candidate.metadata.model}</span>
            )}
            {candidate.metadata.duration && (
              <span>Duration: {Math.round(candidate.metadata.duration / 1000)}s</span>
            )}
            {candidate.metadata.seed && (
              <span>Seed: {candidate.metadata.seed}</span>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

const MemoizedCandidateGrid = memo(CandidateGrid);
MemoizedCandidateGrid.displayName = 'CandidateGrid';
export default MemoizedCandidateGrid;
