import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { App } from './App';

// Mock i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    i18n: {
      language: 'zh',
      changeLanguage: vi.fn(),
    },
    t: (key: string) => {
      const translations: Record<string, string> = {
        'app.name': 'PaperBanana',
        'app.tagline': '学术论文可视化生成器',
      };
      return translations[key] || key;
    },
  }),
  initReactI18next: {
    type: '3rdParty',
    init: vi.fn(),
  },
}));

// Mock hooks
vi.mock('./hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'app.name': 'PaperBanana',
        'app.tagline': '学术论文可视化生成器',
        'theme.title': '选择主题',
        'language.title': '语言',
        'history.title': '历史记录',
        'history.empty': '暂无历史记录',
        'history.untitled': '未命名',
        'common.loading': '加载中...',
      };
      return translations[key] || key;
    },
    language: 'zh',
    setLanguage: vi.fn(),
    languages: [
      { code: 'zh', name: 'Chinese', nativeName: '简体中文' },
      { code: 'en', name: 'English', nativeName: 'English' },
    ],
  }),
  useTheme: () => ({
    theme: 'pop-art',
    setTheme: vi.fn(),
    themes: [
      { id: 'pop-art', name: 'Pop Art' },
      { id: 'classical-chinese', name: 'Classical Chinese' },
      { id: 'minimalist-bw', name: 'Minimalist B&W' },
    ],
  }),
  useGenerate: () => ({
    isGenerating: false,
    stages: [],
    result: null,
    error: null,
    generate: vi.fn(),
    reset: vi.fn(),
  }),
  useBatchGeneration: () => ({
    isGenerating: false,
    progress: null,
    result: null,
    error: null,
    startBatch: vi.fn(),
    resetBatch: vi.fn(),
  }),
  useRefine: () => ({
    isRefining: false,
    result: null,
    refine: vi.fn(),
    reset: vi.fn(),
  }),
  useHistory: () => ({
    sessions: [],
    isLoading: false,
    error: null,
    refresh: vi.fn(),
  }),
  useToast: () => ({
    toasts: [],
    addToast: vi.fn(),
    removeToast: vi.fn(),
  }),
  useKeyboardShortcuts: vi.fn(),
}));

vi.mock('./hooks/useProviders', () => ({
  useProviders: () => ({
    providers: [],
    loading: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

describe('App', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('renders app with header', () => {
    render(<App />);
    expect(screen.getByText('PaperBanana')).toBeInTheDocument();
  });

  it('renders generate panel with dual textareas', () => {
    render(<App />);
    // DualInputPanel provides two textareas
    const textareas = screen.getAllByRole('textbox');
    expect(textareas.length).toBeGreaterThanOrEqual(2);
  });

  it('renders history sidebar', () => {
    render(<App />);
    expect(screen.getByText('历史记录')).toBeInTheDocument();
  });
});
