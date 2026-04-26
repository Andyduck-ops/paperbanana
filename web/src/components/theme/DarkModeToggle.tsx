import { useCallback, useEffect, useMemo } from 'react';
import { useAppStore, type ColorScheme } from '../../stores/appStore';

function getSystemPreference(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'light';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function getEffectiveScheme(scheme: ColorScheme): 'light' | 'dark' {
  if (scheme === 'system') return getSystemPreference();
  return scheme;
}

export function useDarkMode() {
  const scheme = useAppStore((state) => state.colorScheme);
  const storeSetColorScheme = useAppStore((state) => state.setColorScheme);

  const effectiveScheme = useMemo(() => getEffectiveScheme(scheme), [scheme]);

  const setScheme = useCallback((newScheme: ColorScheme) => {
    storeSetColorScheme(newScheme);
  }, [storeSetColorScheme]);

  // Listen for system preference changes when in system mode
  useEffect(() => {
    if (typeof window === 'undefined') return;
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => {
      if (scheme === 'system') {
        const effective = getSystemPreference();
        document.documentElement.setAttribute('data-color-scheme', effective);
      }
    };
    mediaQuery.addEventListener('change', handler);
    return () => mediaQuery.removeEventListener('change', handler);
  }, [scheme]);

  return { scheme, effectiveScheme, setScheme };
}

export interface DarkModeToggleProps {
  className?: string;
}

export function DarkModeToggle({ className = '' }: DarkModeToggleProps) {
  const { scheme, setScheme } = useDarkMode();

  const options: { value: ColorScheme; label: string; icon: string }[] = [
    { value: 'light', label: 'Light', icon: 'M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z' },
    { value: 'dark', label: 'Dark', icon: 'M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z' },
    { value: 'system', label: 'Auto', icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
  ];

  return (
    <div className={`inline-flex items-center gap-0.5 p-0.5 rounded-full border border-border bg-muted ${className}`}>
      {options.map((option) => (
        <button
          key={option.value}
          onClick={() => setScheme(option.value)}
          className={`
            flex items-center gap-1 px-2.5 py-1.5 rounded-full text-xs font-medium transition-all duration-200
            ${scheme === option.value
              ? 'bg-primary text-background shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
            }
          `}
          aria-pressed={scheme === option.value}
          title={option.label}
        >
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d={option.icon} />
          </svg>
          <span className="hidden sm:inline">{option.label}</span>
        </button>
      ))}
    </div>
  );
}
