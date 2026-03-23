import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useLanguage } from './useLanguage';

// Create a mutable mock i18n object
const mockI18n = {
  language: 'zh',
  changeLanguage: vi.fn((lang: string) => {
    mockI18n.language = lang;
  }),
};

// Mock i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    i18n: mockI18n,
    t: (key: string) => key,
  }),
  initReactI18next: {
    type: '3rdParty',
    init: vi.fn(),
  },
}));

describe('useLanguage', () => {
  beforeEach(() => {
    localStorage.clear();
    mockI18n.language = 'zh';
    vi.clearAllMocks();
  });

  it('returns current language', () => {
    const { result } = renderHook(() => useLanguage());
    expect(result.current.language).toBe('zh');
  });

  it('provides list of supported languages', () => {
    const { result } = renderHook(() => useLanguage());
    expect(result.current.languages).toHaveLength(2);
    expect(result.current.languages[0].code).toBe('zh');
  });

  it('setLanguage persists to localStorage', () => {
    const { result } = renderHook(() => useLanguage());
    act(() => {
      result.current.setLanguage('en');
    });
    expect(localStorage.getItem('paperbanana-language')).toBe('en');
    expect(mockI18n.changeLanguage).toHaveBeenCalledWith('en');
  });
});
