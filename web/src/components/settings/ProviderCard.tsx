import { useTranslation } from 'react-i18next';
import type { Provider } from '../../hooks/useProviders';

interface ProviderCardProps {
  provider: Provider;
  onEdit: () => void;
  onDelete: () => void;
  onToggle?: () => void;
  onSetDefault?: () => void;
  onRemoveKeys?: () => void;
}

export function ProviderCard({
  provider,
  onEdit,
  onDelete,
  onToggle,
  onSetDefault,
  onRemoveKeys,
}: ProviderCardProps) {
  const { t } = useTranslation();

  const statusConfig = {
    configured: {
      className: 'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-status-success/10 text-status-success',
      label: t('settings.status.configured'),
    },
    no_keys: {
      className: 'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-status-warning/10 text-status-warning',
      label: t('settings.status.no_keys'),
    },
    invalid: {
      className: 'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-status-error/10 text-status-error',
      label: t('settings.status.invalid'),
    },
  };

  const status = statusConfig[provider.status] || statusConfig.invalid;
  const displayName = provider.display_name || provider.name;
  const hasModels = provider.models && provider.models.length > 0;
  const hasLegacyModels = !hasModels && (provider.query_model || provider.gen_model);
  const canDelete = !provider.is_system;
  const canRemoveKeys = provider.is_system && (provider.status === 'configured' || provider.enabled || provider.is_default);

  return (
    <article className={`rounded-xl border border-border/50 bg-card p-4 transition-opacity ${!provider.enabled ? 'opacity-60' : ''}`}>
      <div className="flex flex-col sm:flex-row sm:items-start gap-4">
        <div className="flex-1 min-w-0 space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="text-base font-semibold text-foreground">{displayName}</h3>
            {provider.is_default && (
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary">{t('settings.default')}</span>
            )}
            <span className={status.className}>{status.label}</span>
            {provider.type && provider.type !== 'custom' && (
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground">{provider.type}</span>
            )}
            {provider.is_system && (
              <span className="text-xs text-muted-foreground">{t('settings.system')}</span>
            )}
          </div>

          {hasModels ? (
            <div className="space-y-1.5">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{t('settings.models')}</p>
              <div className="flex flex-wrap gap-1.5">
                {provider.models!.slice(0, 5).map((model) => (
                  <span key={model.id} className="inline-flex items-center px-2 py-0.5 rounded-md text-xs bg-muted/70 text-foreground border border-border/40">
                    {model.id}
                  </span>
                ))}
                {provider.models!.length > 5 && (
                  <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs bg-muted/50 text-muted-foreground">
                    +{provider.models!.length - 5} more
                  </span>
                )}
              </div>
            </div>
          ) : hasLegacyModels ? (
            <div className="space-y-1.5">
              {provider.query_model && (
                <p className="flex items-center gap-2 text-sm">
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{t('settings.queryModel')}</span>
                  <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs bg-muted/70 text-foreground border border-border/40">{provider.query_model}</span>
                </p>
              )}
              {provider.gen_model && provider.gen_model !== provider.query_model && (
                <p className="flex items-center gap-2 text-sm">
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{t('settings.genModel')}</span>
                  <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs bg-muted/70 text-foreground border border-border/40">{provider.gen_model}</span>
                </p>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">{t('settings.noModels')}</p>
          )}

          {provider.base_url && <p className="text-xs text-muted-foreground truncate">{provider.base_url}</p>}
        </div>

        <div className="flex flex-wrap items-center gap-2 sm:justify-end">
          {onToggle && (
            <label className="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" checked={provider.enabled} onChange={onToggle} className="sr-only peer" />
              <span className="w-9 h-5 bg-muted peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary/50 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-primary" />
            </label>
          )}

          {onSetDefault && !provider.is_default && provider.enabled && provider.status === 'configured' && (
            <button onClick={onSetDefault} className="inline-flex items-center px-3 py-1.5 text-xs font-medium text-muted-foreground hover:text-foreground bg-muted/50 hover:bg-muted rounded-lg transition-colors">
              {t('settings.setAsDefault')}
            </button>
          )}

          <button onClick={onEdit} className="inline-flex items-center px-3 py-1.5 text-xs font-medium text-foreground bg-muted/70 hover:bg-muted border border-border/50 rounded-lg transition-colors">
            {t('common.edit')}
          </button>

          {canDelete && (
            <button onClick={onDelete} className="inline-flex items-center px-3 py-1.5 text-xs font-medium text-status-error bg-status-error/10 hover:bg-status-error/20 rounded-lg transition-colors">
              {t('common.delete')}
            </button>
          )}

          {canRemoveKeys && onRemoveKeys && (
            <button onClick={onRemoveKeys} className="inline-flex items-center px-3 py-1.5 text-xs font-medium text-status-warning bg-status-warning/10 hover:bg-status-warning/20 rounded-lg transition-colors">
              {t('settings.removeKeys')}
            </button>
          )}
        </div>
      </div>
    </article>
  );
}
