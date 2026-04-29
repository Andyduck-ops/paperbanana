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
      return 'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-status-success/10 text-status-success';
    case 'failed':
      return 'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-status-error/10 text-status-error';
    case 'running':
      return 'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-status-running/10 text-status-running';
    default:
      return 'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground';
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
      className={`w-full text-left px-4 py-3 transition-colors hover:bg-muted/50 ${isSelected ? 'bg-primary/5 border-l-2 border-primary' : 'border-l-2 border-transparent'}`}
    >
      <div className="flex items-start gap-3">
        <div className={`w-8 h-8 rounded-lg flex items-center justify-center text-xs font-semibold shrink-0 ${mode === 'refine' ? 'bg-amber-100 text-amber-700' : mode === 'batch' ? 'bg-blue-100 text-blue-700' : 'bg-emerald-100 text-emerald-700'}`} aria-hidden="true">
          {mode === 'refine' ? '修' : mode === 'batch' ? '批' : '生'}
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-foreground truncate">{title}</p>
          <p className="text-xs text-muted-foreground mt-0.5 flex items-center gap-1.5 flex-wrap">
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
        <p className="mt-2 text-xs text-muted-foreground line-clamp-2 pl-11">
          {session.prompt}
        </p>
      )}
    </button>
  );
}
