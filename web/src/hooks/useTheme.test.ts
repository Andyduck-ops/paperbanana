import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTheme } from './useTheme';

describe('useTheme', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  it('returns default theme when no stored preference', () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('pop-art');
  });

  it('restores theme from localStorage', () => {
    localStorage.setItem('paperbanana-theme', 'classical-chinese');
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('classical-chinese');
  });

  it('updates theme and persists to localStorage', () => {
    const { result } = renderHook(() => useTheme());
    act(() => {
      result.current.setTheme('minimalist-bw');
    });
    expect(result.current.theme).toBe('minimalist-bw');
    expect(localStorage.getItem('paperbanana-theme')).toBe('minimalist-bw');
    expect(document.documentElement.getAttribute('data-theme')).toBe('minimalist-bw');
  });

  it('provides list of available themes', () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.themes).toHaveLength(4);
    expect(result.current.themes[0].id).toBe('pop-art');
  });
});
