import { useHistory, type HistorySession } from '../../hooks/useHistory';

export interface HistoryPopoverProps {
  onClose: () => void;
  onSelectSession: (sessionId: string) => void;
  projectId?: string;
}

export function HistoryPopover({ onClose, onSelectSession, projectId }: HistoryPopoverProps) {
  const { sessions, isLoading, error } = useHistory(projectId);

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <span className="text-green-500 text-xs">✓</span>;
      case 'failed':
        return <span className="text-red-500 text-xs">✗</span>;
      case 'running':
        return <span className="text-primary text-xs animate-pulse">●</span>;
      default:
        return <span className="text-muted-foreground text-xs">○</span>;
    }
  };

  const truncateTitle = (session: HistorySession, maxLength = 24) => {
    const title = session.prompt || `Session ${session.id.slice(0, 6)}`;
    return title.length > maxLength ? title.slice(0, maxLength) + '...' : title;
  };

  return (
    <div
      className="w-72 bg-card/95 backdrop-blur-lg rounded-xl border border-border/30 shadow-2xl overflow-hidden"
      onClick={(e) => e.stopPropagation()}
    >
      <div className="flex items-center justify-between px-4 py-3 border-b border-border/20">
        <h3 className="text-sm font-medium text-foreground">History</h3>
        <button
          onClick={onClose}
          className="w-6 h-6 rounded-md flex items-center justify-center hover:bg-muted/50 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div className="max-h-80 overflow-y-auto">
        {isLoading && (
          <div className="px-4 py-8 text-center text-sm text-muted-foreground">
            Loading...
          </div>
        )}

        {error && (
          <div className="px-4 py-4 text-center text-sm text-red-500">
            {error}
          </div>
        )}

        {!isLoading && !error && sessions.length === 0 && (
          <div className="px-4 py-8 text-center text-sm text-muted-foreground">
            No history yet
          </div>
        )}

        {!isLoading && !error && sessions.map((session) => (
          <button
            key={session.id}
            onClick={() => onSelectSession(session.id)}
            className="w-full px-4 py-3 flex items-start gap-3 hover:bg-muted/30 transition-colors text-left"
          >
            <div className="w-8 h-8 rounded-lg bg-muted/50 flex items-center justify-center shrink-0">
              <svg className="w-4 h-4 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-foreground truncate">
                {truncateTitle(session)}
              </p>
              <p className="text-xs text-muted-foreground">{formatTime(session.createdAt)}</p>
            </div>
            {getStatusIcon(session.status)}
          </button>
        ))}
      </div>

      {sessions.length > 0 && (
        <div className="px-4 py-3 border-t border-border/20">
          <button
            className="w-full text-sm text-primary hover:text-primary/80 transition-colors"
            onClick={() => {
              // TODO: Navigate to full history view
            }}
          >
            View All History →
          </button>
        </div>
      )}
    </div>
  );
}
