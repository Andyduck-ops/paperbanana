import { useCallback, useEffect } from 'react';
import { useAppStore, Theme, LanguageCode, Page, MainTab } from '../stores';

// ============================================================================
// App Store Adapter Hook
// 
// This hook provides a bridge between the existing component API and
// the new Zustand store. It allows gradual migration without breaking
// existing components.
// ============================================================================

export interface AppStoreAdapter {
  // Navigation
  currentPage: Page;
  editingProvider: string | undefined;
  mainTab: MainTab;
  setCurrentPage: (page: Page) => void;
  setEditingProvider: (providerId: string | undefined) => void;
  setMainTab: (tab: MainTab) => void;
  
  // Drawers
  isHistoryPanelOpen: boolean;
  isSettingsOpen: boolean;
  openHistoryPanel: () => void;
  closeHistoryPanel: () => void;
  openSettings: () => void;
  closeSettings: () => void;
  toggleHistoryPanel: () => void;
  toggleSettings: () => void;
  
  // Modals
  isExportOpen: boolean;
  isShortcutsHelpOpen: boolean;
  isWelcomeWizardOpen: boolean;
  exportArtifact: unknown;
  openExport: (artifact?: unknown) => void;
  closeExport: () => void;
  openShortcutsHelp: () => void;
  closeShortcutsHelp: () => void;
  openWelcomeWizard: () => void;
  closeWelcomeWizard: () => void;
  
  // Theme
  theme: Theme;
  setTheme: (theme: Theme) => void;
  
  // Language
  language: LanguageCode;
  setLanguage: (language: LanguageCode) => void;
  
  // Other
  examplePrompt: string | null;
  setExamplePrompt: (prompt: string | null) => void;
}

export function useAppStoreAdapter(): AppStoreAdapter {
  // Get state from Zustand store
  const currentPage = useAppStore((state) => state.currentPage);
  const editingProvider = useAppStore((state) => state.editingProvider);
  const mainTab = useAppStore((state) => state.mainTab);
  const drawers = useAppStore((state) => state.drawers);
  const modals = useAppStore((state) => state.modals);
  const theme = useAppStore((state) => state.theme);
  const language = useAppStore((state) => state.language);
  const examplePrompt = useAppStore((state) => state.examplePrompt);
  
  // Get actions from Zustand store
  const setCurrentPage = useAppStore((state) => state.setCurrentPage);
  const setEditingProvider = useAppStore((state) => state.setEditingProvider);
  const setMainTab = useAppStore((state) => state.setMainTab);
  const openDrawer = useAppStore((state) => state.openDrawer);
  const closeDrawer = useAppStore((state) => state.closeDrawer);
  const toggleDrawer = useAppStore((state) => state.toggleDrawer);
  const openModal = useAppStore((state) => state.openModal);
  const closeModal = useAppStore((state) => state.closeModal);
  const setTheme = useAppStore((state) => state.setTheme);
  const setLanguage = useAppStore((state) => state.setLanguage);
  const setExamplePrompt = useAppStore((state) => state.setExamplePrompt);

  return {
    // Navigation
    currentPage,
    editingProvider,
    mainTab,
    setCurrentPage,
    setEditingProvider,
    setMainTab,
    
    // Drawers
    isHistoryPanelOpen: drawers.history,
    isSettingsOpen: drawers.settings,
    openHistoryPanel: useCallback(() => openDrawer('history'), [openDrawer]),
    closeHistoryPanel: useCallback(() => closeDrawer('history'), [closeDrawer]),
    openSettings: useCallback(() => openDrawer('settings'), [openDrawer]),
    closeSettings: useCallback(() => closeDrawer('settings'), [closeDrawer]),
    toggleHistoryPanel: useCallback(() => toggleDrawer('history'), [toggleDrawer]),
    toggleSettings: useCallback(() => toggleDrawer('settings'), [toggleDrawer]),
    
    // Modals
    isExportOpen: modals.export,
    isShortcutsHelpOpen: modals.shortcutsHelp,
    isWelcomeWizardOpen: modals.welcomeWizard,
    exportArtifact: undefined,
    openExport: useCallback(() => openModal('export'), [openModal]),
    closeExport: useCallback(() => closeModal('export'), [closeModal]),
    openShortcutsHelp: useCallback(() => openModal('shortcutsHelp'), [openModal]),
    closeShortcutsHelp: useCallback(() => closeModal('shortcutsHelp'), [closeModal]),
    openWelcomeWizard: useCallback(() => openModal('welcomeWizard'), [openModal]),
    closeWelcomeWizard: useCallback(() => closeModal('welcomeWizard'), [closeModal]),
    
    // Theme
    theme,
    setTheme,
    
    // Language
    language,
    setLanguage,
    
    // Other
    examplePrompt,
    setExamplePrompt,
  };
}

// ============================================================================
// Hook for initializing app state from localStorage / backend
// ============================================================================

export function useAppStoreInit() {
  const setShowWelcomeWizard = useAppStore((state) => state.setShowWelcomeWizard);

  useEffect(() => {
    // Theme is automatically applied by Zustand persist middleware
    // Language needs to be synced with i18n - handled by i18n init
    
    // Check if wizard should be shown
    const isWizardCompleted = () => {
      try {
        return localStorage.getItem('paperbanana-wizard-completed') === 'true';
      } catch {
        return false;
      }
    };
    
    if (!isWizardCompleted()) {
      setShowWelcomeWizard(true);
    }
  }, [setShowWelcomeWizard]);
}
