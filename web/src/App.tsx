import "./themes/tokens.css";
import "./themes/claude-light.css";
import "./themes/linear-dark.css";
import { useEffect, useRef, lazy, Suspense } from "react";
import { Layout, Header, Footer, Toast, ErrorBoundary, SettingsDrawer, WelcomeWizard, isWizardCompleted, ShortcutsHelpPanel } from "./components";
import {
  GeneratePanel,
  HistoryPanel,
  Workspace,
  ExportModal,
  RefinePanel,
  type Artifact,

} from "./components";
import {
  useGenerationFlow,
  useToast,
  useKeyboardShortcuts,
  useLanguage,
  useHistory,
  useLocalWorkRecords,
} from "./hooks";

import { getArtifactImageUrl } from "./components/ArtifactPreview";
import { copyImageToClipboard } from "./lib/clipboard";


// Lazy loaded pages for code splitting
const ProjectsPage = lazy(() => import("./pages/ProjectsPage"));
const ProviderEditPage = lazy(() => import("./pages/ProviderEditPage"));

// New Zustand store imports
import { useAppStore, useGenerationStore } from "./stores";

function toArtifactPreview(artifact: {
  kind: string;
  mimeType: string;
  summary?: string;
  data?: string;
  assetId?: string;
  projectId?: string;
  uri?: string;
}) {
  return {
    kind: artifact.kind,
    mimeType: artifact.mimeType,
    summary: artifact.summary || artifact.kind,
    data: artifact.data,
    assetId: artifact.assetId,
    projectId: artifact.projectId,
    uri: artifact.uri,
  } satisfies Artifact;
}

