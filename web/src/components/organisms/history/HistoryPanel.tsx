import { useEffect, useRef, useState, useMemo, memo } from 'react';
import { useLanguage, useHistory, useLocalWorkRecords } from '../../../hooks';
import { HistoryItem } from '../../history/HistoryItem';
import type { HistorySession } from '../../../hooks';

export interface HistoryPanelProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectSession: (sessionId: string) => void;
  selectedSessionId?: string;
  projectId?: string;
}

function HistoryPanelComponent({
  isOpen,
  onClose,
  onSelectSession,
  selectedSessionId,
  projectId,
}: HistoryPanelProps) {
  const { t } = useLanguage();
  const { sessions: serverSessions, isLoading, error, refresh } = useHistory(projectId);
  const { records: localRecords } = useLocalWorkRecords();
  const panelRef = useRef<HTMLDivElement>(null);
  
  // Merge server sessions with local work records as fallback
  const sessions = useMemo((): HistorySession[] => {
    // If we have server sessions, use those (they're the source of truth)
    if (serverSessions && serverSessions.length > 0) {
      return serverSessions;
    }
    // Otherwise, fall back to local work records converted to HistorySession format
    return localRecords.map((record): HistorySession => ({
      id: record.id,
      projectId: projectId || 'local',
      createdAt: record.createdAt,
      status: record.status,
      prompt: record.prompt,
      // Mark as local source for display purposes
      source: 'local' as const,
    }));
  }, [serverSessions, localRecords, projectId]);

  const [searchQuery, setSearchQuery] = useState('');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [showFilters, setShowFilters] = useState(false);

  const filteredSessions = useMemo(() => {
    return sessions.filter((session) => {
      if (searchQuery.trim()) {
        const query = searchQuery.toLowerCase();
        const matchesSearch =
          session.id.toLowerCase().includes(query) ||
          (session.prompt?.toLowerCase().includes(query) ?? false);
        if (!matchesSearch) return false;
      }

      if (dateFrom || dateTo) {
        const sessionDate = new Date(session.createdAt);
        if (dateFrom) {
          const fromDate = new Date(dateFrom);
          fromDate.setHours(0, 0, 0, 0);
          if (sessionDate < fromDate) return false;
        }
        if (dateTo) {
          const toDate = new Date(dateTo);
          toDate.setHours(23, 59, 59, 999);
          if (sessionDate > toDate) return false;
        }
      }

      return true;
    });
  }, [sessions, searchQuery, dateFrom, dateTo]);

  const clearFilters = () => {
    setSearchQuery('');
    setDateFrom('');
    setDateTo('');
  };

  const hasActiveFilters = searchQuery || dateFrom || dateTo;

  useEffect(() => {
    if (!isOpen) {
      const mainContent = document.querySelector('[data-main-workspace]');
      if (mainContent instanceof HTMLElement) {
        mainContent.focus();
      }
    }
  }, [isOpen]);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  if (!isOpen) {
    return null;
  }

  return (
    <>
      <div
        className={`
          fixed inset-0 bg-black/25 backdrop-blur-sm z-40
          transition-opacity duration-300 ease-out
          ${isOpen ? 'opacity-100' : 'opacity-0 pointer-events-none'}
        `}
        onClick={handleBackdropClick}
        aria-hidden={!isOpen}
      />

      <div
        ref={panelRef}
        className={`
          history-panel fixed left-0 top-0 bottom-0 z-50
          w-[28rem] max-w-[92vw]
          transform transition-transform duration-300 ease-out
          ${isOpen ? 'translate-x-0' : '-translate-x-full'}
        `}
        role="dialog"
        aria-modal="true"
        aria-label={t('history.title')}
      >
        <div className="history-panel__header">
          <div className="history-panel__header-main">
            <div className="history-panel__icon" aria-hidden="true">
              <svg
                className="w-5 h-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={1.5}
              >
                {/* Document/stack icon representing work records */}
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75 2.25 2.25 0 00-.1-.664m-5.8 0A2.251 2.251 0 0113.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V16.5a2.25 2.25 0 002.25 2.25h.75m-3 .75h9M3.75 6.75h16.5M3.75 12h16.5"
                />
              </svg>
            </div>
            <div>
              <h2 className="history-panel__title">{t('history.title')}</h2>
              <p className="history-panel__subtitle">{t('history.subtitle')}</p>
            </div>
          </div>

          <div className="history-panel__actions">
            <button
              type="button"
              onClick={refresh}
              className="history-panel__action"
              aria-label={t('common.retry')}
              title={t('common.retry')}
            >
              <svg
                className="w-4 h-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
            </button>
            <button
              type="button"
              onClick={onClose}
              className="history-panel__action"
              aria-label={t('common.close')}
            >
              <svg
                className="w-4 h-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
        </div>

        <div className="px-4 py-3 border-b border-border/50 bg-muted/30">
          <div className="relative mb-3">
            <svg
              className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder={t('history.search')}
              className="w-full pl-10 pr-9 py-2 text-sm bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            />
            {searchQuery && (
              <button
                type="button"
                onClick={() => setSearchQuery('')}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            )}
          </div>

          <div className="flex items-center justify-between">
            <button
              type="button"
              onClick={() => setShowFilters(!showFilters)}
              className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground px-2 py-1 rounded hover:bg-muted transition-colors"
            >
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
              </svg>
              {t('history.filter')}
              {hasActiveFilters && <span className="w-1.5 h-1.5 rounded-full bg-primary" />}
            </button>

            {hasActiveFilters && (
              <button
                type="button"
                onClick={clearFilters}
                className="text-xs text-muted-foreground hover:text-foreground"
              >
                {t('history.clearFilters')}
              </button>
            )}
          </div>

          {showFilters && (
            <div className="mt-3 pt-3 border-t border-border/50 grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs text-muted-foreground mb-1">{t('history.from')}</label>
                <input
                  type="date"
                  value={dateFrom}
                  onChange={(e) => setDateFrom(e.target.value)}
                  className="w-full px-2 py-1.5 text-xs bg-background border border-border rounded focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
              <div>
                <label className="block text-xs text-muted-foreground mb-1">{t('history.to')}</label>
                <input
                  type="date"
                  value={dateTo}
                  onChange={(e) => setDateTo(e.target.value)}
                  className="w-full px-2 py-1.5 text-xs bg-background border border-border rounded focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
            </div>
          )}
        </div>

        <div className="history-panel__body">
          {isLoading && (
            <div className="history-panel__state">
              <div className="w-8 h-8 border-2 border-primary/25 border-t-primary rounded-full animate-spin" />
              <p>{t('common.loading')}</p>
            </div>
          )}

          {error && !isLoading && (
            <div className="history-panel__error">
              <p>{error}</p>
              <button type="button" onClick={refresh}>
                {t('common.retry')}
              </button>
            </div>
          )}

          {!isLoading && !error && sessions.length === 0 && (
            <div className="history-panel__empty">
              <div className="history-panel__empty-mark" aria-hidden="true">记</div>
              <p className="history-panel__empty-title">{t('history.empty')}</p>
              <p className="history-panel__empty-copy">{t('history.emptyHint')}</p>
            </div>
          )}

          {!isLoading && !error && sessions.length > 0 && filteredSessions.length === 0 && (
            <div className="history-panel__empty">
              <div className="history-panel__empty-mark" aria-hidden="true">搜</div>
              <p className="history-panel__empty-title">{t('history.noResults')}</p>
              <p className="history-panel__empty-copy">Try adjusting your search or filters</p>
            </div>
          )}

          {!isLoading && !error && filteredSessions.length > 0 && (
            <div className="history-panel__list">
              {filteredSessions.map((session) => (
                <HistoryItem
                  key={session.id}
                  session={session}
                  isSelected={session.id === selectedSessionId}
                  onClick={() => {
                    onSelectSession(session.id);
                    onClose();
                  }}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  );
}

export const HistoryPanel = memo(HistoryPanelComponent);
HistoryPanel.displayName = 'HistoryPanel';
