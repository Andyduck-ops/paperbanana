import { useState } from 'react';
import type { Provider, ModelInfo } from '../../hooks/useProviders';

interface ProviderManagerProps {
  provider: Provider;
  onToggleProvider: (enabled: boolean) => void;
  onToggleModel: (modelId: string, enabled: boolean) => void;
  onConfigureKey: () => void;
}

export function ProviderManager({
  provider,
  onToggleProvider,
  onToggleModel,
  onConfigureKey,
}: ProviderManagerProps) {
  const [expanded, setExpanded] = useState(false);

  const statusIcon = () => {
    switch (provider.status) {
      case 'configured':
        return <span className="text-green-500 text-xs">✓</span>;
      case 'no_keys':
        return <span className="text-orange-500 text-xs">⚠</span>;
      case 'invalid':
        return <span className="text-red-500 text-xs">✗</span>;
      default:
        return null;
    }
  };

  return (
    <div className="border-b border-border/10 last:border-b-0">
      {/* Provider Header */}
      <div className="px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            checked={provider.enabled}
            onChange={(e) => onToggleProvider(e.target.checked)}
            className="w-4 h-4 rounded border-border text-primary focus:ring-primary cursor-pointer"
          />
          <span className="text-sm font-medium text-foreground">
            {provider.display_name || provider.name}
          </span>
          {statusIcon()}
        </div>

        <div className="flex items-center gap-2">
          {provider.status !== 'configured' && (
            <button
              onClick={onConfigureKey}
              className="text-xs px-2 py-1 rounded bg-primary/10 text-primary hover:bg-primary/20 transition-colors"
            >
              🔑 Key
            </button>
          )}
          {provider.models && provider.models.length > 0 && (
            <button
              onClick={() => setExpanded(!expanded)}
              className="w-6 h-6 rounded flex items-center justify-center hover:bg-muted/50 transition-colors"
            >
              <span className="text-xs text-muted-foreground">
                {expanded ? '▲' : '▼'}
              </span>
            </button>
          )}
        </div>
      </div>

      {/* Models List */}
      {expanded && provider.models && (
        <div className="px-4 pb-3 pl-10 space-y-1 animate-in slide-in-from-top-1 duration-200">
          {provider.models.map((model) => (
            <ModelItem
              key={model.id}
              model={model}
              onToggle={(enabled) => onToggleModel(model.id, enabled)}
            />
          ))}
        </div>
      )}
    </div>
  );
}

interface ModelItemProps {
  model: ModelInfo;
  onToggle: (enabled: boolean) => void;
}

function ModelItem({ model, onToggle }: ModelItemProps) {
  return (
    <div className="flex items-center justify-between py-1.5 group">
      <div className="flex items-center gap-2">
        <span className="text-xs text-muted-foreground">{model.name}</span>
        {model.supports_vision && (
          <span className="text-[10px] px-1.5 py-0.5 rounded bg-purple-500/10 text-purple-500">
            Vision
          </span>
        )}
      </div>
      <button
        onClick={() => onToggle(!model.enabled)}
        className={`
          w-6 h-6 rounded-md flex items-center justify-center text-sm font-medium
          transition-all duration-150
          ${model.enabled
            ? 'bg-primary/15 text-primary hover:bg-primary/25'
            : 'bg-muted/40 text-muted-foreground hover:bg-muted/60 group-hover:bg-muted/50'
          }
        `}
        title={model.enabled ? 'Remove model' : 'Add model'}
      >
        {model.enabled ? '−' : '+'}
      </button>
    </div>
  );
}
