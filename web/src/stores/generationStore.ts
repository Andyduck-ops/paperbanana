import { create } from 'zustand';

// ============================================================================
// Types
// ============================================================================

export type LocalWorkMode = 'generate' | 'batch' | 'refine';

export interface Artifact {
  kind: string;
  mimeType: string;
  summary?: string;
  data?: string;
  assetId?: string;
  projectId?: string;
  uri?: string;
}

export interface PendingHistoryContext {
  prompt: string;
  mode: LocalWorkMode;
}

export interface GenerationState {
  // Router state
  currentPath: string;

  // History session selection
  selectedSessionId: string | undefined;

  // Batch selection
  selectedBatchCandidateId: string | null;

  // Refine image data
  refineSeedImageData: string | null;

  // Export modal artifact
  exportArtifact: Artifact | undefined;

  // Pending history context for tracking
  pendingHistoryContext: PendingHistoryContext | null;
}

export interface GenerationActions {
  // Router actions
  setCurrentPath: (path: string) => void;
  navigateToProjects: () => void;
  navigateToMain: () => void;

  // History session actions
  setSelectedSessionId: (sessionId: string | undefined) => void;
  clearSelectedSessionId: () => void;

  // Batch candidate actions
  setSelectedBatchCandidateId: (candidateId: string | null) => void;
  clearSelectedBatchCandidateId: () => void;

  // Refine seed image actions
  setRefineSeedImageData: (imageData: string | null) => void;
  clearRefineSeedImageData: () => void;

  // Export artifact actions
  setExportArtifact: (artifact: Artifact | undefined) => void;
  clearExportArtifact: () => void;

  // Pending history context actions
  setPendingHistoryContext: (context: PendingHistoryContext | null) => void;
  clearPendingHistoryContext: () => void;

  // Reset all generation state
  resetGenerationState: () => void;
}

// ============================================================================
// Initial State
// ============================================================================

const initialState: GenerationState = {
  currentPath: window.location.pathname,
  selectedSessionId: undefined,
  selectedBatchCandidateId: null,
  refineSeedImageData: null,
  exportArtifact: undefined,
  pendingHistoryContext: null,
};

// ============================================================================
// Store
// ============================================================================

export const useGenerationStore = create<GenerationState & GenerationActions>()((set) => ({
  ...initialState,

  // Router actions
  setCurrentPath: (path) => set({ currentPath: path }),
  navigateToProjects: () => {
    window.history.pushState({}, '', '/projects');
    set({ currentPath: '/projects' });
  },
  navigateToMain: () => {
    window.history.pushState({}, '', '/');
    set({ currentPath: '/' });
  },

  // History session actions
  setSelectedSessionId: (sessionId) => set({ selectedSessionId: sessionId }),
  clearSelectedSessionId: () => set({ selectedSessionId: undefined }),

  // Batch candidate actions
  setSelectedBatchCandidateId: (candidateId) => set({ selectedBatchCandidateId: candidateId }),
  clearSelectedBatchCandidateId: () => set({ selectedBatchCandidateId: null }),

  // Refine seed image actions
  setRefineSeedImageData: (imageData) => set({ refineSeedImageData: imageData }),
  clearRefineSeedImageData: () => set({ refineSeedImageData: null }),

  // Export artifact actions
  setExportArtifact: (artifact) => set({ exportArtifact: artifact }),
  clearExportArtifact: () => set({ exportArtifact: undefined }),

  // Pending history context actions
  setPendingHistoryContext: (context) => set({ pendingHistoryContext: context }),
  clearPendingHistoryContext: () => set({ pendingHistoryContext: null }),

  // Reset
  resetGenerationState: () => set(initialState),
}));

// ============================================================================
// Selectors
// ============================================================================

export const selectCurrentPath = (state: GenerationState & GenerationActions) => state.currentPath;
export const selectSelectedSessionId = (state: GenerationState & GenerationActions) => state.selectedSessionId;
export const selectSelectedBatchCandidateId = (state: GenerationState & GenerationActions) => state.selectedBatchCandidateId;
export const selectRefineSeedImageData = (state: GenerationState & GenerationActions) => state.refineSeedImageData;
export const selectExportArtifact = (state: GenerationState & GenerationActions) => state.exportArtifact;
export const selectPendingHistoryContext = (state: GenerationState & GenerationActions) => state.pendingHistoryContext;
export const selectIsProjectsPage = (state: GenerationState & GenerationActions) => state.currentPath === '/projects';
