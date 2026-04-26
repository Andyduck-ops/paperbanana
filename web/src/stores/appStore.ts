import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

// ============================================================================
// Types
// ============================================================================

export type Theme = 'academic' | 'qi-baishi' | 'pop-anime' | 'rococo' | 'japanese-bw';
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
  theme: Theme;
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
  setTheme: (theme: Theme) => void;
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

export const THEMES = [
  { id: 'academic' as const, name: 'Academic' },
  { id: 'qi-baishi' as const, name: 'Qi Baishi' },
  { id: 'pop-anime' as const, name: 'Pop Anime' },
  { id: 'rococo' as const, name: 'Rococo' },
  { id: 'japanese-bw' as const, name: 'Night Mono' },
] as const;

export const LANGUAGES = [
  { id: 'zh' as const, name: '中文' },
  { id: 'en' as const, name: 'English' },
] as const;

// Migration map for old theme IDs
const THEME_MIGRATION_MAP: Record<string, Theme> = {
  'pop-art': 'pop-anime',
  'classical-chinese': 'qi-baishi',
  'minimalist-bw': 'japanese-bw',
};

function migrateTheme(oldTheme: string): Theme {
  return THEME_MIGRATION_MAP[oldTheme] || (isValidTheme(oldTheme) ? oldTheme : 'qi-baishi');
}

function isValidTheme(theme: string): theme is Theme {
  return ['academic', 'qi-baishi', 'pop-anime', 'rococo', 'japanese-bw'].includes(theme);
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
  theme: 'qi-baishi',
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
      setTheme: (theme) => {
        set({ theme });
        if (typeof window !== 'undefined') {
          document.documentElement.setAttribute('data-theme', theme);
        }
      },
      setColorScheme: (colorScheme) => {
        set({ colorScheme });
        if (typeof window !== 'undefined') {
          const effective = colorScheme === 'system'
            ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
            : colorScheme;
          document.documentElement.setAttribute('data-color-scheme', effective);
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
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        theme: state.theme,
        colorScheme: state.colorScheme,
        language: state.language,
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          // Apply theme on rehydrate
          if (typeof window !== 'undefined') {
            // Migrate old theme if needed
            const migratedTheme = migrateTheme(state.theme);
            if (migratedTheme !== state.theme) {
              state.theme = migratedTheme;
            }
            document.documentElement.setAttribute('data-theme', state.theme);
            const effectiveScheme = state.colorScheme === 'system'
              ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
              : state.colorScheme;
            document.documentElement.setAttribute('data-color-scheme', effectiveScheme);
          }
        }
      },
    }
  )
);

// ============================================================================
// Selectors
// ============================================================================

export const selectTheme = (state: UIState & UIActions) => state.theme;
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

export function useThemeFromStore(): { theme: Theme; setTheme: (theme: Theme) => void; themes: typeof THEMES } {
  const theme = useAppStore(selectTheme);
  const setTheme = useAppStore((state) => state.setTheme);
  return { theme, setTheme, themes: THEMES };
}

export function useLanguageFromStore(): {
  language: LanguageCode;
  setLanguage: (language: LanguageCode) => void;
  languages: typeof LANGUAGES
} {
  const language = useAppStore(selectLanguage);
  const setLanguage = useAppStore((state) => state.setLanguage);
  return { language, setLanguage, languages: LANGUAGES };
}
