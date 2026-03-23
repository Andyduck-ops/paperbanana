import { useTheme, useLanguage } from '../hooks';

export function Header() {
  const { theme, setTheme, themes } = useTheme();
  const { language, setLanguage, languages, t } = useLanguage();

  return (
    <div className="px-4 py-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      {/* Logo and title */}
      <div className="flex items-center gap-3">
        <h1 className="text-xl font-heading text-primary">
          {t('app.name')}
        </h1>
        <span className="hidden sm:inline text-sm text-muted-foreground">
          {t('app.tagline')}
        </span>
      </div>

      {/* Controls */}
      <div className="flex flex-wrap items-center gap-2 sm:gap-4">
        {/* Theme selector */}
        <div className="flex items-center gap-1">
          <span className="text-xs text-muted-foreground hidden sm:inline">
            {t('theme.title')}:
          </span>
          <div className="flex gap-1">
            {themes.map((themeItem) => (
              <button
                key={themeItem.id}
                onClick={() => setTheme(themeItem.id)}
                className={`
                  px-2 py-1 text-xs rounded transition-colors
                  ${theme === themeItem.id
                    ? 'bg-primary text-background'
                    : 'bg-muted text-foreground hover:bg-muted/80'
                  }
                `}
                aria-label={`Switch to ${themeItem.name} theme`}
                aria-pressed={theme === themeItem.id}
              >
                {themeItem.id === 'pop-art' && 'P'}
                {themeItem.id === 'classical-chinese' && 'C'}
                {themeItem.id === 'minimalist-bw' && 'B'}
              </button>
            ))}
          </div>
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
                aria-label={`Switch to ${lang.nativeName}`}
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
