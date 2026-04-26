// Atomic Design - Atoms
export {
  Button,
  type ButtonProps,
  type ButtonVariant,
  type ButtonSize,
} from './atoms';

// Layout components
export { Layout, type LayoutProps } from './layout/Layout';
export { Header, type HeaderProps } from './layout/Header';
export { Footer } from './layout/Footer';
export { MinimalLayout, type MinimalLayoutProps } from './layout/MinimalLayout';
export { FloatingInput, type FloatingInputProps } from './input/FloatingInput';
export { ProgressLine, type ProgressLineProps } from './progress/ProgressLine';

// Model Configuration Components (V2 - Unified Model Config System)
export { ModelIndicator } from './model/ModelIndicator';
export { ModelPopover } from './model/ModelPopover';
export { ChannelManager, type ChannelManagerProps } from './model/ChannelManager';
export { ModelSelector, CompactModelSelector, type ModelSelectorProps, type CompactModelSelectorProps } from './model/ModelSelector';
export { RoleMapping, type RoleMappingProps } from './model/RoleMapping';

// Generation components
export { GeneratePanel, type GeneratePanelProps, type GenerateOptions } from './GeneratePanel';
export { ProgressPanel, type ProgressPanelProps } from './ProgressPanel';
export type { StageState, ResumeMetadata } from '../hooks/useGenerate';
export { StageCard, type StageStatus, type StageCardProps } from './StageCard';
export { ResultPanel, type ResultPanelProps } from './ResultPanel';
export { ArtifactPreview, type ArtifactPreviewProps, type Artifact } from './ArtifactPreview';
export { BatchProgressPanel, type BatchProgressPanelProps, type BatchProgress, type BatchCandidate } from './BatchProgressPanel';

// History components (V2 - Sliding Panel)
export {
  HistoryTrigger,
  HistoryPopover,
  HistoryPanel,
  HistoryItem,
  type HistoryPanelProps,
  type HistoryItemProps,
  type HistoryPopoverProps,
} from './history';

// Legacy History components removed — use V2 history/ exports above

// Export components
export { ExportModal, type ExportModalProps, type ExportFormat } from './ExportModal';

// Refine components
export { RefinePanel, type RefinePanelProps } from './RefinePanel';
export { IterationTimeline, type IterationTimelineProps, type RefineIteration } from './refine/IterationTimeline';
export { BatchVariantGrid, type BatchVariantGridProps, type BatchVariant } from './refine/BatchVariantGrid';

// UX components
export { Toast, type ToastProps } from './Toast';
export { ErrorBoundary } from './ErrorBoundary';
export { FieldError, type FieldErrorProps } from './FieldError';

// Theme components
export { DarkModeToggle, useDarkMode } from './theme/DarkModeToggle';

// Loading components
export { SkeletonLoader } from './SkeletonLoader';

// Settings components
export { SettingsDrawer, type SettingsDrawerProps } from './settings/SettingsDrawer';

// Workspace Core Components (V2)
export {
  Workspace,
  WorkspaceHero,
  ModeSwitcher,
  EmptyState,
  CandidateGrid,
  ResultArea,
  type WorkspaceProps,
  type WorkspaceHeroProps,
  type ModeSwitcherProps,
  type WorkspaceMode,
  type EmptyStateProps,
  type CandidateGridProps,
  type Candidate,
  type ResultAreaProps,
} from './workspace';

// Welcome Wizard
export {
  WelcomeWizard,
  isWizardCompleted,
  markWizardCompleted,
  type WelcomeWizardProps,
} from './WelcomeWizard';

// Project Selector
export {
  ProjectSelector,
  type ProjectSelectorProps,
  type Project,
} from './ProjectSelector';

// Template Selector
export {
  TemplateSelector,
  type TemplateSelectorProps,
} from './TemplateSelector';

// Shortcuts Help Panel
export {
  ShortcutsHelpPanel,
  type ShortcutsHelpPanelProps,
} from './ShortcutsHelpPanel';

// Folder & Version Management Components
export {
  FolderTree,
  VersionTimeline,
  TemplateManager,
  type FolderTreeProps,
  type VersionTimelineProps,
  type TemplateManagerProps,
} from './folder';
