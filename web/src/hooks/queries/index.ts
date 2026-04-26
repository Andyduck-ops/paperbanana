// TanStack Query hooks for server state management

export {
  useProvidersQuery,
  useProviderQuery,
  useCreateProviderMutation,
  useUpdateProviderMutation,
  useDeleteProviderMutation,
  useProvidersCompat,
} from './useProvidersQuery';

export {
  useGenerateMutation,
  useGenerateCompat,
  type GenerateOptions,
} from './useGenerateMutation';

export {
  useHistoryQuery,
  useHistoryItemQuery,
  useRestoreSessionMutation,
  useDeleteHistoryItemMutation,
  useClearHistoryMutation,
  useHistoryCompat,
  type HistoryItem,
} from './useHistoryQuery';

export {
  useRefineMutation,
  useRefineCompat,
  type RefineImage,
  type RefineOptions,
  type RefineResult,
} from './useRefineMutation';

export {
  useProjectsQuery,
  useProjectQuery,
  useCreateProjectMutation,
  useUpdateProjectMutation,
  useDeleteProjectMutation,
  useProjectFoldersQuery,
  useProjectItemsQuery,
  useProjectsCompat,
  type Project,
  type ProjectFolder,
  type ProjectItem,
} from './useProjectsQuery';
