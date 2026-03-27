import { useLanguage } from '../../hooks';

interface HistoryTriggerProps {
  onClick: () => void;
  isActive?: boolean;
  count?: number;
  hasUnread?: boolean;
  compact?: boolean;
}

export function HistoryTrigger({
  onClick,
  isActive,
  count = 0,
  hasUnread = false,
  compact = false,
}: HistoryTriggerProps) {
  const { t } = useLanguage();

  return (
    <button
      type="button"
      onClick={onClick}
      className={`history-trigger ${compact ? 'history-trigger--compact' : ''} ${isActive ? 'history-trigger--active' : ''}`}
      aria-label={t('history.title')}
      aria-expanded={isActive}
    >
      <div className="history-trigger__icon" aria-hidden="true">
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

      <div className="history-trigger__content">
        <span className="history-trigger__label">{t('history.title')}</span>
        {!compact && <span className="history-trigger__hint">{t('history.subtitle')}</span>}
      </div>

      <div className="history-trigger__meta">
        {hasUnread ? (
          <span className="history-trigger__badge history-trigger__badge--pulse" />
        ) : (
          <span className="history-trigger__count">{count > 99 ? '99+' : count}</span>
        )}
      </div>
    </button>
  );
}
