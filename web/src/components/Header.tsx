import { useTheme, useLanguage } from '../hooks';

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
  const { theme, setTheme, themes } = useTheme();
  const { language, setLanguage, languages, t } = useLanguage();

  return (
    <div className="px-4 py-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      {/* Logo and title */}
      <div className="flex items-center gap-3">
        <a href="/" className="hover:opacity-80 transition-opacity">
          <h1 className="text-xl font-heading text-primary">
            {t('app.name')}
          </h1>
        </a>
        <span className="hidden sm:inline text-sm text-muted-foreground">
          {t('app.tagline')}
        </span>
      </div>

      {/* Controls */}
      <div className="flex flex-wrap items-center gap-2 sm:gap-4">
        {/* Projects link */}
        <a
          href="/projects"
          className="flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg bg-muted text-foreground hover:bg-muted/80 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <span className="hidden sm:inline">{t('projects.title') || 'Projects'}</span>
        </a>

        {/* History button */}
        {onHistoryClick && (
          <button
            onClick={onHistoryClick}
            className={`
              flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg transition-colors
              ${isHistoryOpen
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-foreground hover:bg-muted/80'
              }
            `}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="hidden sm:inline">{t('history.title') || 'History'}</span>
            {historyCount !== undefined && historyCount > 0 && (
              <span className="ml-1 px-1.5 py-0.5 text-xs bg-primary-foreground text-primary rounded-full">
                {historyCount}
              </span>
            )}
          </button>
        )}

        {/* Settings button */}
        {onSettingsClick && (
          <button
            onClick={onSettingsClick}
            className={`
              flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg transition-colors
              ${isSettingsOpen
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-foreground hover:bg-muted/80'
              }
            `}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <span className="hidden sm:inline">{t('settings.title') || 'Settings'}</span>
          </button>
        )}

        {/* Theme selector - Visual theme blocks */}
        <div className="theme-selector" role="radiogroup" aria-label={t('theme.title')}>
          {themes.map((themeItem) => (
            <button
              key={themeItem.id}
              onClick={() => setTheme(themeItem.id)}
              className={`
                theme-swatch theme-swatch--tone-${themeItem.id}
                ${theme === themeItem.id ? 'theme-swatch--selected' : ''}
              `}
              role="radio"
              aria-checked={theme === themeItem.id}
              aria-label={t('accessibility.switchTheme', { theme: themeItem.name })}
              title={`${themeItem.name}: ${t(`theme.options.${themeItem.id}.description`)}`}
            >
              <span className="theme-swatch__surface" aria-hidden="true">
                <span className="theme-swatch__shape theme-swatch__shape--primary" />
                <span className="theme-swatch__shape theme-swatch__shape--secondary" />
              </span>
            </button>
          ))}
        </div>

        {/* Language selector */}
        <div className="flex items-center gap-1">
          <span className="text-xs text-muted-foreground hidden sm:inline">
            {t('language.title')}:
          </span>
          <div className="flex gap-1">
            {languages.map((lang) => (
              <button
                key={lang.code}
                onClick={() => setLanguage(lang.code as 'zh' | 'en')}
                className={`
                  px-2 py-1 text-xs rounded transition-colors
                  ${language === lang.code
                    ? 'bg-primary text-background'
                    : 'bg-muted text-foreground hover:bg-muted/80'
                  }
                `}
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
