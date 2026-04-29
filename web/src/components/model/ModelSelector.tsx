import { useState, useMemo } from 'react';
import type { Provider } from '../../stores';

export interface ModelSelectorProps {
  channels?: Provider[];
  selectedModelId?: string;
  onModelSelect?: (modelId: string) => void;
  onSelect?: (channelId: string, modelId: string) => void;
  placeholder?: string;
  className?: string;
}

export interface CompactModelSelectorProps extends ModelSelectorProps {
  showProvider?: boolean;
}

export function ModelSelector({
  channels = [],
  selectedModelId,
  onModelSelect,
  onSelect,
  placeholder = 'Select a model',
  className = ''
}: ModelSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);

  // Flatten all models from all channels
  const allModels = useMemo(() => {
    const models: Array<{ id: string; name: string; provider: string }> = [];
    for (const channel of channels) {
      if (!channel.enabled) continue;
      for (const model of channel.models) {
        if (model.enabled) {
          models.push({
            id: model.id,
            name: model.name,
            provider: channel.display_name || channel.name
          });
        }
      }
    }
    return models;
  }, [channels]);

  const selectedModel = allModels.find(m => m.id === selectedModelId);

  const handleSelect = (modelId: string) => {
    // Find which channel this model belongs to
    for (const channel of channels) {
      const model = channel.models.find(m => m.id === modelId);
      if (model) {
        onSelect?.(channel.id, modelId);
        break;
      }
    }
    onModelSelect?.(modelId);
    setIsOpen(false);
  };

  return (
    <div className={`relative ${className}`}>
      <button
        className="w-full flex items-center justify-between px-3 py-2 text-sm bg-background border border-border rounded-lg hover:border-primary/50 transition-colors"
        onClick={() => setIsOpen(!isOpen)}
      >
        <span className={selectedModel ? 'text-foreground' : 'text-muted-foreground'}>
          {selectedModel?.name || placeholder}
        </span>
        <svg className={`w-4 h-4 text-muted-foreground transition-transform ${isOpen ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-card border border-border rounded-lg shadow-lg max-h-60 overflow-auto">
          {allModels.map((model) => (
            <button
              key={model.id}
              className={`w-full text-left px-3 py-2 text-sm hover:bg-muted/60 transition-colors flex items-center justify-between ${selectedModelId === model.id ? 'bg-primary/10 text-primary' : 'text-foreground'}`}
              onClick={() => handleSelect(model.id)}
            >
              <span>{model.name}</span>
              {model.provider && <span className="text-xs text-muted-foreground">{model.provider}</span>}
            </button>
          ))}
          {allModels.length === 0 && (
            <div className="px-3 py-4 text-center text-sm text-muted-foreground">
              <p>No models available</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export function CompactModelSelector({
  channels = [],
  selectedModelId,
  onModelSelect,
  onSelect,
  showProvider = true,
  placeholder = 'Select model',
  className = ''
}: CompactModelSelectorProps) {
  // Flatten all models from all channels
  const allModels = useMemo(() => {
    const models: Array<{ id: string; name: string; provider: string }> = [];
    for (const channel of channels) {
      if (!channel.enabled) continue;
      for (const model of channel.models) {
        if (model.enabled) {
          models.push({
            id: model.id,
            name: model.name,
            provider: channel.display_name || channel.name
          });
        }
      }
    }
    return models;
  }, [channels]);

  const handleChange = (modelId: string) => {
    // Find which channel this model belongs to
    for (const channel of channels) {
      const model = channel.models.find(m => m.id === modelId);
      if (model) {
        onSelect?.(channel.id, modelId);
        break;
      }
    }
    onModelSelect?.(modelId);
  };

  return (
    <div className={className}>
      <select
        value={selectedModelId || ''}
        onChange={(e) => handleChange(e.target.value)}
        className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
      >
        <option value="" disabled>{placeholder}</option>
        {allModels.map((model) => (
          <option key={model.id} value={model.id}>
            {model.name}{showProvider && model.provider ? ` (${model.provider})` : ''}
          </option>
        ))}
      </select>
    </div>
  );
}
