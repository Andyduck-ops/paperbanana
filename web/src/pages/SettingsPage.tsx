import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useProviders, usePresets } from '../hooks/useProviders';
import { ProviderCard } from '../components/settings/ProviderCard';
import { DangerZone } from '../components/settings/DangerZone';

interface SettingsPageProps {
  onBack: () => void;
  onAddProvider: () => void;
  onEditProvider: (id: string) => void;
}

export function SettingsPage({ onBack, onAddProvider, onEditProvider }: SettingsPageProps) {
  const { t } = useTranslation();
  const { providers, loading, error, refetch } = useProviders();
  const { presets } = usePresets();
  const [enabledOverrides, setEnabledOverrides] = useState<Record<string, boolean>>({});

  const configuredProviders = providers
    .map((provider) => ({
      ...provider,
      enabled: enabledOverrides[provider.id] !== undefined ? enabledOverrides[provider.id] : provider.enabled,
    }))
    .filter((provider) => !provider.is_system || provider.status === 'configured' || provider.is_default || provider.enabled);

  const handleDelete = async (id: string) => {
    if (!window.confirm(t('settings.confirmDelete'))) return;

    try {
      const res = await fetch(`/api/v1/providers/${id}`, {
        method: 'DELETE',
      });
      if (!res.ok) throw new Error('Failed to delete provider');
      refetch();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Delete failed');
    }
  };

  const handleToggle = async (id: string, enabled: boolean) => {
    const provider = configuredProviders.find((item) => item.id === id);
    if (provider?.is_default && !enabled) {
      alert(t('settings.disableDefaultProviderFirst'));
      return;
    }

    setEnabledOverrides((prev) => ({ ...prev, [id]: enabled }));

    try {
      const res = await fetch(`/api/v1/providers/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled }),
      });

      if (!res.ok) {
        setEnabledOverrides((prev) => {
          const next = { ...prev };
          delete next[id];
          return next;
        });
        throw new Error('Failed to update provider');
      }

      setEnabledOverrides((prev) => {
        const next = { ...prev };
        delete next[id];
        return next;
      });
      refetch();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Update failed');
    }
  };

  const handleSetDefault = async (id: string) => {
    try {
      const res = await fetch(`/api/v1/providers/${id}/default`, {
        method: 'POST',
      });
      if (!res.ok) throw new Error('Failed to set default provider');
      refetch();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Set default failed');
    }
  };

  const handleRemoveKeys = async (provider: (typeof configuredProviders)[number]) => {
    if (!window.confirm(t('settings.confirmRemoveKeys'))) return;
    if (provider.is_default) {
      alert(t('settings.removeDefaultProviderFirst'));
      return;
    }

    try {
      const keysRes = await fetch(`/api/v1/providers/${provider.id}/keys`);
      const keysData = await keysRes.json();

      if (keysData.keys && keysData.keys.length > 0) {
        for (const key of keysData.keys) {
          await fetch(`/api/v1/providers/${provider.id}/keys/${key.id}`, {
            method: 'DELETE',
          });
        }
      }
      await fetch(`/api/v1/providers/${provider.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled: false }),
      });
      refetch();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to remove keys');
    }
  };

  const pageHeader = (
    <header className="settings-shell__header">
      <div className="settings-shell__title-block">
        <span className="settings-shell__eyebrow">Workspace controls</span>
        <div>
          <h1 className="settings-shell__title">{t('settings.title')}</h1>
          <p className="settings-shell__description">{t('settings.pageDescription')}</p>
        </div>
      </div>

      <button onClick={onBack} className="settings-shell__back-button">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="m15 18-6-6 6-6" />
        </svg>
        {t('common.back')}
      </button>
    </header>
  );

  if (loading) {
    return (
      <div className="settings-shell">
        {pageHeader}
        <section className="workspace-stage__surface settings-shell__section">
          <div className="animate-pulse space-y-4">
            {[1, 2, 3].map((item) => (
              <div key={item} className="h-28 rounded-[1.5rem] bg-secondary/20" />
            ))}
          </div>
        </section>
      </div>
    );
  }

  if (error) {
    return (
      <div className="settings-shell">
        {pageHeader}
        <section className="workspace-stage__surface settings-shell__section">
          <div className="rounded-[1.25rem] border border-destructive/20 bg-destructive/10 px-4 py-4 text-sm font-medium text-destructive">
            {t('common.error')}: {error}
          </div>
        </section>
      </div>
    );
  }

  return (
    <div className="settings-shell">
      {pageHeader}

      <section className="workspace-stage__surface settings-shell__section">
        <div className="settings-shell__section-head">
          <div>
            <h2 className="settings-shell__section-title">{t('settings.providers')}</h2>
            <p className="settings-shell__section-copy">{t('settings.providersDescription')}</p>
          </div>

          <button onClick={onAddProvider} className="header-link header-link--primary settings-shell__primary-button">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M5 12h14" />
              <path d="M12 5v14" />
            </svg>
            {t('settings.addProvider')}
          </button>
        </div>

        {configuredProviders.length === 0 ? (
          <div className="settings-empty-state">
            <p className="settings-empty-state__title">{t('settings.noProviders')}</p>
            <p className="settings-empty-state__copy">{t('settings.noProvidersHint')}</p>
            <div className="settings-empty-state__chips">
              {presets.slice(0, 5).map((preset) => (
                <span key={preset.type} className="settings-empty-state__chip">
                  {preset.display_name}
                </span>
              ))}
              {presets.length > 5 && (
                <span className="settings-empty-state__more">+{presets.length - 5} more</span>
              )}
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            {configuredProviders.map((provider) => (
              <ProviderCard
                key={provider.id}
                provider={provider}
                onEdit={() => onEditProvider(provider.id)}
                onDelete={() => handleDelete(provider.id)}
                onToggle={() => handleToggle(provider.id, !provider.enabled)}
                onSetDefault={() => handleSetDefault(provider.id)}
                onRemoveKeys={() => handleRemoveKeys(provider)}
              />
            ))}
          </div>
        )}
      </section>

      <DangerZone />
    </div>
  );
}
