/**
 * 动态主题加载 Hook
 * 
 * 使用方法:
 * 1. 从 App.tsx 移除所有主题 CSS 的静态导入
 * 2. 在需要主题选择的地方使用此 hook
 * 3. 主题 CSS 将按需加载
 */

import { useState, useEffect, useCallback } from 'react';

export type Theme = 'qi-baishi' | 'pop-anime' | 'rococo' | 'japanese-bw';

const THEME_STORAGE_KEY = 'paperbanana-theme';

// 旧主题 ID 迁移映射
const THEME_MIGRATION_MAP: Record<string, Theme> = {
  'pop-art': 'pop-anime',
  'classical-chinese': 'qi-baishi',
  'minimalist-bw': 'japanese-bw',
};

function migrateTheme(oldTheme: string): Theme {
  return THEME_MIGRATION_MAP[oldTheme] || (isValidTheme(oldTheme) ? oldTheme : 'qi-baishi');
}

function isValidTheme(theme: string): theme is Theme {
  return ['qi-baishi', 'pop-anime', 'rococo', 'japanese-bw'].includes(theme);
}

function getInitialTheme(): Theme {
  if (typeof window === 'undefined') return 'qi-baishi';
  
  const stored = localStorage.getItem(THEME_STORAGE_KEY);
  if (stored) {
    const migrated = migrateTheme(stored);
    if (migrated !== stored) {
      localStorage.setItem(THEME_STORAGE_KEY, migrated);
    }
    return migrated;
  }
  return 'qi-baishi';
}

export interface UseDynamicThemeReturn {
  /** 当前主题 */
  theme: Theme;
  /** 设置主题 */
  setTheme: (theme: Theme) => Promise<void>;
  /** 是否正在加载主题 */
  isLoading: boolean;
  /** 主题是否已就绪 */
  isReady: boolean;
  /** 加载错误信息 */
  error: string | null;
  /** 可用主题列表 */
  themes: readonly { id: Theme; name: string }[];
}

/**
 * 动态主题加载 Hook
 * 
 * 特性:
 * - 按需加载主题 CSS，减少首屏加载
 * - 自动迁移旧主题 ID
 * - 加载失败时回退到默认主题
 * - 缓存已加载的主题
 */
export function useDynamicTheme(): UseDynamicThemeReturn {
  const [theme, setThemeState] = useState<Theme>(getInitialTheme);
  const [isLoading, setIsLoading] = useState(false);
  const [isReady, setIsReady] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadedThemes, setLoadedThemes] = useState<Set<Theme>>(new Set());

  const loadTheme = useCallback(async (themeToLoad: Theme): Promise<boolean> => {
    // 已加载的主题直接返回成功
    if (loadedThemes.has(themeToLoad)) {
      return true;
    }

    setIsLoading(true);
    setError(null);

    try {
      // 动态导入主题 CSS
      // 注意: 需要确保 vite 能够正确处理这些动态导入
      await import(/* @vite-ignore */ `./../themes/${themeToLoad}.css`);
      
      setLoadedThemes(prev => new Set(prev).add(themeToLoad));
      setIsLoading(false);
      return true;
    } catch (err) {
      console.error(`[Theme] Failed to load theme: ${themeToLoad}`, err);
      setError(`Failed to load theme: ${themeToLoad}`);
      setIsLoading(false);
      return false;
    }
  }, [loadedThemes]);

  const setTheme = useCallback(async (newTheme: Theme) => {
    const success = await loadTheme(newTheme);
    if (success) {
      setThemeState(newTheme);
      localStorage.setItem(THEME_STORAGE_KEY, newTheme);
      document.documentElement.setAttribute('data-theme', newTheme);
    } else if (newTheme !== 'qi-baishi') {
      // 回退到默认主题
      const fallbackSuccess = await loadTheme('qi-baishi');
      if (fallbackSuccess) {
        setThemeState('qi-baishi');
        localStorage.setItem(THEME_STORAGE_KEY, 'qi-baishi');
        document.documentElement.setAttribute('data-theme', 'qi-baishi');
      }
    }
  }, [loadTheme]);

  // 初始加载
  useEffect(() => {
    let isMounted = true;
    
    const initTheme = async () => {
      const initialTheme = getInitialTheme();
      const success = await loadTheme(initialTheme);
      if (isMounted) {
        setIsReady(success);
        document.documentElement.setAttribute('data-theme', initialTheme);
      }
    };

    initTheme();

    return () => {
      isMounted = false;
    };
  }, [loadTheme]);

  return {
    theme,
    setTheme,
    isLoading,
    isReady,
    error,
    themes: [
      { id: 'qi-baishi' as const, name: 'Qi Baishi' },
      { id: 'pop-anime' as const, name: 'Pop Anime' },
      { id: 'rococo' as const, name: 'Rococo' },
      { id: 'japanese-bw' as const, name: 'Night Mono' },
    ] as const,
  };
}
