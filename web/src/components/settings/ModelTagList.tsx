import { useState } from 'react';
import { useTranslation } from 'react-i18next';

interface ModelTag {
  id: string;
  name: string;
}

interface ModelTagListProps {
  models: ModelTag[];
  onChange: (models: ModelTag[]) => void;
  remoteModels?: ModelTag[]; // Models from API
  loading?: boolean;
}

export function ModelTagList({ models, onChange, remoteModels, loading }: ModelTagListProps) {
  const { t } = useTranslation();
  const [showAdd, setShowAdd] = useState(false);
  const [customInput, setCustomInput] = useState('');

  const handleAdd = (modelId: string, modelName: string) => {
    if (!modelId.trim()) return;
    if (models.some(m => m.id === modelId)) return;

    onChange([...models, { id: modelId, name: modelName || modelId }]);
    setShowAdd(false);
    setCustomInput('');
  };

  const handleRemove = (modelId: string) => {
    onChange(models.filter(m => m.id !== modelId));
  };

  const availableModels = remoteModels?.filter(
    rm => !models.some(m => m.id === rm.id)
  ) || [];

  return (
    <div className="model-tag-list">
      <label className="block text-sm font-medium mb-2">
        {t('settings.models')}
      </label>

      {/* Tag list */}
      <div className="flex flex-wrap gap-2 mb-2">
        {models.map(model => (
          <div
            key={model.id}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-gray-100 rounded-md text-sm
                       hover:bg-gray-200 transition-colors duration-200 group"
          >
            <span className="font-mono text-gray-700">{model.id}</span>
            <button
              type="button"
              onClick={() => handleRemove(model.id)}
              className="w-5 h-5 flex items-center justify-center rounded-full
                         text-gray-400 hover:text-red-500 hover:bg-red-50
                         opacity-0 group-hover:opacity-100 transition-all duration-200"
              aria-label={t('common.remove')}
            >
              ×
            </button>
          </div>
        ))}

        {/* Add button */}
        <button
          type="button"
          onClick={() => setShowAdd(!showAdd)}
          disabled={loading}
          className="inline-flex items-center gap-1 px-3 py-1.5
                     border-2 border-dashed border-gray-300 rounded-md text-sm
                     text-gray-500 hover:border-blue-400 hover:text-blue-500
                     transition-colors duration-200 min-w-[44px] min-h-[36px]"
        >
          <span className="text-lg leading-none">+</span>
        </button>
      </div>

      {/* Add dropdown/input */}
      {showAdd && (
        <div className="p-3 bg-gray-50 rounded-lg border border-gray-200 animate-in fade-in duration-200">
          {availableModels.length > 0 ? (
            <div className="space-y-2">
              <p className="text-xs text-gray-500">{t('settings.selectFromApi')}</p>
              <div className="flex flex-wrap gap-1.5">
                {availableModels.slice(0, 7).map(model => (
                  <button
                    key={model.id}
                    type="button"
                    onClick={() => handleAdd(model.id, model.name)}
                    className="px-2.5 py-1 text-xs bg-white border border-gray-200 rounded
                               hover:border-blue-400 hover:text-blue-600 transition-colors"
                  >
                    {model.id}
                  </button>
                ))}
              </div>
            </div>
          ) : null}

          <div className="mt-2 pt-2 border-t border-gray-200">
            <p className="text-xs text-gray-500 mb-1.5">{t('settings.enterManually')}</p>
            <div className="flex gap-2">
              <input
                type="text"
                value={customInput}
                onChange={e => setCustomInput(e.target.value)}
                onKeyDown={e => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    handleAdd(customInput, customInput);
                  }
                }}
                placeholder="model-id"
                className="flex-1 px-2.5 py-1.5 text-sm border rounded
                           focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
              <button
                type="button"
                onClick={() => handleAdd(customInput, customInput)}
                disabled={!customInput.trim()}
                className="px-3 py-1.5 text-sm bg-blue-500 text-white rounded
                           hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed
                           transition-colors"
              >
                {t('common.add')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Hint */}
      <p className="text-xs text-gray-400 mt-1">
        {t('settings.modelsHint')}
      </p>
    </div>
  );
}
