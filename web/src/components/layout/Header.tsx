import { useLanguage } from '../../hooks';

export interface HeaderProps {
  onSettingsClick?: () => void;
  onHistoryClick?: () => void;
  isHistoryOpen?: boolean;
  isSettingsOpen?: boolean;
  historyCount?: number;
}

export function Header({
  onSettingsClick,
  onHistoryClick,
  isHistoryOpen,
  isSettingsOpen,
  historyCount,
}: HeaderProps) {
  const { language, setLanguage, languages, t } = useLanguage();

  return (
    <div className="rounded-2xl border border-border bg-card px-4 py-3 shadow-[var(--theme-surface-shadow)] sm:px-5">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <a href="/" className="flex items-center gap-3 hover:opacity-80 transition-opacity flex-shrink-0">
          <img src="/images/logo.png" alt="" className="h-10 w-10 rounded-2xl object-cover ring-1 ring-border" />
          <div className="min-w-0">
            <div className="text-[0.72rem] font-semibold uppercase tracking-[0.22em] text-muted-foreground">
              {t('workspace.kicker')}
            </div>
            <span className="block text-xl font-heading text-primary whitespace-nowrap">{t('app.name')}</span>
          </div>
        </a>

        <div className="flex flex-wrap items-center gap-2 lg:justify-end">
          <a
            href="/projects"
            className="flex min-h-10 items-center gap-2 rounded-xl border border-border bg-muted px-3 py-2 text-sm text-foreground hover:bg-muted/70 transition-colors"
            title={t('projects.title') || 'Projects'}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            <span>{t('projects.title') || 'Projects'}</span>
          </a>

          {onHistoryClick && (
            <button
              onClick={onHistoryClick}
              className={`flex min-h-10 items-center gap-2 rounded-xl border px-3 py-2 text-sm transition-colors ${
                isHistoryOpen ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-muted text-foreground hover:bg-muted/70'
              }`}
              title={t('history.title') || 'History'}
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>{t('history.title') || 'History'}</span>
              {historyCount !== undefined && historyCount > 0 && (
                <span className="ml-0.5 rounded-full bg-primary-foreground px-1.5 py-0.5 text-xs text-primary">
                  {historyCount}
                </span>
              )}
            </button>
          )}

          {onSettingsClick && (
            <button
              onClick={onSettingsClick}
              className={`flex min-h-10 items-center gap-2 rounded-xl border px-3 py-2 text-sm transition-colors ${
                isSettingsOpen ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-muted text-foreground hover:bg-muted/70'
              }`}
              title={t('settings.title') || 'Settings'}
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              <span>{t('settings.title') || 'Settings'}</span>
            </button>
          )}

          <div className="flex items-center gap-1 rounded-xl border border-border bg-background p-1">
            {languages.map((lang) => (
              <button
                key={lang.code}
                onClick={() => setLanguage(lang.code as 'zh' | 'en')}
                className={`rounded-lg px-2.5 py-1.5 text-xs font-medium transition-colors ${
                  language === lang.code ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'
                }`}
                aria-label={t('accessibility.switchLanguage', { language: lang.nativeName })}
                aria-pressed={language === lang.code}
              >
                {lang.code.toUpperCase()}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
