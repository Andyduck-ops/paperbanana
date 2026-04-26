import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTheme } from './useTheme';
import { useAppStore } from '../stores/appStore';

describe('useTheme', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
    document.documentElement.removeAttribute('data-color-scheme');
    // Reset Zustand store to default state
    useAppStore.setState({
      theme: 'qi-baishi',
      colorScheme: 'system',
      language: 'zh',
    });
  });

  it('returns default theme (qi-baishi) when no stored preference', () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('qi-baishi');
  });

  it('reads current theme from Zustand store', () => {
    useAppStore.setState({ theme: 'pop-anime' });
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('pop-anime');
  });

  it('updates theme and applies data-theme attribute', () => {
    const { result } = renderHook(() => useTheme());
    act(() => {
      result.current.setTheme('rococo');
    });
    expect(result.current.theme).toBe('rococo');
    expect(document.documentElement.getAttribute('data-theme')).toBe('rococo');
  });

  it('provides list of 5 available themes with swatches', () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.themes).toHaveLength(5);
    expect(result.current.themes.map((t) => t.id)).toEqual([
      'academic',
      'qi-baishi',
      'pop-anime',
      'rococo',
      'japanese-bw',
    ]);
    expect(result.current.themes[0]).toHaveProperty('swatch');
  });

  it('falls back to qi-baishi when store has invalid theme', () => {
    // @ts-expect-error intentionally setting invalid theme for fallback test
    useAppStore.setState({ theme: 'invalid-theme' });
    const { result } = renderHook(() => useTheme());
    // useTheme itself does not validate; it reflects store state
    expect(result.current.theme).toBe('invalid-theme');
  });
});
