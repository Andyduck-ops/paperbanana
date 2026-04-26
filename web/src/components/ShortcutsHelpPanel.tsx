import { useEffect } from 'react';
import { useLanguage, useFocusTrap } from '../hooks';

export interface ShortcutsHelpPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

interface Shortcut {
  key: string;
  label: string;
  description?: string;
}

export function ShortcutsHelpPanel({ isOpen, onClose }: ShortcutsHelpPanelProps) {
  const { t } = useLanguage();
  const dialogRef = useFocusTrap<HTMLDivElement>(isOpen);

  const shortcuts: Shortcut[] = [
    { key: 'Ctrl/Cmd + N', label: t('shortcuts.newGeneration') },
    { key: 'Ctrl/Cmd + E', label: t('shortcuts.export') },
    { key: 'Ctrl/Cmd + H', label: t('shortcuts.focusHistory') },
    { key: '?', label: t('shortcuts.showHelp') },
    { key: 'Esc', label: t('shortcuts.closeModal') },
  ];

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      <div
        ref={dialogRef}
        className="relative bg-popover border border-border rounded-xl shadow-2xl max-w-md w-full p-6"
        role="dialog"
        aria-modal="true"
        aria-labelledby="shortcuts-title"
      >
        <div className="flex items-center justify-between mb-6">
          <h2 id="shortcuts-title" className="text-lg font-semibold">
            {t('shortcuts.title')}
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground p-1 rounded"
            aria-label={t('common.close')}
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="space-y-3">
          {shortcuts.map((shortcut) => (
            <div
              key={shortcut.key}
              className="flex items-center justify-between py-2 border-b border-border/50 last:border-b-0"
            >
              <span className="text-sm text-foreground">{shortcut.label}</span>
              <kbd className="px-2 py-1 text-xs font-mono bg-muted rounded border border-border">
                {shortcut.key}
              </kbd>
            </div>
          ))}
        </div>

        <p className="mt-6 text-xs text-muted-foreground text-center">
          Press <kbd className="px-1 py-0.5 bg-muted rounded">?</kbd> to show this help anytime
        </p>
      </div>
    </div>
  );
}