export function App() {
  // Migration: App UI state moved to Zustand store
  const currentPage = useAppStore((state) => state.currentPage);
  const editingProvider = useAppStore((state) => state.editingProvider);
  const mainTab = useAppStore((state) => state.mainTab);
  const drawers = useAppStore((state) => state.drawers);
  const modals = useAppStore((state) => state.modals);
  const examplePrompt = useAppStore((state) => state.examplePrompt);

  // Actions
  const setCurrentPage = useAppStore((state) => state.setCurrentPage);
  const setEditingProvider = useAppStore((state) => state.setEditingProvider);
  const setMainTab = useAppStore((state) => state.setMainTab);
  const openDrawer = useAppStore((state) => state.openDrawer);
  const closeDrawer = useAppStore((state) => state.closeDrawer);
  const openModal = useAppStore((state) => state.openModal);
  const closeModal = useAppStore((state) => state.closeModal);
  const setExamplePrompt = useAppStore((state) => state.setExamplePrompt);
  const setShowWelcomeWizard = useAppStore((state) => state.setShowWelcomeWizard);

  // Migration: Generation state moved to Zustand store
  const currentPath = useGenerationStore((state) => state.currentPath);
  const selectedSessionId = useGenerationStore((state) => state.selectedSessionId);
  const selectedBatchCandidateId = useGenerationStore((state) => state.selectedBatchCandidateId);
  const refineSeedImageData = useGenerationStore((state) => state.refineSeedImageData);
  const exportArtifact = useGenerationStore((state) => state.exportArtifact);
  const pendingHistoryContext = useGenerationStore((state) => state.pendingHistoryContext);

  // Generation actions
  const setCurrentPath = useGenerationStore((state) => state.setCurrentPath);
  const setSelectedBatchCandidateId = useGenerationStore((state) => state.setSelectedBatchCandidateId);
  const setRefineSeedImageData = useGenerationStore((state) => state.setRefineSeedImageData);
  const clearSelectedBatchCandidateId = useGenerationStore((state) => state.clearSelectedBatchCandidateId);
  const clearRefineSeedImageData = useGenerationStore((state) => state.clearRefineSeedImageData);

  const { t } = useLanguage();

  const {
    isGenerating,
    isBatchGenerating,
    isRefining,
    stages,
    result,
    batchResult,
    batchProgress,
    batchError,
    refineResult,
    refineArtifact,
    refineError,
    error,
    handleGenerate,
    handleRefine,
    handleExport,
    handleSelectSession,
    resetGenerate,
    resetBatch,
    resetRefine,
    cancel,
  } = useGenerationFlow();
  const { toasts, addToast, removeToast } = useToast();

  // Simple router based on URL path - sync with store
  useEffect(() => {
    const handlePopState = () => {
      setCurrentPath(window.location.pathname);
    };
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, [setCurrentPath]);

  const lastRecordedHistoryId = useRef<string | null>(null);

  useEffect(() => {
    if (!isWizardCompleted()) {
      setShowWelcomeWizard(true);
    }
  }, [setShowWelcomeWizard]);

  // History hook for count badge
  const { count: historyCount } = useHistory();

  // Local work records hook for fallback when server history unavailable
  const { addRecord: addLocalWorkRecord } = useLocalWorkRecords();



  useEffect(() => {
    const handleLoadExample = (event: CustomEvent<{ prompt: string }>) => {
      setExamplePrompt(event.detail.prompt);
      setMainTab("generate");
      resetGenerate();
      resetBatch();
      resetRefine();
      clearSelectedBatchCandidateId();
      clearRefineSeedImageData();
    };

    window.addEventListener('workspace:loadExample', handleLoadExample as EventListener);
    return () => {
      window.removeEventListener('workspace:loadExample', handleLoadExample as EventListener);
    };
  }, [resetGenerate, resetBatch, resetRefine, setExamplePrompt, setMainTab, clearSelectedBatchCandidateId, clearRefineSeedImageData]);





  const handleCopy = async (artifact: Artifact) => {
    if (artifact.data) {
      const success = await copyImageToClipboard(artifact.data);
      if (success) {
        addToast("Image copied to clipboard", "success");
      } else {
        addToast("Failed to copy image", "error");
      }
    }
  };



  useKeyboardShortcuts({
    onNewGeneration: () => {
      if (!isGenerating && !isBatchGenerating) {
        resetGenerate();
        resetBatch();
        resetRefine();
        clearSelectedBatchCandidateId();
        clearRefineSeedImageData();
        setMainTab("generate");
      }
    },
    onExport: () => {
      if (mainTab === "refine" && refineArtifact) {
        handleExport(refineArtifact);
        return;
      }

      if (generateWorkspaceResult?.artifacts.length) {
        handleExport(generateWorkspaceResult.artifacts[0]);
      }
    },
    onEscape: () => {
      closeModal('export');
      closeModal('shortcutsHelp');
    },
    onShowShortcuts: () => {
      openModal('shortcutsHelp');
    },
  });

  useEffect(() => {
    if (!result || pendingHistoryContext?.mode !== "generate") return;
    if (lastRecordedHistoryId.current === result.sessionId) return;

    addLocalWorkRecord({
      id: result.sessionId,
      createdAt: new Date().toISOString(),
      status: "completed",
      prompt: pendingHistoryContext.prompt,
      mode: "generate",
    });
    lastRecordedHistoryId.current = result.sessionId;
    // Show toast when generation completes
    addToast(t('history.savedToHistory') || 'Results saved to history', 'success');
  }, [pendingHistoryContext, result, addToast, t, addLocalWorkRecord]);

  useEffect(() => {
    if (!batchResult || pendingHistoryContext?.mode !== "batch") return;
    if (lastRecordedHistoryId.current === batchResult.batchId) return;

    addLocalWorkRecord({
      id: batchResult.batchId,
      createdAt: batchResult.startedAt,
      status: batchResult.status,
      prompt: pendingHistoryContext.prompt,
      mode: "batch",
      candidateSessionIds: batchResult.candidates
        .map((candidate) => `${batchResult.batchId}-${candidate.candidateId}`),
    });
    lastRecordedHistoryId.current = batchResult.batchId;
  }, [batchResult, pendingHistoryContext, addLocalWorkRecord]);

  useEffect(() => {
    if (!refineResult || pendingHistoryContext?.mode !== "refine") return;
    if (lastRecordedHistoryId.current === refineResult.sessionId) return;

    addLocalWorkRecord({
      id: refineResult.sessionId,
      createdAt: new Date().toISOString(),
      status: refineResult.status,
      prompt: pendingHistoryContext.prompt,
      mode: "refine",
    });
    lastRecordedHistoryId.current = refineResult.sessionId;
  }, [pendingHistoryContext, refineResult, addLocalWorkRecord]);

  useEffect(() => {
    if (!batchResult?.candidates.length || selectedBatchCandidateId) return;

    const preferredCandidateId =
      batchResult.candidates.find((candidate) => candidate.status === "completed")
        ?.candidateId?.toString();

    if (preferredCandidateId) {
      setSelectedBatchCandidateId(preferredCandidateId);
    }
  }, [batchResult, selectedBatchCandidateId, setSelectedBatchCandidateId]);

  const batchCandidates = (batchResult?.candidates || []).map((candidate) => {
    const candidateId = `${batchResult?.batchId || "batch"}-${candidate.candidateId}`;

    return {
      id: candidateId,
      index: candidate.candidateId,
      status: candidate.status,
      artifacts: (candidate.artifacts || []).map((artifact) =>
        toArtifactPreview({
          kind: artifact.kind,
          mimeType: artifact.mime_type,
          summary: artifact.kind,
          data: artifact.data,
          assetId: artifact.asset_id || artifact.id,
          projectId: artifact.metadata?.project_id,
          uri: artifact.uri,
        })
      ),
      error: candidate.error,
    };
  });

  const activeBatchCandidate =
    batchCandidates.find((candidate) => candidate.id === selectedBatchCandidateId) ||
    batchCandidates.find((candidate) => candidate.status === "completed") ||
    batchCandidates[0] ||
    null;

  const generateWorkspaceResult = result
    ? {
        sessionId: result.sessionId,
        artifacts: result.artifacts.map(toArtifactPreview),
      }
    : activeBatchCandidate && activeBatchCandidate.artifacts?.length
    ? {
        sessionId: activeBatchCandidate.id,
        artifacts: activeBatchCandidate.artifacts,
      }
    : null;

  const primaryGenerateArtifact = generateWorkspaceResult?.artifacts[0] || null;

  const activeWorkspaceError =
    mainTab === "refine"
      ? refineError?.message || null
      : error || batchError;

  // Render projects page for /projects route
  if (currentPath === '/projects') {
    return (
      <ErrorBoundary>
        <Suspense fallback={<div className="loading-screen">Loading...</div>}>
          <ProjectsPage />
        </Suspense>
        <Toast toasts={toasts} onRemove={removeToast} />
      </ErrorBoundary>
    );
  }

  // Provider edit/new page
  if (currentPage === "provider-new" || currentPage === "provider-edit") {
    return (
      <ErrorBoundary>
        <Suspense fallback={<div className="loading-screen">Loading...</div>}>
          <ProviderEditPage
            providerId={editingProvider}
            isNew={currentPage === "provider-new"}
            onBack={() => {
              setEditingProvider(undefined);
              setCurrentPage("main");
              openDrawer('settings');
            }}
            onSaveSuccess={() => addToast(t('settings.providerSaved') || 'Provider saved successfully', 'success')}
          />
        </Suspense>
        <Toast toasts={toasts} onRemove={removeToast} />
      </ErrorBoundary>
    );
  }

  // Main page
  return (
    <ErrorBoundary>
      {/* Welcome Wizard for first-time users */}
      {modals.welcomeWizard && (
        <WelcomeWizard
          onComplete={() => closeModal('welcomeWizard')}
          onNavigateToSettings={() => {
            closeModal('welcomeWizard');
            openDrawer('settings');
          }}
        />
      )}

      {/* History Panel - Sliding from left */}
      {drawers.history && (
        <HistoryPanel
          isOpen={drawers.history}
          onClose={() => closeDrawer('history')}
          onSelectSession={handleSelectSession}
          selectedSessionId={selectedSessionId}
        />
      )}

      {drawers.settings && (
        <SettingsDrawer
          isOpen={drawers.settings}
          onClose={() => closeDrawer('settings')}
        />
      )}

      <Layout
        header={
          <Header
            onSettingsClick={() => openDrawer('settings')}
            onHistoryClick={() => openDrawer('history')}
            isHistoryOpen={drawers.history}
            isSettingsOpen={drawers.settings}
            historyCount={historyCount}
          />
        }
        footer={<Footer />}
      >
        <div className="workspace-shell" data-main-workspace tabIndex={-1}>
          <section className="workspace-stage">
            <div className="workspace-stage__surface">
              <Workspace
                mode={mainTab}
                onModeChange={(mode) => setMainTab(mode)}
                isGenerating={isGenerating}
                stages={mainTab === "generate" ? stages : []}
                result={mainTab === "generate" ? generateWorkspaceResult : null}
                error={activeWorkspaceError}
                isBatchGenerating={isBatchGenerating}
                batchCandidates={mainTab === "generate" && batchCandidates.length > 0 ? batchCandidates : undefined}
                batchProgress={batchProgress ? {
                  batchId: batchProgress.batchId,
                  total: batchProgress.candidates.length,
                  completed: batchProgress.successful + batchProgress.failed,
                  failed: batchProgress.failed,
                } : null}
                refineResult={mainTab === "refine" ? refineArtifact : null}
                isRefining={isRefining}
                generateInput={
                  <GeneratePanel
                    onGenerate={handleGenerate}
                    isGenerating={isGenerating || isBatchGenerating}
                    onNavigateToSettings={() => openDrawer('settings')}
                    initialCaption={examplePrompt || undefined}
                  />
                }
                refineInput={
                  <RefinePanel
                    onRefine={handleRefine}
                    isRefining={isRefining}
                    initialImageData={refineSeedImageData}
                  />
                }
                onExport={handleExport}
                onCopy={handleCopy}
                onRefineResult={() => {
                  if (!primaryGenerateArtifact) {
                    addToast("Result image is not available for refine", "error");
                    return;
                  }

                  const imageUrl = getArtifactImageUrl(primaryGenerateArtifact);
                  if (!imageUrl) {
                    addToast("Result image is not available for refine", "error");
                    return;
                  }

                  setRefineSeedImageData(imageUrl);
                  setMainTab("refine");
                  resetRefine();
                }}
                onNewGeneration={() => {
                  resetGenerate();
                  resetBatch();
                  resetRefine();
                  clearSelectedBatchCandidateId();
                  clearRefineSeedImageData();
                  setMainTab("generate");
                }}
                onSelectCandidate={setSelectedBatchCandidateId}
                onRefineCandidate={(candidateId) => {
                  const candidate = batchCandidates.find((item) => item.id === candidateId);
                  const sourceArtifact = candidate?.artifacts?.[0];
                  const imageUrl = sourceArtifact ? getArtifactImageUrl(sourceArtifact) : null;

                  if (!imageUrl) {
                    addToast("Candidate image is not available for refine", "error");
                    return;
                  }

                  setSelectedBatchCandidateId(candidateId);
                  setRefineSeedImageData(imageUrl);
                  setMainTab("refine");
                  resetRefine();
                }}
                onDeleteCandidate={() => {
                  addToast("Candidate removal is not implemented yet", "info");
                }}
                onCancel={cancel}
              />
            </div>
          </section>
        </div>

        <ExportModal
          isOpen={modals.export}
          onClose={() => closeModal('export')}
          imageData={exportArtifact?.data}
        />

        <Toast toasts={toasts} onRemove={removeToast} />

        <ShortcutsHelpPanel
          isOpen={modals.shortcutsHelp}
          onClose={() => closeModal('shortcutsHelp')}
        />
      </Layout>
    </ErrorBoundary>
  );
}
