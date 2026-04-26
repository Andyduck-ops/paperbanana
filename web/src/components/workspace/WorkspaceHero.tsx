import { useLanguage } from '../../hooks';

export interface WorkspaceHeroProps {
  onOpenSettings?: () => void;
  onOpenHistory?: () => void;
}

export function WorkspaceHero({ onOpenSettings, onOpenHistory }: WorkspaceHeroProps) {
  const { t } = useLanguage();

  return (
    <section className="workspace-hero rounded-[2rem] border border-border/70 bg-card/95 px-5 py-5 shadow-[var(--theme-surface-shadow)] sm:px-7 sm:py-7">
      <div className="workspace-hero__grid grid gap-5 lg:grid-cols-[minmax(0,1.6fr)_minmax(18rem,0.9fr)] lg:items-start">
        <div className="space-y-4">
          <div className="space-y-3">
            <p className="text-[0.7rem] font-semibold uppercase tracking-[0.22em] text-primary/80">
              {t('workspace.kicker')}
            </p>
            <div className="space-y-2">
              <h1 className="max-w-3xl text-3xl font-semibold leading-tight text-foreground sm:text-[2.4rem]">
                {t('workspace.heroTitle')}
              </h1>
              <p className="max-w-2xl text-sm leading-6 text-muted-foreground sm:text-[0.95rem]">
                {t('workspace.heroDescription')}
              </p>
            </div>
          </div>

          <div className="grid gap-3 sm:grid-cols-3">
            <div className="rounded-2xl border border-border/60 bg-background/75 px-4 py-3">
              <div className="text-[0.72rem] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                {t('workspace.heroCards.contextLabel')}
              </div>
              <p className="mt-2 text-sm leading-6 text-foreground">
                {t('workspace.heroCards.contextBody')}
              </p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-background/75 px-4 py-3">
              <div className="text-[0.72rem] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                {t('workspace.heroCards.outputLabel')}
              </div>
              <p className="mt-2 text-sm leading-6 text-foreground">
                {t('workspace.heroCards.outputBody')}
              </p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-background/75 px-4 py-3">
              <div className="text-[0.72rem] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                {t('workspace.heroCards.workflowLabel')}
              </div>
              <p className="mt-2 text-sm leading-6 text-foreground">
                {t('workspace.heroCards.workflowBody')}
              </p>
            </div>
          </div>
        </div>

        <aside className="rounded-[1.75rem] border border-border/70 bg-background/85 p-4">
          <div className="text-[0.72rem] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
            {t('workspace.quickActionsTitle')}
          </div>
          <div className="mt-3 space-y-3">
            <div className="rounded-2xl bg-muted/60 px-4 py-3">
              <div className="text-sm font-medium text-foreground">{t('workspace.quickActions.primaryLabel')}</div>
              <p className="mt-1 text-sm leading-6 text-muted-foreground">
                {t('workspace.quickActions.primaryBody')}
              </p>
            </div>
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-1">
              <button
                type="button"
                onClick={onOpenHistory}
                className="flex min-h-11 items-center justify-between rounded-2xl border border-border/70 bg-card px-4 py-3 text-left text-sm text-foreground transition-colors hover:border-primary/40 hover:bg-primary/5"
              >
                <span>{t('workspace.quickActions.history')}</span>
                <span aria-hidden="true">&rarr;</span>
              </button>
              <button
                type="button"
                onClick={onOpenSettings}
                className="flex min-h-11 items-center justify-between rounded-2xl border border-border/70 bg-card px-4 py-3 text-left text-sm text-foreground transition-colors hover:border-primary/40 hover:bg-primary/5"
              >
                <span>{t('workspace.quickActions.settings')}</span>
                <span aria-hidden="true">&rarr;</span>
              </button>
            </div>
          </div>
        </aside>
      </div>
    </section>
  );
}
