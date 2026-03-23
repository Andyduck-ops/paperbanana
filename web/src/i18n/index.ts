import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import en from './locales/en.json';
import zh from './locales/zh.json';

export const resources = {
  en: { translation: en },
  zh: { translation: zh },
} as const;

export const defaultNS = 'translation' as const;

export const supportedLanguages = [
  { code: 'zh', name: '中文', nativeName: '简体中文' },
  { code: 'en', name: 'English', nativeName: 'English' },
] as const;

i18n
  .use(initReactI18next)
  .init({
    resources,
    lng: 'zh', // Default to Chinese for primary user base
    fallbackLng: 'en',
    defaultNS,
    interpolation: {
      escapeValue: false, // React handles XSS protection
    },
    // Return key if translation missing (helpful for debugging)
    returnEmptyString: false,
    returnNull: false,
  });

export default i18n;
