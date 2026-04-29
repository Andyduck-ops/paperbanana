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

  const isDrawer = variant === 'drawer';

  return (
    <div className={`${isDrawer ? 'h-full overflow-auto' : 'w-full max-w-6xl mx-auto px-4 py-6'}`}>
      <header className="flex items-start justify-between gap-4 mb-6">
        <div className="flex-1 min-w-0">
          <h1 className="text-2xl font-semibold text-foreground">{t('settings.title')}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t('settings.pageDescription')}
          </p>
        </div>

        <button
          onClick={onBack}
          className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
          type="button"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="m15 18-6-6 6-6" />
          </svg>
          {t('common.back')}
        </button>
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="space-y-6">
          <section className="bg-card rounded-xl border border-border/30 p-6">
            <div className="mb-4">
              <h2 className="text-lg font-semibold text-foreground">
                {t('modelConfig.channels', 'Channels')}
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                {t('modelConfig.channelsDescription', 'Only connected channels are shown here; built-in presets stay hidden until configured.')}
              </p>
            </div>

            {!hasVisibleChannels && (
              <div className="mb-4 p-4 rounded-lg bg-muted/50 text-center">
                <p className="text-sm font-medium text-foreground">
                  {t('modelConfig.noChannels', 'No channels configured')}
                </p>
                <p className="text-xs text-muted-foreground mt-1">
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

          <section className="bg-card rounded-xl border border-border/30 p-6">
            <div className="mb-4">
              <h2 className="text-lg font-semibold text-foreground">
                {t('modelConfig.availableModels', 'Available Models')}
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                {t('modelConfig.modelsDescription', 'Browse models from channels you have already connected.')}
              </p>
            </div>

            {hasAvailableModels ? (
              <ModelSelector
                channels={configuredChannels}
              />
            ) : (
              <div className="p-4 rounded-lg bg-muted/50 text-center">
                <p className="text-sm font-medium text-foreground">
                  {t('modelConfig.noModelsAvailable', 'No models available')}
                </p>
                <p className="text-xs text-muted-foreground mt-1">
                  {hasVisibleChannels
                    ? t('modelConfig.fetchModelsHint', 'Fetch models from a channel to get started')
                    : t('modelConfig.hiddenSystemChannelsHint', 'Add and configure a channel before models appear here.')}
                </p>
              </div>
            )}
          </section>
        </div>

        <div className="space-y-6">
          {/* Appearance Settings — placed at top of right column for visibility */}
          <section className="bg-card rounded-xl border border-border/30 p-6">
            <div className="mb-4">
              <h2 className="text-lg font-semibold text-foreground">
                {t('settings.appearance', 'Appearance')}
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                {t('settings.appearanceDescription', 'Customize the look and feel of the application.')}
              </p>
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

          <section className="bg-card rounded-xl border border-border/30 p-6">
            <div className="mb-4">
              <h2 className="text-lg font-semibold text-foreground">
                {t('modelConfig.roleMapping', 'Role Mapping')}
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                {t('modelConfig.roleMappingDescription', 'Assign roles only from enabled, configured models')}
              </p>
            </div>

            {hasAvailableModels ? (
              <RoleMapping
                channels={configuredChannels}
                assignments={role_assignments}
                onAssign={assignRole}
                onClear={clearRole}
              />
            ) : (
              <div className="p-4 rounded-lg bg-muted/50 text-center">
                <p className="text-sm font-medium text-foreground">
                  {t('modelConfig.roleMapping', 'Role Mapping')}
                </p>
                <p className="text-xs text-muted-foreground mt-1">
                  {t('modelConfig.roleMappingEmptyHint', 'Configure a channel first, then assign retrieval and generation roles here.')}
                </p>
              </div>
            )}
          </section>

          <section className="bg-card rounded-xl border border-border/30 p-6">
            <h2 className="text-base font-semibold text-foreground">
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
      <div className="mt-6">
        <section className="bg-card rounded-xl border border-border/30 p-6">
          <div className="mb-4">
            <h2 className="text-lg font-semibold text-foreground">
              {t('template.title', 'Template Management')}
            </h2>
            <p className="text-sm text-muted-foreground mt-1">
              {t('template.description', 'Manage reusable prompt templates for common figure types.')}
            </p>
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
  );
}
