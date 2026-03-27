import { useState, useEffect, useCallback } from 'react';

export type Theme = 'qi-baishi' | 'pop-anime' | 'rococo' | 'japanese-bw';

const THEME_STORAGE_KEY = 'paperbanana-theme';

// Migration map for old theme IDs
const THEME_MIGRATION_MAP: Record<string, Theme> = {
  'pop-art': 'pop-anime',
  'classical-chinese': 'qi-baishi',
  'minimalist-bw': 'japanese-bw',
};

function migrateTheme(oldTheme: string): Theme {
  return THEME_MIGRATION_MAP[oldTheme] || (isValidTheme(oldTheme) ? oldTheme : 'qi-baishi');
}

function getInitialTheme(): Theme {
  // Check localStorage first
  if (typeof window !== 'undefined') {
    const stored = localStorage.getItem(THEME_STORAGE_KEY);
    if (stored) {
      // Migrate old theme IDs to new ones
      const migrated = migrateTheme(stored);
      if (migrated !== stored) {
        // Update localStorage with new theme ID
        localStorage.setItem(THEME_STORAGE_KEY, migrated);
      }
      return migrated;
    }
  }
  // Default theme: Qi Baishi represents academic paper aesthetic
  return 'qi-baishi';
}

function isValidTheme(theme: string): theme is Theme {
  return ['qi-baishi', 'pop-anime', 'rococo', 'japanese-bw'].includes(theme);
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(getInitialTheme);

  const setTheme = useCallback((newTheme: Theme) => {
    setThemeState(newTheme);
    if (typeof window !== 'undefined') {
      localStorage.setItem(THEME_STORAGE_KEY, newTheme);
      document.documentElement.setAttribute('data-theme', newTheme);
    }
  }, []);

  // Apply theme on mount and changes
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  return {
    theme,
    setTheme,
    themes: [
      { id: 'qi-baishi' as const, name: 'Qi Baishi' },
      { id: 'pop-anime' as const, name: 'Pop Anime' },
      { id: 'rococo' as const, name: 'Rococo' },
      { id: 'japanese-bw' as const, name: 'Night Mono' },
    ] as const,
  };
}
