import { useState, useEffect, useCallback } from 'react';

export type Theme = 'academic' | 'qi-baishi' | 'pop-anime' | 'rococo' | 'japanese-bw';

const THEME_STORAGE_KEY = 'paperbanana-theme';

// Migration map for old theme IDs
const THEME_MIGRATION_MAP: Record<string, Theme> = {
  'pop-art': 'pop-anime',
  'classical-chinese': 'qi-baishi',
  'minimalist-bw': 'japanese-bw',
};

function migrateTheme(oldTheme: string): Theme {
  return THEME_MIGRATION_MAP[oldTheme] || (isValidTheme(oldTheme) ? oldTheme : 'academic');
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
  // Default theme: Academic represents professional baseline
  return 'academic';
}

function isValidTheme(theme: string): theme is Theme {
  return ['academic', 'qi-baishi', 'pop-anime', 'rococo', 'japanese-bw'].includes(theme);
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
      { id: 'academic' as const, name: 'Academic', swatch: { bg: '#f6f2ea', ink: '#2a4a7a', accent: '#6a9a7a' } },
      { id: 'qi-baishi' as const, name: 'Qi Baishi', swatch: { bg: '#fcfaf5', ink: '#c44a3a', accent: '#5a9a7a' } },
      { id: 'pop-anime' as const, name: 'Pop Anime', swatch: { bg: '#f9f2dc', ink: '#e83a2a', accent: '#2a6aea' } },
      { id: 'rococo' as const, name: 'Rococo', swatch: { bg: '#fffcf7', ink: '#c48a8a', accent: '#c4a45a' } },
      { id: 'japanese-bw' as const, name: 'Night Mono', swatch: { bg: '#1c1c1c', ink: '#ebebeb', accent: '#a0a0a0' } },
    ] as const,
  };
}
