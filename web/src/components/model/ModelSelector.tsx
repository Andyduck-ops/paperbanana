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
    <div className={`model-selector ${className}`}>
      <button
        className="model-selector__trigger"
        onClick={() => setIsOpen(!isOpen)}
      >
        {selectedModel?.name || placeholder}
      </button>
      {isOpen && (
        <div className="model-selector__dropdown">
          {allModels.map((model) => (
            <button
              key={model.id}
              className={`model-selector__option ${selectedModelId === model.id ? 'selected' : ''}`}
              onClick={() => handleSelect(model.id)}
            >
              {model.name}
              {model.provider && <span className="model-selector__provider">{model.provider}</span>}
            </button>
          ))}
          {allModels.length === 0 && (
            <div className="model-selector__empty">
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
    <div className={`compact-model-selector ${className}`}>
      <select
        value={selectedModelId || ''}
        onChange={(e) => handleChange(e.target.value)}
        className="compact-model-selector__select"
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
