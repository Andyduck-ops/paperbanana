import { useTranslation } from 'react-i18next';
import { useProvider, usePresets } from '../hooks/useProviders';
import { ProviderForm, type ProviderFormData } from '../components/settings/ProviderForm';

interface ProviderEditPageProps {
  providerId?: string;
  isNew: boolean;
  onBack: () => void;
}

export function ProviderEditPage({ providerId, isNew, onBack }: ProviderEditPageProps) {
  const { t } = useTranslation();
  const { provider, loading, error } = useProvider(isNew ? '' : providerId || '');
  const { presets } = usePresets();

  const handleSave = async (data: ProviderFormData) => {
    const primaryModel = data.models.find((model) => model.enabled)?.id || data.models[0]?.id;

    if (isNew) {
      const listRes = await fetch('/api/v1/providers');
      const listData = await listRes.json();
      const existingProvider = listData.providers?.find(
        (item: { id: string; name: string; type: string; is_system: boolean }) =>
          item.is_system ? item.type === data.type : item.name === data.name,
      );

      let targetId: string;

      if (existingProvider) {
        targetId = existingProvider.id;
        const updateRes = await fetch(`/api/v1/providers/${targetId}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            display_name: data.display_name,
            api_host: data.base_url,
            models: data.models,
            query_model: primaryModel,
            gen_model: primaryModel,
            enabled: existingProvider.enabled,
          }),
        });

        if (!updateRes.ok) {
          const errData = await updateRes.json().catch(() => ({}));
          throw new Error(errData.error || 'Failed to update provider');
        }
      } else {
        const createRes = await fetch('/api/v1/providers', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            type: data.type,
            name: data.name,
            display_name: data.display_name,
            api_host: data.base_url,
            models: data.models,
            query_model: primaryModel,
            gen_model: primaryModel,
            enabled: true,
          }),
        });

        if (!createRes.ok) {
          const errData = await createRes.json().catch(() => ({}));
          throw new Error(errData.error || 'Failed to create provider');
        }

        const createData = await createRes.json();
        targetId = createData.provider.id;
      }

      if (data.api_key) {
        const keyRes = await fetch(`/api/v1/providers/${targetId}/keys`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ api_key: data.api_key }),
        });

        if (!keyRes.ok) {
          const errData = await keyRes.json().catch(() => ({}));
          throw new Error(errData.error || 'Failed to add API key');
        }
      }

      onBack();
      return;
    }

    const updateRes = await fetch(`/api/v1/providers/${providerId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        display_name: data.display_name,
        api_host: data.base_url,
        models: data.models,
        query_model: primaryModel,
        gen_model: primaryModel,
        enabled: provider?.enabled ?? true,
      }),
    });

    if (!updateRes.ok) {
      const errData = await updateRes.json().catch(() => ({}));
      throw new Error(errData.error || 'Failed to update provider');
    }

    if (data.api_key) {
      const keyRes = await fetch(`/api/v1/providers/${providerId}/keys`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ api_key: data.api_key }),
      });

      if (!keyRes.ok) {
        const errData = await keyRes.json().catch(() => ({}));
        throw new Error(errData.error || 'Failed to add API key');
      }
    }

    onBack();
  };

  const pageTitle = isNew ? t('settings.addProvider') : t('settings.editProvider');
  const pageDescription = isNew
    ? t('settings.providerCreateDescription')
    : t('settings.providerEditDescription');

  if (!isNew && loading) {
    return (
      <div className="settings-shell settings-shell--compact">
        <header className="settings-shell__header settings-shell__header--stacked">
          <div className="settings-shell__title-block">
            <span className="settings-shell__eyebrow">Provider setup</span>
            <div>
              <h1 className="settings-shell__title">{pageTitle}</h1>
              <p className="settings-shell__description">{pageDescription}</p>
            </div>
          </div>
        </header>

        <section className="workspace-stage__surface settings-shell__section">
          <div className="animate-pulse space-y-4">
            {[1, 2, 3, 4, 5].map((item) => (
              <div key={item} className="h-16 rounded-[1.25rem] bg-secondary/20" />
            ))}
          </div>
        </section>
      </div>
    );
  }

  if (!isNew && error) {
    return (
      <div className="settings-shell settings-shell--compact">
        <header className="settings-shell__header settings-shell__header--stacked">
          <div className="settings-shell__title-block">
            <span className="settings-shell__eyebrow">Provider setup</span>
            <div>
              <h1 className="settings-shell__title">{pageTitle}</h1>
              <p className="settings-shell__description">{pageDescription}</p>
            </div>
          </div>
          <button onClick={onBack} className="settings-shell__back-button">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="m15 18-6-6 6-6" />
            </svg>
            {t('common.back')}
          </button>
        </header>

        <section className="workspace-stage__surface settings-shell__section">
          <div className="rounded-[1.25rem] border border-destructive/20 bg-destructive/10 px-4 py-4 text-sm font-medium text-destructive">
            {t('common.error')}: {error}
          </div>
        </section>
      </div>
    );
  }

  return (
    <div className="settings-shell settings-shell--compact">
      <header className="settings-shell__header settings-shell__header--stacked">
        <div className="settings-shell__title-block">
          <span className="settings-shell__eyebrow">Provider setup</span>
          <div>
            <h1 className="settings-shell__title">{pageTitle}</h1>
            <p className="settings-shell__description">{pageDescription}</p>
          </div>
        </div>

        <button onClick={onBack} className="settings-shell__back-button">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="m15 18-6-6 6-6" />
          </svg>
          {t('common.back')}
        </button>
      </header>

      <section className="workspace-stage__surface settings-shell__section">
        <ProviderForm
          provider={isNew ? undefined : provider || undefined}
          presets={presets}
          onSave={handleSave}
          onCancel={onBack}
        />
      </section>
    </div>
  );
}
