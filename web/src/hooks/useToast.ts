import { useCallback } from 'react';
import { useToastStore, type ToastType } from '../stores/toastStore';

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

export function useToast() {
  const toasts = useToastStore((state) => state.toasts);
  const addToast = useToastStore((state) => state.addToast);
  const removeToast = useToastStore((state) => state.removeToast);

  // Wrapped addToast to support the legacy 5s auto-dismiss behavior
  const wrappedAddToast = useCallback((message: string, type: ToastType = 'info') => {
    addToast(message, type);
  }, [addToast]);

  return {
    toasts,
    addToast: wrappedAddToast,
    removeToast,
  };
}
