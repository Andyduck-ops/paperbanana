// ============================================================================
// Store Exports
// ============================================================================

// App Store - UI State, Theme, Language, Drawers, Modals
export {
  useAppStore,
  useThemeFromStore,
  useLanguageFromStore,
  THEMES,
  LANGUAGES,
  // Selectors
  selectTheme,
  selectLanguage,
  selectCurrentPage,
  selectMainTab,
  selectDrawers,
  selectModals,
  selectIsHistoryOpen,
  selectIsSettingsOpen,
  selectIsExportOpen,
  selectIsShortcutsHelpOpen,
  selectIsWelcomeWizardOpen,
} from './appStore';

export type {
  Theme,
  LanguageCode,
  Page,
  MainTab,
  DrawerState,
  ModalState,
  UIState,
  UIActions,
} from './appStore';

// Provider Store - Provider/Channel State, Role Assignments
export {
  useProviderStore,
  fetchProviders,
  createProvider,
  updateProviderAPI,
  deleteProviderAPI,
  fetchProviderModelsAPI,
  assignRoleAPI,
  // Selectors
  selectProviders,
  selectRoleAssignments,
  selectLoading,
  selectError,
  selectSnapshots,
} from './providerStore';

export type {
  ModelInfo,
  Provider,
  Channel,
  WorkflowRole,
  RoleAssignment,
  ModelSnapshot,
  ProviderState,
  ProviderActions,
} from './providerStore';

// Generation Store - Generation State, Router, Session Selection
export {
  useGenerationStore,
  // Selectors
  selectCurrentPath,
  selectSelectedSessionId,
  selectSelectedBatchCandidateId,
  selectRefineSeedImageData,
  selectExportArtifact,
  selectPendingHistoryContext,
  selectIsProjectsPage,
} from './generationStore';

export type {
  LocalWorkMode,
  Artifact,
  PendingHistoryContext,
  GenerationState,
  GenerationActions,
} from './generationStore';

// Toast Store - Global Toast Notifications
export {
  useToastStore,
  selectToasts,
  selectAddToast,
  selectRemoveToast,
} from './toastStore';

export type {
  ToastType,
  Toast,
  ToastState,
  ToastActions,
} from './toastStore';
