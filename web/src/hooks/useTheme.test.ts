import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTheme } from './useTheme';

describe('useTheme', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  it('returns default theme (qi-baishi) when no stored preference', () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('qi-baishi');
  });

  it('restores theme from localStorage', () => {
    localStorage.setItem('paperbanana-theme', 'pop-anime');
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('pop-anime');
  });

  it('migrates old theme ID pop-art to pop-anime', () => {
    localStorage.setItem('paperbanana-theme', 'pop-art');
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('pop-anime');
    expect(localStorage.getItem('paperbanana-theme')).toBe('pop-anime');
  });

  it('migrates old theme ID classical-chinese to qi-baishi', () => {
    localStorage.setItem('paperbanana-theme', 'classical-chinese');
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('qi-baishi');
    expect(localStorage.getItem('paperbanana-theme')).toBe('qi-baishi');
  });

  it('migrates old theme ID minimalist-bw to japanese-bw', () => {
    localStorage.setItem('paperbanana-theme', 'minimalist-bw');
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('japanese-bw');
    expect(localStorage.getItem('paperbanana-theme')).toBe('japanese-bw');
  });

  it('updates theme and persists to localStorage', () => {
    const { result } = renderHook(() => useTheme());
    act(() => {
      result.current.setTheme('rococo');
    });
    expect(result.current.theme).toBe('rococo');
    expect(localStorage.getItem('paperbanana-theme')).toBe('rococo');
    expect(document.documentElement.getAttribute('data-theme')).toBe('rococo');
  });

  it('provides list of exactly 4 available themes', () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.themes).toHaveLength(4);
    expect(result.current.themes[0].id).toBe('qi-baishi');
    expect(result.current.themes[1].id).toBe('pop-anime');
    expect(result.current.themes[2].id).toBe('rococo');
    expect(result.current.themes[3].id).toBe('japanese-bw');
  });

  it('falls back to qi-baishi for invalid stored themes', () => {
    localStorage.setItem('paperbanana-theme', 'invalid-theme');
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('qi-baishi');
  });
});
