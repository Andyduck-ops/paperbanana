import { useState } from 'react';

export interface ThemeSelectorProps {
  themes?: Array<{ id: string; name: string }>;
  selectedThemeId?: string;
  onThemeChange?: (themeId: string) => void;
  className?: string;
}

export function ThemeSelector({
  themes = [
    { id: 'base', name: 'Base' },
    { id: 'pop-anime', name: 'Pop Anime' },
    { id: 'classical-chinese', name: 'Classical Chinese' },
    { id: 'minimalist-bw', name: 'Minimalist B&W' },
  ],
  selectedThemeId = 'base',
  onThemeChange,
  className = ''
}: ThemeSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const selectedTheme = themes.find(t => t.id === selectedThemeId);

  return (
    <div className={`theme-selector ${className}`}>
      <button
        className="theme-selector__trigger"
        onClick={() => setIsOpen(!isOpen)}
      >
        {selectedTheme?.name || 'Select Theme'}
      </button>
      {isOpen && (
        <div className="theme-selector__dropdown">
          {themes.map((theme) => (
            <button
              key={theme.id}
              className={`theme-selector__option ${selectedThemeId === theme.id ? 'selected' : ''}`}
              onClick={() => {
                onThemeChange?.(theme.id);
                setIsOpen(false);
              }}
            >
              {theme.name}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
