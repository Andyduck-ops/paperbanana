import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

// ============================================================================
// Types
// ============================================================================

export type LanguageCode = 'zh' | 'en';
export type ColorScheme = 'light' | 'dark' | 'system';

export type Page = 'main' | 'provider-new' | 'provider-edit';
export type MainTab = 'generate' | 'refine';

export interface DrawerState {
  history: boolean;
  settings: boolean;
}

export interface ModalState {
  export: boolean;
  shortcutsHelp: boolean;
  welcomeWizard: boolean;
}

export interface UIState {
  // Navigation
  currentPage: Page;
  editingProvider: string | undefined;
  mainTab: MainTab;

  // Drawers
  drawers: DrawerState;

  // Modals
  modals: ModalState;

  // Theme & Language
  colorScheme: ColorScheme;
  language: LanguageCode;

  // Other UI state
  examplePrompt: string | null;
  showWelcomeWizard: boolean;
}

export interface UIActions {
  // Navigation actions
  setCurrentPage: (page: Page) => void;
  setEditingProvider: (providerId: string | undefined) => void;
  setMainTab: (tab: MainTab) => void;

  // Drawer actions
  openDrawer: (drawer: keyof DrawerState) => void;
  closeDrawer: (drawer: keyof DrawerState) => void;
  toggleDrawer: (drawer: keyof DrawerState) => void;
  closeAllDrawers: () => void;

  // Modal actions
  openModal: (modal: keyof ModalState) => void;
  closeModal: (modal: keyof ModalState) => void;
  toggleModal: (modal: keyof ModalState) => void;
  closeAllModals: () => void;

  // Theme actions
  setColorScheme: (scheme: ColorScheme) => void;

  // Language actions
  setLanguage: (language: LanguageCode) => void;

  // Other actions
  setExamplePrompt: (prompt: string | null) => void;
  setShowWelcomeWizard: (show: boolean) => void;

  // Reset
  resetUI: () => void;
}

// ============================================================================
// Constants
// ============================================================================

export const LANGUAGES = [
  { id: 'zh' as const, name: '中文' },
  { id: 'en' as const, name: 'English' },
] as const;

// ============================================================================
// Helpers
// ============================================================================

function getEffectiveScheme(scheme: ColorScheme): 'light' | 'dark' {
  if (scheme !== 'system') return scheme;
  if (typeof window === 'undefined') return 'light';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

// ============================================================================
// Store
// ============================================================================

const initialState: UIState = {
  currentPage: 'main',
  editingProvider: undefined,
  mainTab: 'generate',
  drawers: {
    history: false,
    settings: false,
  },
  modals: {
    export: false,
    shortcutsHelp: false,
    welcomeWizard: false,
  },
  colorScheme: 'system',
  language: 'zh',
  examplePrompt: null,
  showWelcomeWizard: false,
};

export const useAppStore = create<UIState & UIActions>()(
  persist(
    (set) => ({
      ...initialState,

      // Navigation actions
      setCurrentPage: (page) => set({ currentPage: page }),
      setEditingProvider: (providerId) => set({ editingProvider: providerId }),
      setMainTab: (tab) => set({ mainTab: tab }),

      // Drawer actions
      openDrawer: (drawer) =>
        set((state) => ({
          drawers: { ...state.drawers, [drawer]: true },
        })),
      closeDrawer: (drawer) =>
        set((state) => ({
          drawers: { ...state.drawers, [drawer]: false },
        })),
      toggleDrawer: (drawer) =>
        set((state) => ({
          drawers: { ...state.drawers, [drawer]: !state.drawers[drawer] },
        })),
      closeAllDrawers: () =>
        set({
          drawers: { history: false, settings: false },
        }),

      // Modal actions
      openModal: (modal) =>
        set((state) => ({
          modals: { ...state.modals, [modal]: true },
        })),
      closeModal: (modal) =>
        set((state) => ({
          modals: { ...state.modals, [modal]: false },
        })),
      toggleModal: (modal) =>
        set((state) => ({
          modals: { ...state.modals, [modal]: !state.modals[modal] },
        })),
      closeAllModals: () =>
        set({
          modals: { export: false, shortcutsHelp: false, welcomeWizard: false },
        }),

      // Theme actions
      setColorScheme: (colorScheme) => {
        set({ colorScheme });
        if (typeof window !== 'undefined') {
          document.documentElement.setAttribute('data-color-scheme', getEffectiveScheme(colorScheme));
        }
      },

      // Language actions
      setLanguage: (language) => set({ language }),

      // Other actions
      setExamplePrompt: (prompt) => set({ examplePrompt: prompt }),
      setShowWelcomeWizard: (show) => set({ showWelcomeWizard: show }),

      // Reset
      resetUI: () => set(initialState),
    }),
    {
      name: 'paperbanana-app-store',
      version: 2,
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        colorScheme: state.colorScheme,
        language: state.language,
      }),
      migrate: (persisted: unknown, _version) => {
        // v1 → v2: drop legacy `theme` field. Cast through unknown so the
        // pre-migration shape doesn't leak into the v2 type.
        if (persisted && typeof persisted === 'object') {
          const next = { ...(persisted as Record<string, unknown>) };
          delete next.theme;
          return next as Partial<UIState>;
        }
        return persisted as Partial<UIState>;
      },
      onRehydrateStorage: () => (state) => {
        if (!state || typeof window === 'undefined') return;
        // Clear any legacy `data-theme` attribute left over from the 14-theme era.
        document.documentElement.removeAttribute('data-theme');
        document.documentElement.setAttribute(
          'data-color-scheme',
          getEffectiveScheme(state.colorScheme)
        );
      },
    }
  )
);

// ============================================================================
// Selectors
// ============================================================================

export const selectColorScheme = (state: UIState & UIActions) => state.colorScheme;
export const selectLanguage = (state: UIState & UIActions) => state.language;
export const selectCurrentPage = (state: UIState & UIActions) => state.currentPage;
export const selectMainTab = (state: UIState & UIActions) => state.mainTab;
export const selectDrawers = (state: UIState & UIActions) => state.drawers;
export const selectModals = (state: UIState & UIActions) => state.modals;
export const selectIsHistoryOpen = (state: UIState & UIActions) => state.drawers.history;
export const selectIsSettingsOpen = (state: UIState & UIActions) => state.drawers.settings;
export const selectIsExportOpen = (state: UIState & UIActions) => state.modals.export;
export const selectIsShortcutsHelpOpen = (state: UIState & UIActions) => state.modals.shortcutsHelp;
export const selectIsWelcomeWizardOpen = (state: UIState & UIActions) => state.modals.welcomeWizard;

// ============================================================================
// Hooks for backward compatibility
// ============================================================================

export function useLanguageFromStore(): {
  language: LanguageCode;
  setLanguage: (language: LanguageCode) => void;
  languages: typeof LANGUAGES;
} {
  const language = useAppStore(selectLanguage);
  const setLanguage = useAppStore((state) => state.setLanguage);
  return { language, setLanguage, languages: LANGUAGES };
}
