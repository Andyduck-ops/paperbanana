import { useLanguage } from '../hooks';

export function Footer() {
  const { t } = useLanguage();
  const year = new Date().getFullYear();

  return (
    <div className="px-4 py-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 text-sm text-muted-foreground">
      <p>
        {t('app.name')} &copy; {year}
      </p>
      <div className="flex gap-4">
        <a
          href="https://github.com/example/paperbanana"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-primary transition-colors"
        >
          GitHub
        </a>
        <a
          href="#"
          className="hover:text-primary transition-colors"
        >
          Docs
        </a>
      </div>
    </div>
  );
}
