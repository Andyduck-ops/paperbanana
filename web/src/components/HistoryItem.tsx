import { useLanguage } from '../hooks';
import type { HistorySession } from '../hooks/useHistory';

export interface HistoryItemProps {
  session: HistorySession;
  isSelected?: boolean;
  onClick?: () => void;
}

export function HistoryItem({ session, isSelected, onClick }: HistoryItemProps) {
  const { t, language } = useLanguage();

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString(language === 'zh' ? 'zh-CN' : 'en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <button
      onClick={onClick}
      className={`
        w-full p-3 rounded-lg text-left transition-colors
        ${isSelected
          ? 'bg-primary/20 border-primary'
          : 'bg-background hover:bg-muted border-border'
        }
        border
      `}
    >
      <div className="flex gap-3">
        {session.thumbnailUrl && (
          <div className="w-12 h-12 rounded bg-muted shrink-0 overflow-hidden">
            <img
              src={session.thumbnailUrl}
              alt=""
              className="w-full h-full object-cover"
            />
          </div>
        )}
        <div className="flex-1 min-w-0">
          <p className="text-sm text-foreground truncate">
            {session.prompt || t('history.untitled')}
          </p>
          <p className="text-xs text-muted-foreground mt-1">
            {formatDate(session.createdAt)}
          </p>
        </div>
      </div>
    </button>
  );
}
