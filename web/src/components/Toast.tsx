import type { Toast as ToastType } from '../hooks/useToast';

export interface ToastProps {
  toasts: ToastType[];
  onRemove: (id: string) => void;
}

export function Toast({ toasts, onRemove }: ToastProps) {
  if (toasts.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50 space-y-2">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`
            px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 min-w-64
            ${toast.type === 'error' ? 'bg-red-600 text-white' : ''}
            ${toast.type === 'success' ? 'bg-green-600 text-white' : ''}
            ${toast.type === 'info' ? 'bg-primary text-background' : ''}
          `}
        >
          <span className="flex-1">{toast.message}</span>
          <button
            onClick={() => onRemove(toast.id)}
            className="text-current opacity-70 hover:opacity-100"
          >
            ✕
          </button>
        </div>
      ))}
    </div>
  );
}
