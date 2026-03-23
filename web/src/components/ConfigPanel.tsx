import { useState } from 'react';
import { useLanguage } from '../hooks';
import { Provider } from '../hooks/useProviders';

export interface GenerationConfig {
  aspectRatio: '21:9' | '16:9' | '3:2';
  criticRounds: number;
  retrievalMode: 'auto' | 'manual' | 'random' | 'none';
  pipelineMode: 'full' | 'planner-critic' | 'vanilla';
  queryModel?: string;
  genModel?: string;
}

export interface ConfigPanelProps {
  config: GenerationConfig;
  onChange: (config: GenerationConfig) => void;
  providers?: Provider[];
  disabled?: boolean;
  onNavigateToSettings?: () => void;
}

function ChevronIcon({ className }: { className?: string }) {
  return (
    <svg
      className={`w-5 h-5 transition-transform duration-200 ${className}`}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
    </svg>
  );
}

export function ConfigPanel({
  config,
  onChange,
  providers = [],
  disabled = false,
  onNavigateToSettings,
}: ConfigPanelProps) {
  const { t } = useLanguage();
  const [expanded, setExpanded] = useState(false);

  const handleAspectRatioChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onChange({
      ...config,
      aspectRatio: e.target.value as GenerationConfig['aspectRatio'],
    });
  };

  const handleCriticRoundsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = Math.max(1, Math.min(5, parseInt(e.target.value) || 1));
    onChange({
      ...config,
      criticRounds: value,
    });
  };

  const handleRetrievalModeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onChange({
      ...config,
      retrievalMode: e.target.value as GenerationConfig['retrievalMode'],
    });
  };

  const handlePipelineModeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onChange({
      ...config,
      pipelineMode: e.target.value as GenerationConfig['pipelineMode'],
    });
  };

  const handleQueryModelChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onChange({
      ...config,
      queryModel: e.target.value || undefined,
    });
  };

  const handleGenModelChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onChange({
      ...config,
      genModel: e.target.value || undefined,
    });
  };

  // Filter to only configured providers with models
  const configuredProviders = providers.filter(
    (p) => p.enabled && p.status === 'configured' && p.models && p.models.length > 0
  );

  // Check if any models available
  const hasAvailableModels = configuredProviders.some(
    (p) => p.models && p.models.some((m) => m.enabled)
  );

  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        disabled={disabled}
        className="w-full px-4 py-3 flex items-center justify-between bg-background hover:bg-muted/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <span className="text-sm font-medium text-foreground">
          {t('generate.advancedSettings')}
        </span>
        <ChevronIcon className={expanded ? 'rotate-180' : ''} />
      </button>

      {expanded && (
        <div className="p-4 border-t border-border bg-background grid grid-cols-2 gap-4">
          {/* Aspect Ratio */}
          <div>
            <label
              htmlFor="aspect-ratio"
              className="block text-sm font-medium text-foreground mb-2"
            >
              {t('generate.aspectRatio')}
            </label>
            <select
              id="aspect-ratio"
              value={config.aspectRatio}
              onChange={handleAspectRatioChange}
              disabled={disabled}
              className="w-full px-3 py-2 rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
              aria-label={t('generate.aspectRatio')}
            >
              <option value="21:9">21:9</option>
              <option value="16:9">16:9</option>
              <option value="3:2">3:2</option>
            </select>
          </div>

          {/* Critic Rounds */}
          <div>
            <label
              htmlFor="critic-rounds"
              className="block text-sm font-medium text-foreground mb-2"
            >
              {t('generate.criticRounds')}
            </label>
            <input
              id="critic-rounds"
              type="number"
              min={1}
              max={5}
              value={config.criticRounds}
              onChange={handleCriticRoundsChange}
              disabled={disabled}
              className="w-full px-3 py-2 rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
              aria-label={t('generate.criticRounds')}
            />
          </div>

          {/* Retrieval Mode */}
          <div>
            <label
              htmlFor="retrieval-mode"
              className="block text-sm font-medium text-foreground mb-2"
            >
              {t('generate.retrievalMode')}
            </label>
            <select
              id="retrieval-mode"
              value={config.retrievalMode}
              onChange={handleRetrievalModeChange}
              disabled={disabled}
              className="w-full px-3 py-2 rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
              aria-label={t('generate.retrievalMode')}
            >
              <option value="auto">{t('generate.modes.auto')}</option>
              <option value="manual">{t('generate.modes.manual')}</option>
              <option value="random">{t('generate.modes.random')}</option>
              <option value="none">{t('generate.modes.none')}</option>
            </select>
          </div>

          {/* Pipeline Mode */}
          <div>
            <label
              htmlFor="pipeline-mode"
              className="block text-sm font-medium text-foreground mb-2"
            >
              {t('generate.pipelineMode')}
            </label>
            <select
              id="pipeline-mode"
              value={config.pipelineMode}
              onChange={handlePipelineModeChange}
              disabled={disabled}
              className="w-full px-3 py-2 rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
              aria-label={t('generate.pipelineMode')}
            >
              <option value="full">{t('generate.pipelines.full')}</option>
              <option value="planner-critic">{t('generate.pipelines.planner-critic')}</option>
              <option value="vanilla">{t('generate.pipelines.vanilla')}</option>
            </select>
          </div>

          {/* Query Model */}
          <div>
            <label
              htmlFor="query-model"
              className="block text-sm font-medium text-foreground mb-2"
            >
              {t('generate.queryModel')}
            </label>
            {!hasAvailableModels ? (
              <p className="text-sm text-muted-foreground py-2">
                <button
                  type="button"
                  onClick={onNavigateToSettings}
                  className="text-primary hover:underline"
                >
                  {t('generate.configureProviders')}
                </button>
              </p>
            ) : (
              <select
                id="query-model"
                value={config.queryModel || ''}
                onChange={handleQueryModelChange}
                disabled={disabled}
                className="w-full px-3 py-2 rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                aria-label={t('generate.queryModel')}
              >
                <option value="">{t('settings.default')}</option>
                {configuredProviders.map((provider) => {
                  const enabledModels = provider.models?.filter((m) => m.enabled);
                  if (!enabledModels || enabledModels.length === 0) return null;
                  return (
                    <optgroup key={provider.name} label={provider.display_name}>
                      {enabledModels.map((model) => (
                        <option
                          key={`${provider.name}:${model.id}`}
                          value={`${provider.name}:${model.id}`}
                        >
                          {model.name}
                        </option>
                      ))}
                    </optgroup>
                  );
                })}
              </select>
            )}
          </div>

          {/* Generation Model */}
          <div>
            <label
              htmlFor="gen-model"
              className="block text-sm font-medium text-foreground mb-2"
            >
              {t('generate.genModel')}
            </label>
            {!hasAvailableModels ? (
              <p className="text-sm text-muted-foreground py-2">
                <button
                  type="button"
                  onClick={onNavigateToSettings}
                  className="text-primary hover:underline"
                >
                  {t('generate.configureProviders')}
                </button>
              </p>
            ) : (
              <select
                id="gen-model"
                value={config.genModel || ''}
                onChange={handleGenModelChange}
                disabled={disabled}
                className="w-full px-3 py-2 rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                aria-label={t('generate.genModel')}
              >
                <option value="">{t('settings.default')}</option>
                {configuredProviders.map((provider) => {
                  const enabledModels = provider.models?.filter((m) => m.enabled);
                  if (!enabledModels || enabledModels.length === 0) return null;
                  return (
                    <optgroup key={provider.name} label={provider.display_name}>
                      {enabledModels.map((model) => (
                        <option
                          key={`${provider.name}:${model.id}`}
                          value={`${provider.name}:${model.id}`}
                        >
                          {model.name}
                        </option>
                      ))}
                    </optgroup>
                  );
                })}
              </select>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
