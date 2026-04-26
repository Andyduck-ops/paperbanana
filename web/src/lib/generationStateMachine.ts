/**
 * Generation State Machine
 * 
 * Polanyi-inspired explicit knowledge representation:
 * - Tacit knowledge about generation flow made explicit
 * - State transitions are declarative and testable
 * - Side effects are isolated and observable
 * 
 * States represent the "what", transitions represent the "when"
 */

export type GenerationMode = 'idle' | 'single' | 'batch' | 'refine';

export interface GenerationState {
  // Current mode
  mode: GenerationMode;
  
  // Active process flags
  isGenerating: boolean;
  isRefining: boolean;
  isBatchGenerating: boolean;
  
  // Current operation state
  currentSessionId: string | null;
  currentBatchId: string | null;
  selectedCandidateId: string | null;
  
  // UI state
  pendingPrompt: string | null;
  refineSeedImage: string | null;
  
  // Error state
  error: string | null;
}

export type GenerationAction =
  | { type: 'START_SINGLE'; prompt: string; sessionId: string }
  | { type: 'START_BATCH'; prompt: string; batchId: string; count: number }
  | { type: 'START_REFINE'; imageData: string; instructions: string }
  | { type: 'COMPLETE_SINGLE'; sessionId: string }
  | { type: 'COMPLETE_BATCH'; batchId: string }
  | { type: 'COMPLETE_REFINE'; sessionId: string }
  | { type: 'FAIL'; error: string }
  | { type: 'CANCEL' }
  | { type: 'RESET_ALL' }
  | { type: 'RESET_FOR_NEW' }
  | { type: 'SELECT_CANDIDATE'; candidateId: string }
  | { type: 'LOAD_EXAMPLE'; prompt: string }
  | { type: 'CLEAR_ERROR' };

// Initial state factory - ensures clean state creation
export function createInitialState(): GenerationState {
  return {
    mode: 'idle',
    isGenerating: false,
    isRefining: false,
    isBatchGenerating: false,
    currentSessionId: null,
    currentBatchId: null,
    selectedCandidateId: null,
    pendingPrompt: null,
    refineSeedImage: null,
    error: null,
  };
}

// Pure state reducer - all state logic in one place
export function generationReducer(
  state: GenerationState,
  action: GenerationAction
): GenerationState {
  switch (action.type) {
    case 'START_SINGLE':
      return {
        ...state,
        mode: 'single',
        isGenerating: true,
        isBatchGenerating: false,
        isRefining: false,
        currentSessionId: action.sessionId,
        currentBatchId: null,
        selectedCandidateId: null,
        pendingPrompt: action.prompt,
        refineSeedImage: null,
        error: null,
      };

    case 'START_BATCH':
      return {
        ...state,
        mode: 'batch',
        isGenerating: false,
        isBatchGenerating: true,
        isRefining: false,
        currentSessionId: null,
        currentBatchId: action.batchId,
        selectedCandidateId: null,
        pendingPrompt: action.prompt,
        refineSeedImage: null,
        error: null,
      };

    case 'START_REFINE':
      return {
        ... state,
        mode: 'refine',
        isGenerating: false,
        isBatchGenerating: false,
        isRefining: true,
        currentSessionId: null,
        pendingPrompt: action.instructions,
        refineSeedImage: action.imageData,
        error: null,
      };

    case 'COMPLETE_SINGLE':
      return {
        ...state,
        isGenerating: false,
        mode: state.mode === 'single' ? 'idle' : state.mode,
      };

    case 'COMPLETE_BATCH':
      return {
        ...state,
        isBatchGenerating: false,
        mode: state.mode === 'batch' ? 'idle' : state.mode,
      };

    case 'COMPLETE_REFINE':
      return {
        ...state,
        isRefining: false,
        mode: state.mode === 'refine' ? 'idle' : state.mode,
      };

    case 'FAIL':
      return {
        ...state,
        isGenerating: false,
        isBatchGenerating: false,
        isRefining: false,
        error: action.error,
      };

    case 'CANCEL':
      return {
        ...state,
        isGenerating: false,
        isBatchGenerating: false,
        isRefining: false,
        error: null,
      };

    case 'RESET_ALL':
      return createInitialState();

    case 'RESET_FOR_NEW':
      // Preserve mode but clear operation state
      return {
        ...createInitialState(),
        mode: 'idle',
      };

    case 'SELECT_CANDIDATE':
      return {
        ...state,
        selectedCandidateId: action.candidateId,
      };

    case 'LOAD_EXAMPLE':
      return {
        ...createInitialState(),
        pendingPrompt: action.prompt,
      };

    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,
      };

    default:
      return state;
  }
}

// State selectors - derived state logic
export const GenerationSelectors = {
  isActive: (state: GenerationState) => 
    state.isGenerating || state.isBatchGenerating || state.isRefining,
  
  isIdle: (state: GenerationState) => 
    state.mode === 'idle' && !GenerationSelectors.isActive(state),
  
  canStartNew: (state: GenerationState) => 
    !GenerationSelectors.isActive(state),
  
  hasSelection: (state: GenerationState) => 
    state.selectedCandidateId !== null,
  
  currentOperationId: (state: GenerationState) => 
    state.currentSessionId ?? state.currentBatchId,
};

// Valid transitions - explicit state machine edges
export const ValidTransitions: Record<GenerationMode, GenerationMode[]> = {
  idle: ['single', 'batch', 'refine'],
  single: ['idle'],
  batch: ['idle'],
  refine: ['idle'],
};

export function canTransition(from: GenerationMode, to: GenerationMode): boolean {
  return ValidTransitions[from].includes(to);
}
