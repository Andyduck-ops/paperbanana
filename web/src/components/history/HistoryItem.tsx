import { useLanguage, type HistorySession } from '../../hooks';

export interface HistoryItemProps {
  session: HistorySession;
  isSelected?: boolean;
  onClick?: () => void;
}

function generateSemanticSummary(
  session: HistorySession,
  fallbackTitle: string,
  maxLength = 30
): string {
  const seed = session.summary || session.prompt;
  if (!seed || seed.trim().length === 0) {
    return fallbackTitle;
  }

  const trimmed = seed.trim();
  const firstSentence = trimmed.split(/[.!?。！？]/)[0].trim();
  if (firstSentence.length <= maxLength) {
    return firstSentence;
  }

  return `${firstSentence.slice(0, maxLength - 1)}...`;
}

function detectMode(session: HistorySession): 'generate' | 'refine' | 'batch' | 'workspace' {
  if (session.mode) {
    return session.mode;
  }

  const prompt = `${session.summary || ''} ${session.prompt || ''}`.toLowerCase();

  if (prompt.includes('refine') || prompt.includes('精修')) return 'refine';
  if (prompt.includes('batch') || prompt.includes('批量') || prompt.includes('variants')) return 'batch';
  if (prompt.length > 0) return 'generate';
  return 'workspace';
}

function getModeLabel(mode: ReturnType<typeof detectMode>, language: string) {
  if (language === 'zh') {
    return {
      generate: '生成',
      refine: '精修',
      batch: '批量',
      workspace: '工作流',
    }[mode];
  }

  return {
    generate: 'Generate',
    refine: 'Refine',
    batch: 'Batch',
    workspace: 'Workspace',
  }[mode];
}

function getStatusLabel(status: string, language: string) {
  if (language === 'zh') {
    return {
      completed: '已完成',
      failed: '失败',
      running: '运行中',
      pending: '等待中',
    }[status] || status;
  }

  return {
    completed: 'Completed',
    failed: 'Failed',
    running: 'Running',
    pending: 'Pending',
  }[status] || status;
}

function statusClass(status: string) {
  switch (status) {
    case 'completed':
      return 'history-item__status history-item__status--completed';
    case 'failed':
      return 'history-item__status history-item__status--failed';
    case 'running':
      return 'history-item__status history-item__status--running';
    default:
      return 'history-item__status';
  }
}

export function HistoryItem({ session, isSelected, onClick }: HistoryItemProps) {
  const { language } = useLanguage();

  const date = new Date(session.createdAt);
  const isValidDate = !Number.isNaN(date.getTime());
  const timeLabel = isValidDate
    ? date.toLocaleDateString(language === 'zh' ? 'zh-CN' : 'en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    : session.createdAt;

  const mode = detectMode(session);
  const title = generateSemanticSummary(
    session,
    language === 'zh' ? '最近工作任务' : 'Recent workspace task'
  );
  const status = getStatusLabel(session.status, language);
  const sourceLabel = session.source === 'local'
    ? language === 'zh' ? '本地工作记录' : 'Local workspace record'
    : language === 'zh' ? '数据库记录' : 'Server record';

  return (
    <button
      type="button"
      onClick={onClick}
      className={`history-item ${isSelected ? 'history-item--selected' : ''}`}
    >
      <div className="history-item__header">
        <div className="history-item__stamp" aria-hidden="true">
          {mode === 'refine' ? '修' : mode === 'batch' ? '批' : '生'}
        </div>
        <div className="history-item__copy">
          <p className="history-item__title">{title}</p>
          <p className="history-item__meta">
            <span>{timeLabel}</span>
            <span>·</span>
            <span>{getModeLabel(mode, language)}</span>
            <span>·</span>
            <span>{sourceLabel}</span>
          </p>
        </div>
        <span className={statusClass(session.status)}>{status}</span>
      </div>

      {session.prompt && (
        <p className="history-item__prompt">
          {session.prompt}
        </p>
      )}
    </button>
  );
}
