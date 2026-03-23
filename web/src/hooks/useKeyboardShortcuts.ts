import { useEffect, useCallback } from 'react';

export interface ShortcutHandlers {
  onNewGeneration?: () => void;
  onExport?: () => void;
  onFocusHistory?: () => void;
  onEscape?: () => void;
}

export function useKeyboardShortcuts(handlers: ShortcutHandlers) {
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    const isMeta = e.metaKey || e.ctrlKey;

    // Ctrl/Cmd + N: New generation
    if (isMeta && e.key === 'n') {
      e.preventDefault();
      handlers.onNewGeneration?.();
    }

    // Ctrl/Cmd + E: Export
    if (isMeta && e.key === 'e') {
      e.preventDefault();
      handlers.onExport?.();
    }

    // Ctrl/Cmd + H: Focus history
    if (isMeta && e.key === 'h') {
      e.preventDefault();
      handlers.onFocusHistory?.();
    }

    // Escape: Close modals
    if (e.key === 'Escape') {
      handlers.onEscape?.();
    }
  }, [handlers]);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);
}
