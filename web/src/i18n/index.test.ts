import { describe, it, expect } from 'vitest';
import en from './locales/en.json';
import zh from './locales/zh.json';

// Helper to flatten nested object keys
function flattenKeys(obj: Record<string, unknown>, prefix = ''): string[] {
  const keys: string[] = [];
  for (const key in obj) {
    const fullKey = prefix ? `${prefix}.${key}` : key;
    if (typeof obj[key] === 'object' && obj[key] !== null) {
      keys.push(...flattenKeys(obj[key] as Record<string, unknown>, fullKey));
    } else {
      keys.push(fullKey);
    }
  }
  return keys;
}

describe('i18n configuration', () => {
  it('has matching keys in both languages', () => {
    const enKeys = flattenKeys(en).sort();
    const zhKeys = flattenKeys(zh).sort();
    expect(enKeys).toEqual(zhKeys);
  });

  it('has no empty translation values in English', () => {
    const checkEmpty = (obj: Record<string, unknown>, path = ''): string[] => {
      const emptyPaths: string[] = [];
      for (const key in obj) {
        const fullPath = path ? `${path}.${key}` : key;
        if (typeof obj[key] === 'string' && obj[key] === '') {
          emptyPaths.push(fullPath);
        } else if (typeof obj[key] === 'object') {
          emptyPaths.push(...checkEmpty(obj[key] as Record<string, unknown>, fullPath));
        }
      }
      return emptyPaths;
    };
    const emptyEn = checkEmpty(en);
    const emptyZh = checkEmpty(zh);
    expect(emptyEn, `Empty values in en.json: ${emptyEn.join(', ')}`).toHaveLength(0);
    expect(emptyZh, `Empty values in zh.json: ${emptyZh.join(', ')}`).toHaveLength(0);
  });

  it('has required translation keys', () => {
    const requiredKeys = [
      'app.name',
      'app.tagline',
      'theme.title',
      'theme.popArt',
      'theme.classicalChinese',
      'theme.minimalistBw',
      'language.title',
      'common.loading',
      'common.error',
    ];
    const enKeys = flattenKeys(en);
    requiredKeys.forEach(key => {
      expect(enKeys).toContain(key);
    });
  });
});
