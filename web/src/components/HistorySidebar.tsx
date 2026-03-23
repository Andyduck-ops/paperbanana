import { useLanguage, useHistory } from '../hooks';
import { HistoryItem } from './HistoryItem';

export interface HistorySidebarProps {
  projectId?: string;
  selectedSessionId?: string;
  onSelectSession?: (sessionId: string) => void;
}

export function HistorySidebar({
  projectId,
  selectedSessionId,
  onSelectSession,
}: HistorySidebarProps) {
  const { t } = useLanguage();
  const { sessions, isLoading, error } = useHistory(projectId);

  return (
    <div className="h-full flex flex-col">
      <h3 className="text-lg font-heading text-foreground mb-4">
        {t('history.title')}
      </h3>

      {isLoading && (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-muted-foreground">{t('common.loading')}</div>
        </div>
      )}

      {error && (
        <div className="p-4 rounded-lg bg-red-500/10 text-red-600">
          {error}
        </div>
      )}

      {!isLoading && sessions.length === 0 && (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-muted-foreground">{t('history.empty')}</div>
        </div>
      )}

      {!isLoading && sessions.length > 0 && (
        <div className="flex-1 overflow-y-auto space-y-2">
          {sessions.map((session) => (
            <HistoryItem
              key={session.id}
              session={session}
              isSelected={session.id === selectedSessionId}
              onClick={() => onSelectSession?.(session.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
