// Layout components
export { Layout, type LayoutProps } from './Layout';
export { Header } from './Header';
export { Footer } from './Footer';

// Minimal layout components (V2)
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
export { ProgressPanel, type ProgressPanelProps, type StageState } from './ProgressPanel';
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

// Legacy History components (deprecated, use V2 above)
export { HistorySidebar, type HistorySidebarProps } from './HistorySidebar';
export { HistoryItem as HistoryItemLegacy, type HistoryItemProps as HistoryItemLegacyProps } from './HistoryItem';

// Export components
export { ExportModal, type ExportModalProps, type ExportFormat } from './ExportModal';

// Refine components
export { RefinePanel, type RefinePanelProps } from './RefinePanel';
export { IterationTimeline, type IterationTimelineProps, type RefineIteration } from './refine/IterationTimeline';
export { BatchVariantGrid, type BatchVariantGridProps, type BatchVariant } from './refine/BatchVariantGrid';

// UX components
export { Toast, type ToastProps } from './Toast';
export { ErrorBoundary } from './ErrorBoundary';

// Theme components
export { ThemeSelector } from './ThemeSelector';

// Workspace Core Components (V2)
export {
  Workspace,
  ModeSwitcher,
  EmptyState,
  CandidateGrid,
  ResultArea,
  type WorkspaceProps,
  type ModeSwitcherProps,
  type WorkspaceMode,
  type EmptyStateProps,
  type EmptyStateAction,
  type CandidateGridProps,
  type Candidate,
  type ResultAreaProps,
} from './workspace';
