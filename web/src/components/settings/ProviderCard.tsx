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
      className: 'provider-card__status provider-card__status--configured',
      label: t('settings.status.configured'),
    },
    no_keys: {
      className: 'provider-card__status provider-card__status--warning',
      label: t('settings.status.no_keys'),
    },
    invalid: {
      className: 'provider-card__status provider-card__status--danger',
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
    <article className={`provider-card ${!provider.enabled ? 'provider-card--muted' : ''}`}>
      <div className="provider-card__body">
        <div className="provider-card__content">
          <div className="provider-card__heading-row">
            <h3 className="provider-card__title">{displayName}</h3>
            {provider.is_default && (
              <span className="provider-card__pill provider-card__pill--primary">{t('settings.default')}</span>
            )}
            <span className={status.className}>{status.label}</span>
            {provider.type && provider.type !== 'custom' && (
              <span className="provider-card__pill provider-card__pill--neutral">{provider.type}</span>
            )}
            {provider.is_system && (
              <span className="provider-card__meta-note">{t('settings.system')}</span>
            )}
          </div>

          {hasModels ? (
            <div className="provider-card__models-block">
              <p className="provider-card__section-label">{t('settings.models')}</p>
              <div className="provider-card__models-list">
                {provider.models!.slice(0, 5).map((model) => (
                  <span key={model.id} className="provider-card__model-chip">
                    {model.id}
                  </span>
                ))}
                {provider.models!.length > 5 && (
                  <span className="provider-card__model-chip provider-card__model-chip--muted">
                    +{provider.models!.length - 5} more
                  </span>
                )}
              </div>
            </div>
          ) : hasLegacyModels ? (
            <div className="provider-card__legacy-models">
              {provider.query_model && (
                <p className="provider-card__legacy-line">
                  <span className="provider-card__section-label">{t('settings.queryModel')}</span>
                  <span className="provider-card__inline-chip">{provider.query_model}</span>
                </p>
              )}
              {provider.gen_model && provider.gen_model !== provider.query_model && (
                <p className="provider-card__legacy-line">
                  <span className="provider-card__section-label">{t('settings.genModel')}</span>
                  <span className="provider-card__inline-chip">{provider.gen_model}</span>
                </p>
              )}
            </div>
          ) : (
            <p className="provider-card__empty-copy">{t('settings.noModels')}</p>
          )}

          {provider.base_url && <p className="provider-card__endpoint">{provider.base_url}</p>}
        </div>

        <div className="provider-card__actions">
          {onToggle && (
            <label className="provider-card__toggle">
              <input type="checkbox" checked={provider.enabled} onChange={onToggle} className="sr-only peer" />
              <span className="provider-card__toggle-track" />
            </label>
          )}

          {onSetDefault && !provider.is_default && provider.enabled && provider.status === 'configured' && (
            <button onClick={onSetDefault} className="provider-card__action provider-card__action--ghost">
              {t('settings.setAsDefault')}
            </button>
          )}

          <button onClick={onEdit} className="provider-card__action provider-card__action--secondary">
            {t('common.edit')}
          </button>

          {canDelete && (
            <button onClick={onDelete} className="provider-card__action provider-card__action--danger">
              {t('common.delete')}
            </button>
          )}

          {canRemoveKeys && onRemoveKeys && (
            <button onClick={onRemoveKeys} className="provider-card__action provider-card__action--warning">
              {t('settings.removeKeys')}
            </button>
          )}
        </div>
      </div>
    </article>
  );
}
