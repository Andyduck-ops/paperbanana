import { useLanguage } from '../../hooks';

export type WorkspaceMode = 'generate' | 'refine';

export interface ModeSwitcherProps {
  mode: WorkspaceMode;
  onChange: (mode: WorkspaceMode) => void;
  disabled?: boolean;
}

/**
 * Mode Switcher Component
 *
 * Lightweight mode switching between Generation and Refinement.
 * Designed to feel like a mode change, not a page navigation.
 *
 * Visual design:
 * - Segmented control style with animated sliding indicator
 * - Clear active state with background highlight
 * - Icon + label for quick recognition
 * - Subtle hover states
 */
export function ModeSwitcher({
  mode,
  onChange,
  disabled = false,
}: ModeSwitcherProps) {
  const { t } = useLanguage();

  return (
    <div
      className="mode-switcher inline-flex items-center gap-1 p-1 rounded-xl bg-muted/60 border border-border/50 relative"
      role="radiogroup"
      aria-label="Workspace mode"
    >
      <button
        type="button"
        role="radio"
        aria-checked={mode === 'generate'}
        onClick={() => onChange('generate')}
        disabled={disabled}
        className={`
          mode-switcher__option
          relative z-10 flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium
          transition-all duration-200 ease-out
          focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
          ${mode === 'generate'
            ? 'text-primary-foreground'
            : 'text-muted-foreground hover:text-foreground'
          }
          ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
        `}
      >
        {/* Sparkles icon for generation */}
        <svg
          className="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"
          />
        </svg>
        <span>{t('app.tabGenerate')}</span>
      </button>

      <button
        type="button"
        role="radio"
        aria-checked={mode === 'refine'}
        onClick={() => onChange('refine')}
        disabled={disabled}
        className={`
          mode-switcher__option
          relative z-10 flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium
          transition-all duration-200 ease-out
          focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
          ${mode === 'refine'
            ? 'text-primary-foreground'
            : 'text-muted-foreground hover:text-foreground'
          }
          ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
        `}
      >
        {/* Wand icon for refinement */}
        <svg
          className="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
          />
        </svg>
        <span>{t('app.tabRefine')}</span>
      </button>
    </div>
  );
}

export default ModeSwitcher;
