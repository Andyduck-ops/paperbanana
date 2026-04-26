import { useEffect, useRef } from 'react';

export interface ShortcutHandlers {
  onNewGeneration?: () => void;
  onExport?: () => void;
  onFocusHistory?: () => void;
  onEscape?: () => void;
  onShowShortcuts?: () => void;
}

export function useKeyboardShortcuts(handlers: ShortcutHandlers) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Skip repeated keys (key hold)
      if (e.repeat) return;

      const isMeta = e.metaKey || e.ctrlKey;
      const current = handlersRef.current;

      // Ctrl/Cmd + N: New generation
      if (isMeta && e.key === 'n') {
        e.preventDefault();
        current.onNewGeneration?.();
        return;
      }

      // Ctrl/Cmd + E: Export
      if (isMeta && e.key === 'e') {
        e.preventDefault();
        current.onExport?.();
        return;
      }

      // Ctrl/Cmd + H: Focus history
      if (isMeta && e.key === 'h') {
        e.preventDefault();
        current.onFocusHistory?.();
        return;
      }

      // Escape: Close modals
      if (e.key === 'Escape') {
        current.onEscape?.();
        return;
      }

      // ?: Show shortcuts help (when not in input)
      if (e.key === '?' && !e.metaKey && !e.ctrlKey && !e.altKey) {
        const target = e.target as HTMLElement;
        const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable;
        if (!isInput) {
          e.preventDefault();
          current.onShowShortcuts?.();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);
}
