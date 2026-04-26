import { useTranslation } from 'react-i18next';
import {
  ChannelManager,
  DarkModeToggle,
  ModelSelector,
  RoleMapping,
  TemplateManager,
} from '../components';
import { useTemplates } from '../hooks';
// Migration: Using Zustand store adapter instead of ModelConfigContext
import { useProviderStoreAdapter } from '../hooks/useProviderStoreAdapter';

interface SettingsPageProps {
  onBack: () => void;
  onAddProvider?: () => void;
  onEditProvider?: (id: string) => void;
  variant?: 'page' | 'drawer';
}

export function SettingsPage({ onBack, variant = 'page' }: SettingsPageProps) {
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
  
  const {
    templates,
    isLoading: templatesLoading,
    createTemplate,
    updateTemplate,
    deleteTemplate,
  } = useTemplates();

  const pinnedChannelIds = new Set(
    Object.values(role_assignments)
      .map((assignment) => assignment?.provider_id)
      .filter((value): value is string => Boolean(value))
  );

  const visibleChannels = channels.filter((channel) =>
    channel.status === 'configured' ||
    !channel.is_system ||
    pinnedChannelIds.has(channel.id)
  );

  const configuredChannels = visibleChannels.filter(
    (channel) => channel.enabled && channel.status === 'configured'
  );

  const hasVisibleChannels = visibleChannels.length > 0;
  const hasAvailableModels = configuredChannels.some((channel) =>
    channel.models.some((model) => model.enabled)
  );
  const assignedRolesCount = Object.values(role_assignments).filter(Boolean).length;

  return (
    <div className={`settings-shell ${variant === 'drawer' ? 'settings-shell--drawer' : ''}`}>
      <header className="settings-shell__header">
        <div className="settings-shell__title-block">
          <div>
            <h1 className="settings-shell__title">{t('settings.title')}</h1>
            <p className="settings-shell__description">
              {t('settings.pageDescription')}
            </p>
          </div>
        </div>

        <button onClick={onBack} className="settings-shell__back-button" type="button">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="m15 18-6-6 6-6" />
          </svg>
          {t('common.back')}
        </button>
      </header>

      <div className="settings-main-grid">
        <div className="settings-column">
          <section className="workspace-stage__surface settings-shell__section">
            <div className="settings-shell__section-head">
              <div>
                <h2 className="settings-shell__section-title">
                  {t('modelConfig.channels', 'Channels')}
                </h2>
                <p className="settings-shell__section-copy">
                  {t('modelConfig.channelsDescription', 'Only connected channels are shown here; built-in presets stay hidden until configured.')}
                </p>
              </div>
            </div>

            {!hasVisibleChannels && (
              <div className="settings-empty-state mb-4">
                <p className="settings-empty-state__title">
                  {t('modelConfig.noChannels', 'No channels configured')}
                </p>
                <p className="settings-empty-state__copy">
                  {t('modelConfig.addChannelHint', 'Add a channel to start using models')}
                </p>
              </div>
            )}

            <ChannelManager
              channels={visibleChannels}
              onAddChannel={addChannel}
              onUpdateChannel={updateChannel}
              onDeleteChannel={deleteChannel}
              onFetchModels={fetchChannelModels}
              loading={loading}
            />
          </section>

          <section className="workspace-stage__surface settings-shell__section">
            <div className="settings-shell__section-head">
              <div>
                <h2 className="settings-shell__section-title">
                  {t('modelConfig.availableModels', 'Available Models')}
                </h2>
                <p className="settings-shell__section-copy">
                  {t('modelConfig.modelsDescription', 'Browse models from channels you have already connected.')}
                </p>
              </div>
            </div>

            {hasAvailableModels ? (
              <ModelSelector
                channels={configuredChannels}
              />
            ) : (
              <div className="settings-empty-state">
                <p className="settings-empty-state__title">
                  {t('modelConfig.noModelsAvailable', 'No models available')}
                </p>
                <p className="settings-empty-state__copy">
                  {hasVisibleChannels
                    ? t('modelConfig.fetchModelsHint', 'Fetch models from a channel to get started')
                    : t('modelConfig.hiddenSystemChannelsHint', 'Add and configure a channel before models appear here.')}
                </p>
              </div>
            )}
          </section>
        </div>

        <div className="settings-column">
          <section className="workspace-stage__surface settings-shell__section">
            <div className="settings-shell__section-head">
              <div>
                <h2 className="settings-shell__section-title">
                  {t('modelConfig.roleMapping', 'Role Mapping')}
                </h2>
                <p className="settings-shell__section-copy">
                  {t('modelConfig.roleMappingDescription', 'Assign roles only from enabled, configured models')}
                </p>
              </div>
            </div>

            {hasAvailableModels ? (
              <RoleMapping
                channels={configuredChannels}
                assignments={role_assignments}
                onAssign={assignRole}
                onClear={clearRole}
              />
            ) : (
              <div className="settings-empty-state">
                <p className="settings-empty-state__title">
                  {t('modelConfig.roleMapping', 'Role Mapping')}
                </p>
                <p className="settings-empty-state__copy">
                  {t('modelConfig.roleMappingEmptyHint', 'Configure a channel first, then assign retrieval and generation roles here.')}
                </p>
              </div>
            )}
          </section>

          {/* Appearance Settings */}
          <section className="workspace-stage__surface settings-shell__section">
            <div className="settings-shell__section-head">
              <div>
                <h2 className="settings-shell__section-title">
                  {t('settings.appearance', 'Appearance')}
                </h2>
                <p className="settings-shell__section-copy">
                  {t('settings.appearanceDescription', 'Customize the look and feel of the application.')}
                </p>
              </div>
            </div>

            {/* Color scheme tri-state: Light / Dark / Auto (system) */}
            <div className="space-y-3">
              <span className="block text-sm font-medium text-foreground">
                {t('settings.colorScheme', 'Color scheme')}
              </span>
              <DarkModeToggle />
              <p className="text-xs text-muted-foreground">
                {t(
                  'settings.colorSchemeDescription',
                  'Light follows the Claude anchor, Dark follows the Linear anchor. Auto matches your OS preference.'
                )}
              </p>
            </div>
          </section>

          <section className="workspace-stage__surface settings-shell__section">
            <h2 className="settings-shell__section-title text-base">
              {t('modelConfig.configurationSummary', 'Configuration Summary')}
            </h2>
            <div className="mt-4 space-y-3">
              <div className="flex items-center justify-between border-b border-border/30 py-2">
                <span className="text-muted-foreground">{t('modelConfig.totalChannels', 'Configured Channels')}</span>
                <span className="font-medium">{visibleChannels.length}</span>
              </div>
              <div className="flex items-center justify-between border-b border-border/30 py-2">
                <span className="text-muted-foreground">{t('modelConfig.imageGenAssigned', 'Image Gen Assigned')}</span>
                <span className="font-medium">
                  {role_assignments.image_generation ? t('modelConfig.assigned', 'Assigned') : '—'}
                </span>
              </div>
              <div className="flex items-center justify-between py-2">
                <span className="text-muted-foreground">{t('modelConfig.totalAssignments', 'Assigned Roles')}</span>
                <span className="font-medium">{assignedRolesCount} / 2</span>
              </div>
            </div>
          </section>
        </div>
      </div>

      {/* Template Management Section */}
      <div className="settings-main-grid">
        <div className="settings-column" style={{ gridColumn: '1 / -1' }}>
          <section className="workspace-stage__surface settings-shell__section">
            <div className="settings-shell__section-head">
              <div>
                <h2 className="settings-shell__section-title">
                  {t('template.title', 'Template Management')}
                </h2>
                <p className="settings-shell__section-copy">
                  {t('template.description', 'Manage reusable prompt templates for common figure types.')}
                </p>
              </div>
            </div>
            <TemplateManager
              templates={templates}
              isLoading={templatesLoading}
              onCreate={createTemplate}
              onUpdate={updateTemplate}
              onDelete={deleteTemplate}
            />
          </section>
        </div>
      </div>
    </div>
  );
}
