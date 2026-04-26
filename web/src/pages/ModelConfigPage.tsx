import { useTranslation } from 'react-i18next';
import {
  ChannelManager,
  ModelSelector,
  RoleMapping,
} from '../components';
// Migration: Using Zustand store adapter instead of ModelConfigContext
import { useProviderStoreAdapter } from '../hooks/useProviderStoreAdapter';

/**
 * Model Configuration Page
 *
 * Demonstrates the unified Channel → Model → Role configuration system.
 * This page provides:
 * - Channel management (add/edit/delete)
 * - Model fetching and selection
 * - Role mapping (Image Generation / Retrieval-Reasoning)
 * 
 * Migration Note: Now uses Zustand store instead of React Context
 */
export function ModelConfigPage() {
  const { t } = useTranslation();
  
  // Migration: Using Zustand store adapter
  const {
    channels,
    role_assignments,
    loading,
    addChannel,
    updateChannel,
    deleteChannel,
    fetchChannelModels,
    assignRole,
    clearRole,
  } = useProviderStoreAdapter();

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-6xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-foreground">
            {t('modelConfig.title', 'Model Configuration')}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t('modelConfig.subtitle', 'Manage channels, models, and role assignments')}
          </p>
        </div>

        {/* Two Column Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Left Column: Channel Management */}
          <div className="space-y-6">
            <section className="bg-card rounded-xl border border-border/30 p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">
                {t('modelConfig.channels', 'Channels')}
              </h2>
              <ChannelManager
                channels={channels}
                onAddChannel={addChannel}
                onUpdateChannel={updateChannel}
                onDeleteChannel={deleteChannel}
                onFetchModels={fetchChannelModels}
                loading={loading}
              />
            </section>

            {/* Model Selector Demo */}
            <section className="bg-card rounded-xl border border-border/30 p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">
                {t('modelConfig.availableModels', 'Available Models')}
              </h2>
              <ModelSelector
                channels={channels}
                onSelect={() => {
                }}
              />
            </section>
          </div>

          {/* Right Column: Role Mapping */}
          <div>
            <section className="bg-card rounded-xl border border-border/30 p-6">
              <RoleMapping
                channels={channels}
                assignments={role_assignments}
                onAssign={assignRole}
                onClear={clearRole}
              />
            </section>
          </div>
        </div>
      </div>
    </div>
  );
}
