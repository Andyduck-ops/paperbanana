import { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { supportedLanguages } from '../i18n';

const LANGUAGE_STORAGE_KEY = 'paperbanana-language';

export type LanguageCode = 'zh' | 'en';

export function useLanguage() {
  const { i18n, t } = useTranslation();

  const setLanguage = useCallback((code: LanguageCode) => {
    i18n.changeLanguage(code);
    if (typeof window !== 'undefined') {
      localStorage.setItem(LANGUAGE_STORAGE_KEY, code);
    }
  }, [i18n]);

  // Restore language from localStorage on mount
  const restoreLanguage = useCallback(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(LANGUAGE_STORAGE_KEY) as LanguageCode | null;
      if (stored && ['zh', 'en'].includes(stored)) {
        i18n.changeLanguage(stored);
      }
    }
  }, [i18n]);

  return {
    language: i18n.language as LanguageCode,
    setLanguage,
    restoreLanguage,
    languages: supportedLanguages,
    t,
  };
}
