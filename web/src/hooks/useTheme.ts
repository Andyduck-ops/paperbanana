import { useState, useEffect, useCallback } from 'react';

export type Theme = 'pop-art' | 'classical-chinese' | 'minimalist-bw';

const THEME_STORAGE_KEY = 'paperbanana-theme';

function getInitialTheme(): Theme {
  // Check localStorage first
  if (typeof window !== 'undefined') {
    const stored = localStorage.getItem(THEME_STORAGE_KEY) as Theme | null;
    if (stored && isValidTheme(stored)) {
      return stored;
    }
  }
  // Default theme
  return 'pop-art';
}

function isValidTheme(theme: string): theme is Theme {
  return ['pop-art', 'classical-chinese', 'minimalist-bw'].includes(theme);
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
      { id: 'pop-art', name: 'Pop Art' },
      { id: 'classical-chinese', name: 'Classical Chinese' },
      { id: 'minimalist-bw', name: 'Minimalist B&W' },
    ] as const,
  };
}
