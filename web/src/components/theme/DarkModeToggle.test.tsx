import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDarkMode } from './DarkModeToggle';
import { useAppStore } from '../../stores/appStore';

type MediaListener = (event: MediaQueryListEvent) => void;

function installMatchMedia(initialDark: boolean) {
  let matches = initialDark;
  const listeners = new Set<MediaListener>();

  const mediaQuery = {
    get matches() {
      return matches;
    },
    media: '(prefers-color-scheme: dark)',
    onchange: null,
    addEventListener: (_event: string, handler: MediaListener) => {
      listeners.add(handler);
    },
    removeEventListener: (_event: string, handler: MediaListener) => {
      listeners.delete(handler);
    },
    addListener: (handler: MediaListener) => listeners.add(handler),
    removeListener: (handler: MediaListener) => listeners.delete(handler),
    dispatchEvent: () => true,
  };

  Object.defineProperty(window, 'matchMedia', {
    configurable: true,
    writable: true,
    value: vi.fn().mockReturnValue(mediaQuery),
  });

  return {
    setPrefersDark(next: boolean) {
      matches = next;
      listeners.forEach((listener) =>
        listener({ matches, media: mediaQuery.media } as MediaQueryListEvent)
      );
    },
  };
}

describe('useDarkMode (color-scheme tri-state)', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-color-scheme');
    document.documentElement.removeAttribute('data-theme');
    useAppStore.setState({ colorScheme: 'system', language: 'zh' });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('defaults to system and resolves to OS preference', () => {
    installMatchMedia(false);
    const { result } = renderHook(() => useDarkMode());
    expect(result.current.scheme).toBe('system');
    expect(result.current.effectiveScheme).toBe('light');
  });

  it('reflects dark OS preference when scheme is system', () => {
    installMatchMedia(true);
    const { result } = renderHook(() => useDarkMode());
    expect(result.current.effectiveScheme).toBe('dark');
  });

  it('forcing light overrides OS dark preference', () => {
    installMatchMedia(true);
    const { result } = renderHook(() => useDarkMode());
    act(() => {
      result.current.setScheme('light');
    });
    expect(result.current.scheme).toBe('light');
    expect(result.current.effectiveScheme).toBe('light');
    expect(document.documentElement.getAttribute('data-color-scheme')).toBe('light');
  });

  it('forcing dark overrides OS light preference', () => {
    installMatchMedia(false);
    const { result } = renderHook(() => useDarkMode());
    act(() => {
      result.current.setScheme('dark');
    });
    expect(result.current.effectiveScheme).toBe('dark');
    expect(document.documentElement.getAttribute('data-color-scheme')).toBe('dark');
  });

  it('reacts to OS preference changes when scheme is system', () => {
    const media = installMatchMedia(false);
    renderHook(() => useDarkMode());
    act(() => {
      media.setPrefersDark(true);
    });
    expect(document.documentElement.getAttribute('data-color-scheme')).toBe('dark');
  });

  it('does not write data-theme (legacy attribute is forbidden)', () => {
    installMatchMedia(false);
    const { result } = renderHook(() => useDarkMode());
    act(() => {
      result.current.setScheme('dark');
    });
    expect(document.documentElement.getAttribute('data-theme')).toBeNull();
  });
});
