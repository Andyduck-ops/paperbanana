import { useLanguage } from '../../hooks';

export interface EmptyStateProps {
  mode: 'generate' | 'refine';
}

/**
 * Empty State Component - Minimal
 *
 * A clean, unobtrusive welcome message.
 * The input form is always visible below this component.
 */
export function EmptyState({
  mode,
}: EmptyStateProps) {
  const { t } = useLanguage();

  return (
    <div className="empty-state text-center py-8 px-6">
      <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-primary/10 mb-3">
        {mode === 'generate' ? (
          <svg className="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
          </svg>
        ) : (
          <svg className="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
          </svg>
        )}
      </div>
      <h2 className="text-lg font-heading font-semibold text-foreground mb-1">
        {mode === 'generate'
          ? t('emptyState.generateTitle')
          : t('emptyState.refineTitle')}
      </h2>
      <p className="text-sm text-muted-foreground max-w-md mx-auto">
        {mode === 'generate'
          ? t('emptyState.generateDescription')
          : t('emptyState.refineDescription')}
      </p>
    </div>
  );
}

export default EmptyState;
