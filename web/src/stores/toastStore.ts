import { create } from 'zustand';

// ============================================================================
// Types
// ============================================================================

export type ToastType = 'success' | 'error' | 'info';

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

export interface ToastState {
  toasts: Toast[];
}

export interface ToastActions {
  addToast: (message: string, type?: ToastType) => void;
  removeToast: (id: string) => void;
  clearToasts: () => void;
}

// ============================================================================
// Store
// ============================================================================

export const useToastStore = create<ToastState & ToastActions>((set) => ({
  toasts: [],

  addToast: (message, type = 'info') => {
    const id = `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;
    set((state) => ({
      toasts: [...state.toasts, { id, message, type }],
    }));
    // Auto-dismiss after 5 seconds
    setTimeout(() => {
      set((state) => ({
        toasts: state.toasts.filter((t) => t.id !== id),
      }));
    }, 5000);
  },

  removeToast: (id) =>
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    })),

  clearToasts: () => set({ toasts: [] }),
}));

// ============================================================================
// Selectors
// ============================================================================

export const selectToasts = (state: ToastState & ToastActions) => state.toasts;
export const selectAddToast = (state: ToastState & ToastActions) => state.addToast;
export const selectRemoveToast = (state: ToastState & ToastActions) => state.removeToast;
