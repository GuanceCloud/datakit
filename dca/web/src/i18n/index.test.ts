import fs from 'fs';
import path from 'path';

describe('i18n source', () => {
  it('configures fallback language detection and toggle logic', () => {
    const source = fs.readFileSync(path.join(__dirname, 'index.ts'), 'utf8');

    expect(source).toContain('fallbackLng');
    expect(source).toContain('LanguageDetector');
    expect(source).toContain("order: ['sessionStorage', 'localStorage']");
    expect(source).toContain('initLanguage');
    expect(source).toContain('toggleLanguage');
    expect(source).toContain('i18n.changeLanguage');
  });
});
