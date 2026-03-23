import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { Provider, ProviderPreset } from '../../hooks/useProviders';
import { ModelTagList } from './ModelTagList';

interface ModelTag {
  id: string;
  name: string;
}

interface ProviderFormProps {
  provider?: Provider;
  presets: ProviderPreset[];
  onSave: (data: ProviderFormData) => Promise<void>;
  onCancel: () => void;
}

export interface ProviderFormData {
  type: string;
  name: string;
  display_name: string;
  base_url: string;
  api_key: string;
  models: { id: string; name: string; enabled: boolean }[];
}

interface RemoteModel {
  id: string;
  name?: string;
}

export function ProviderForm({ provider, presets, onSave, onCancel }: ProviderFormProps) {
  const { t } = useTranslation();
  const isEditMode = !!provider;

  const [formData, setFormData] = useState<ProviderFormData>({
    type: provider?.type || '',
    name: provider?.name || '',
    display_name: provider?.display_name || '',
    base_url: provider?.base_url || '',
    api_key: '',
    models: [],
  });
  const [models, setModels] = useState<ModelTag[]>([]);
  const [showApiKey, setShowApiKey] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ valid: boolean; message: string } | null>(null);
  const [saving, setSaving] = useState(false);
  const [errors, setErrors] = useState<Partial<Record<keyof ProviderFormData | 'models', string>>>({});
  const [remoteModels, setRemoteModels] = useState<RemoteModel[]>([]);
  const [saveError, setSaveError] = useState<string | null>(null);

  useEffect(() => {
    if (provider) {
      if (provider.models && provider.models.length > 0) {
        setModels(provider.models.map((model) => ({ id: model.id, name: model.name })));
      } else {
        const migrated: ModelTag[] = [];
        if (provider.query_model) {
          migrated.push({ id: provider.query_model, name: provider.query_model });
        }
        if (provider.gen_model && provider.gen_model !== provider.query_model) {
          migrated.push({ id: provider.gen_model, name: provider.gen_model });
        }
        setModels(migrated);
      }
    }
  }, [provider]);

  useEffect(() => {
    if (!isEditMode && formData.type) {
      const preset = presets.find((item) => item.type === formData.type);
      if (preset) {
        setFormData((prev) => ({
          ...prev,
          name: prev.name || formData.type,
          display_name: prev.display_name || preset.display_name,
          base_url: prev.base_url || preset.api_host,
        }));
        if (preset.default_models && preset.default_models.length > 0) {
          setModels(preset.default_models.map((model) => ({ id: model.id, name: model.name })));
        }
      }
    }
  }, [formData.type, presets, isEditMode]);

  const validate = () => {
    const newErrors: Partial<Record<keyof ProviderFormData | 'models', string>> = {};

    if (!formData.type && !isEditMode) {
      newErrors.type = t('validation.providerRequired');
    }
    if (!isEditMode && !formData.api_key) {
      newErrors.api_key = t('validation.apiKeyRequired');
    }
    if (isEditMode && provider?.status === 'no_keys' && !formData.api_key) {
      newErrors.api_key = t('validation.apiKeyRequired');
    }
    if (models.length === 0) {
      newErrors.models = t('validation.modelRequired');
    }
    if (formData.base_url && !/^https?:\/\/.+/.test(formData.base_url)) {
      newErrors.base_url = t('validation.invalidUrl');
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleTest = async () => {
    if (!formData.api_key) return;

    setTesting(true);
    setTestResult(null);
    setRemoteModels([]);

    try {
      const res = await fetch('/api/v1/providers/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          provider: formData.type || 'custom',
          api_key: formData.api_key,
          base_url: formData.base_url,
        }),
      });
      const data = await res.json();

      if (data.valid) {
        setTestResult({
          valid: true,
          message: t('settings.connectionSuccess', { count: data.models_available || 0 }),
        });
        if (data.models && data.models.length > 0) {
          setRemoteModels(
            data.models.map((model: RemoteModel) => ({
              id: model.id,
              name: model.name || model.id,
            })),
          );
        }
      } else {
        setTestResult({
          valid: false,
          message: data.errors?.[0]?.message || t('settings.connectionFailed'),
        });
      }
    } catch {
      setTestResult({ valid: false, message: t('settings.connectionFailed') });
    } finally {
      setTesting(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    setSaving(true);
    setSaveError(null);
    try {
      await onSave({
        ...formData,
        models: models.map((model) => ({ id: model.id, name: model.name, enabled: true })),
      });
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  const updateField = (field: keyof ProviderFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
    }
  };

  const domesticPresets = presets.filter((preset) =>
    ['deepseek', 'zhipu', 'moonshot', 'qwen', 'doubao', 'baichuan', 'minimax', 'yi', 'hunyuan', 'stepfun', 'silicon'].includes(preset.type),
  );
  const internationalPresets = presets.filter((preset) =>
    ['openai', 'anthropic', 'gemini', 'mistral', 'grok', 'perplexity'].includes(preset.type),
  );
  const otherPresets = presets.filter(
    (preset) =>
      !domesticPresets.some((domestic) => domestic.type === preset.type) &&
      !internationalPresets.some((international) => international.type === preset.type),
  );

  const selectedPreset = presets.find((preset) => preset.type === formData.type);

  return (
    <form onSubmit={handleSubmit} className="provider-form">
      <div className="provider-form__grid">
        <section className="provider-form__card">
          <div className="provider-form__section-head">
            <div>
              <h2 className="provider-form__section-title">{t('settings.providerIdentity')}</h2>
              <p className="provider-form__section-copy">{t('settings.providerIdentityHint')}</p>
            </div>
          </div>

          <div className="provider-form__field-grid">
            <div className="provider-form__field">
              <label className="provider-form__label">
                {t('settings.providerType')} <span className="provider-form__required">*</span>
              </label>
              <select
                value={formData.type}
                onChange={(e) => updateField('type', e.target.value)}
                className="provider-form__control"
                required
                disabled={isEditMode}
              >
                <option value="">{t('common.select')}</option>
                <optgroup label={t('settings.domesticProviders')}>
                  {domesticPresets.map((preset) => (
                    <option key={preset.type} value={preset.type}>
                      {preset.display_name}
                    </option>
                  ))}
                </optgroup>
                <optgroup label={t('settings.internationalProviders')}>
                  {internationalPresets.map((preset) => (
                    <option key={preset.type} value={preset.type}>
                      {preset.display_name}
                    </option>
                  ))}
                </optgroup>
                <optgroup label={t('settings.otherProviders')}>
                  {otherPresets.map((preset) => (
                    <option key={preset.type} value={preset.type}>
                      {preset.display_name}
                    </option>
                  ))}
                  <option value="custom">{t('settings.customProvider')}</option>
                </optgroup>
              </select>
              {errors.type && <p className="provider-form__error">{errors.type}</p>}
              {selectedPreset?.docs_url && (
                <a
                  href={selectedPreset.docs_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="provider-form__link"
                >
                  {t('settings.getApiKey')} {'->'}
                </a>
              )}
            </div>

            <div className="provider-form__field">
              <label className="provider-form__label">{t('settings.displayName')}</label>
              <input
                type="text"
                value={formData.display_name}
                onChange={(e) => updateField('display_name', e.target.value)}
                className="provider-form__control"
                placeholder={t('settings.displayNamePlaceholder')}
              />
            </div>

            <div className="provider-form__field provider-form__field--full">
              <label className="provider-form__label">{t('settings.baseUrl')}</label>
              <input
                type="url"
                value={formData.base_url}
                onChange={(e) => updateField('base_url', e.target.value)}
                className="provider-form__control"
                placeholder="https://api.example.com/v1"
              />
              {errors.base_url && <p className="provider-form__error">{errors.base_url}</p>}
            </div>
          </div>
        </section>

        <section className="provider-form__card">
          <div className="provider-form__section-head">
            <div>
              <h2 className="provider-form__section-title">{t('settings.credentials')}</h2>
              <p className="provider-form__section-copy">{t('settings.credentialsHint')}</p>
            </div>
            {formData.api_key && (
              <button
                type="button"
                onClick={handleTest}
                disabled={testing}
                className="provider-form__secondary-button"
              >
                {testing ? t('common.testing') : t('settings.testConnection')}
              </button>
            )}
          </div>

          <div className="provider-form__field">
            <label className="provider-form__label">
              {t('settings.apiKey')} {(!isEditMode || provider?.status === 'no_keys') && <span className="provider-form__required">*</span>}
              {isEditMode && provider?.status !== 'no_keys' && (
                <span className="provider-form__hint-inline">({t('settings.apiKeyOptional')})</span>
              )}
            </label>

            <div className="provider-form__password-wrap">
              <input
                type={showApiKey ? 'text' : 'password'}
                value={formData.api_key}
                onChange={(e) => updateField('api_key', e.target.value)}
                className={`provider-form__control provider-form__control--password ${errors.api_key ? 'provider-form__control--error' : ''}`}
                placeholder="sk-..."
              />
              <button
                type="button"
                onClick={() => setShowApiKey(!showApiKey)}
                className="provider-form__toggle-visibility"
              >
                {showApiKey ? t('common.hide') : t('common.show')}
              </button>
            </div>
            {errors.api_key && <p className="provider-form__error">{errors.api_key}</p>}
          </div>

          {testResult && (
            <div className={`provider-form__test-result ${testResult.valid ? 'is-success' : 'is-error'}`}>
              {testResult.message}
            </div>
          )}
        </section>
      </div>

      <section className="provider-form__card provider-form__card--models">
        <div className="provider-form__section-head">
          <div>
            <h2 className="provider-form__section-title">{t('settings.models')}</h2>
            <p className="provider-form__section-copy">{t('settings.modelsHint')}</p>
          </div>
        </div>

        <ModelTagList
          models={models}
          onChange={setModels}
          remoteModels={testResult?.valid ? remoteModels.map((model) => ({ id: model.id, name: model.name || model.id })) : undefined}
          loading={testing}
        />
        {errors.models && <p className="provider-form__error">{errors.models}</p>}
      </section>

      {saveError && <div className="provider-form__save-error">{saveError}</div>}

      <div className="provider-form__actions">
        <button type="submit" disabled={saving} className="provider-form__primary-button">
          {saving ? t('common.saving') : t('common.save')}
        </button>
        <button type="button" onClick={onCancel} className="provider-form__ghost-button">
          {t('common.cancel')}
        </button>
      </div>
    </form>
  );
}
