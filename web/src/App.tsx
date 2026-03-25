import "./themes/base.css";
import "./themes/qi-baishi.css";
import "./themes/pop-anime.css";
import "./themes/rococo.css";
import "./themes/japanese-bw.css";
import "./themes/workspace.css";
import { useState, useCallback, useEffect, useRef } from "react";
import { Layout, Header, Footer, Toast, ErrorBoundary, SettingsDrawer } from "./components";
import {
  GeneratePanel,
  HistoryPanel,
  Workspace,
  ExportModal,
  RefinePanel,
  type Artifact,
  type GenerateOptions,
} from "./components";
import {
  useBatchGeneration,
  useGenerate,
  useRefine,
  useToast,
  useKeyboardShortcuts,
  useLanguage,
  useHistory,
} from "./hooks";
import { getArtifactImageUrl } from "./components/ArtifactPreview";
import { copyImageToClipboard } from "./lib/clipboard";
import { ProviderEditPage } from "./pages/ProviderEditPage";

type Page = "main" | "provider-new" | "provider-edit";
type MainTab = "generate" | "refine";
type LocalWorkMode = "generate" | "batch" | "refine";

const LOCAL_WORK_RECORDS_KEY = "paperbanana-local-work-records";

interface LocalWorkRecord {
  id: string;
  createdAt: string;
  status: string;
  prompt?: string;
  mode: LocalWorkMode;
  candidateSessionIds?: string[];
}

function recordLocalWorkEntry(entry: LocalWorkRecord) {
  if (typeof window === "undefined") return;

  const existing = JSON.parse(
    localStorage.getItem(LOCAL_WORK_RECORDS_KEY) || "[]"
  ) as LocalWorkRecord[];

  const next = [entry, ...existing.filter((item) => item.id !== entry.id)].slice(0, 24);
  localStorage.setItem(LOCAL_WORK_RECORDS_KEY, JSON.stringify(next));
}

async function imageSourceToFile(imageSource: string, filename: string) {
  if (imageSource.startsWith("data:")) {
    const [header, base64 = ""] = imageSource.split(",");
    const mimeType = header.match(/data:(.*?);base64/)?.[1] || "image/png";
    const binary = window.atob(base64);
    const bytes = new Uint8Array(binary.length);

    for (let i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i);
    }

    return new File([bytes], filename, { type: mimeType });
  }

  const response = await fetch(imageSource);
  if (!response.ok) {
    throw new Error(`Failed to load refine source: HTTP ${response.status}`);
  }

  const blob = await response.blob();
  return new File([blob], filename, { type: blob.type || "image/png" });
}

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
  const [currentPage, setCurrentPage] = useState<Page>("main");
  const [editingProvider, setEditingProvider] = useState<string>();
  const [mainTab, setMainTab] = useState<MainTab>("generate");
  const { t } = useLanguage();

  const { isGenerating, stages, result, error, generate, reset, restore: restoreGenerate } =
    useGenerate();
  const { toasts, addToast, removeToast } = useToast();
  const [selectedSessionId, setSelectedSessionId] = useState<string>();
  const [selectedBatchCandidateId, setSelectedBatchCandidateId] = useState<string | null>(null);
  const [refineSeedImageData, setRefineSeedImageData] = useState<string | null>(null);
  const [exportArtifact, setExportArtifact] = useState<Artifact>();
  const [showExport, setShowExport] = useState(false);
  const [isHistoryPanelOpen, setIsHistoryPanelOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [pendingHistoryContext, setPendingHistoryContext] = useState<{
    prompt: string;
    mode: LocalWorkMode;
  } | null>(null);
  const lastRecordedHistoryId = useRef<string | null>(null);

  // History hook for count badge
  const { count: historyCount, restoreSession: restoreHistorySession } = useHistory();

  const {
    isGenerating: isBatchGenerating,
    progress: batchProgress,
    result: batchResult,
    error: batchError,
    startBatch,
    resetBatch,
    restoreBatch,
  } = useBatchGeneration();
  const {
    isRefining,
    result: refineResult,
    error: refineError,
    refine,
    reset: resetRefine,
    restore: restoreRefine,
  } = useRefine({
    onSuccess: () => {
      addToast("Image refined successfully", "success");
    },
    onError: (refineError) => {
      addToast(refineError.message || "Refinement failed", "error");
    },
  });

  const handleGenerate = async (prompt: string, options?: GenerateOptions) => {
    setMainTab("generate");
    setRefineSeedImageData(null);
    resetRefine();

    setPendingHistoryContext({
      prompt,
      mode: options?.numCandidates && options.numCandidates > 1 ? "batch" : "generate",
    });

    if (options?.numCandidates && options.numCandidates > 1) {
      reset();
      setSelectedBatchCandidateId(null);
      // Batch generation
      await handleBatchGenerate(
        prompt,
        options.numCandidates,
        options.visualizerNode,
        options.config,
        options.content,
        options.visualIntent
      );
    } else {
      resetBatch();
      setSelectedBatchCandidateId(null);
      // Single generation
      await generate(prompt, {
        content: options?.content,
        visualIntent: options?.visualIntent,
        visualizerNode: options?.visualizerNode,
        config: options?.config
          ? {
              aspect_ratio: options.config.aspectRatio,
              critic_rounds: options.config.criticRounds,
              retrieval_mode: options.config.retrievalMode,
              pipeline_mode: options.config.pipelineMode,
              query_model: options.config.queryModel,
              gen_model: options.config.genModel,
            }
          : undefined,
      });
    }
  };

  const handleBatchGenerate = async (
    prompt: string,
    numCandidates: number,
    visualizerNode?: string,
    config?: GenerateOptions['config'],
    content?: string,
    visualIntent?: string
  ) => {
    setSelectedBatchCandidateId(null);
    await startBatch(prompt, numCandidates, {
      content,
      visualIntent,
      visualizerNode,
      config: config
        ? {
            aspect_ratio: config.aspectRatio,
            critic_rounds: config.criticRounds,
            retrieval_mode: config.retrievalMode,
            pipeline_mode: config.pipelineMode,
            query_model: config.queryModel,
            gen_model: config.genModel,
          }
        : undefined,
    });
  };

  const handleRefine = async (request: {
    imageData: string;
    instructions: string;
    resolution: "2K" | "4K";
    enableIteration?: boolean;
    maxIterations?: number;
  }) => {
    setPendingHistoryContext({
      prompt: request.instructions || "Refinement task",
      mode: "refine",
    });
    await refine({
      image: {
        file: await imageSourceToFile(request.imageData, "refine-input.png"),
        previewUrl: request.imageData,
      },
      instructions: request.instructions,
      resolution: request.resolution,
      enable_iteration: request.enableIteration,
      max_iterations: request.enableIteration ? request.maxIterations : 1,
    });
  };

  const refineArtifact = refineResult
    ? {
        kind: "image" as const,
        mimeType: refineResult.image.mimeType,
        summary: "Refined image",
        data: refineResult.image.data,
      }
    : null;

  const handleExport = (artifact: Artifact) => {
    setExportArtifact(artifact);
    setShowExport(true);
  };

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

  const handleSelectSession = useCallback(async (sessionId: string) => {
    setSelectedSessionId(sessionId);
    addToast(t('history.restore') + '...', "info");

    const restored = await restoreHistorySession(sessionId);
    if (!restored) {
      addToast(t('history.restoreFailed') || "Failed to restore session", "error");
      return;
    }

    setPendingHistoryContext(null);

    if (restored.mode === "batch" && restored.candidates) {
      reset();
      resetRefine();
      setRefineSeedImageData(null);
      setMainTab("generate");
      restoreBatch({
        batchId: restored.batchId,
        status: "completed",
        candidates: restored.candidates.map((candidate) => ({
          candidateId: candidate.candidateId,
          sessionId: candidate.sessionId,
          status: candidate.status === "completed" ? "completed" : "failed",
          artifacts: candidate.artifacts.map((artifact) => ({
            id: artifact.assetId || `${candidate.sessionId}-${artifact.kind}`,
            kind: artifact.kind,
            mimeType: artifact.mimeType,
            summary: artifact.summary || artifact.kind,
            data: artifact.data,
            assetId: artifact.assetId,
          })),
          error: candidate.error,
        })),
        successful: restored.successful,
        failed: restored.failed,
        startedAt: restored.startedAt,
        completedAt: restored.completedAt,
      });
      setSelectedBatchCandidateId(
        restored.candidates.find((candidate) => candidate.status === "completed")?.sessionId ||
          restored.candidates[0]?.sessionId ||
          null
      );
      addToast(`${t('history.restore')}: ${sessionId.slice(0, 8)}...`, "success");
      return;
    }

    reset();
    resetBatch();
    setSelectedBatchCandidateId(null);

    if (restored.mode === "refine") {
      setMainTab("refine");
      setRefineSeedImageData(null);
      restoreRefine({
        sessionId: restored.id,
        status: restored.status === "completed" ? "completed" : "failed",
        content: restored.prompt,
        image: {
          data: restored.artifacts[0]?.data || "",
          mimeType: restored.artifacts[0]?.mimeType || "image/png",
        },
      });
      addToast(`${t('history.restore')}: ${sessionId.slice(0, 8)}...`, "success");
      return;
    }

    setMainTab("generate");
    resetRefine();
    setRefineSeedImageData(null);
    restoreGenerate({
      sessionId: restored.id,
      artifacts: restored.artifacts.map(toArtifactPreview),
      stages: restored.stages || [],
      error: restored.error,
      resumeMetadata: restored.resumeMetadata,
    });
    addToast(`${t('history.restore')}: ${sessionId.slice(0, 8)}...`, "success");
  }, [addToast, reset, resetBatch, resetRefine, restoreBatch, restoreGenerate, restoreHistorySession, restoreRefine, t]);

  useKeyboardShortcuts({
    onNewGeneration: () => {
      if (!isGenerating && !isBatchGenerating) {
        reset();
        resetBatch();
        resetRefine();
        setSelectedBatchCandidateId(null);
        setRefineSeedImageData(null);
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
      setShowExport(false);
    },
  });

  useEffect(() => {
    if (!result || pendingHistoryContext?.mode !== "generate") return;
    if (lastRecordedHistoryId.current === result.sessionId) return;

    recordLocalWorkEntry({
      id: result.sessionId,
      createdAt: new Date().toISOString(),
      status: "completed",
      prompt: pendingHistoryContext.prompt,
      mode: "generate",
    });
    lastRecordedHistoryId.current = result.sessionId;
  }, [pendingHistoryContext, result]);

  useEffect(() => {
    if (!batchResult || pendingHistoryContext?.mode !== "batch") return;
    if (lastRecordedHistoryId.current === batchResult.batchId) return;

    recordLocalWorkEntry({
      id: batchResult.batchId,
      createdAt: batchResult.startedAt,
      status: batchResult.status,
      prompt: pendingHistoryContext.prompt,
      mode: "batch",
      candidateSessionIds: batchResult.candidates
        .map((candidate) => candidate.sessionId)
        .filter((value): value is string => Boolean(value)),
    });
    lastRecordedHistoryId.current = batchResult.batchId;
  }, [batchResult, pendingHistoryContext]);

  useEffect(() => {
    if (!refineResult || pendingHistoryContext?.mode !== "refine") return;
    if (lastRecordedHistoryId.current === refineResult.sessionId) return;

    recordLocalWorkEntry({
      id: refineResult.sessionId,
      createdAt: new Date().toISOString(),
      status: refineResult.status,
      prompt: pendingHistoryContext.prompt,
      mode: "refine",
    });
    lastRecordedHistoryId.current = refineResult.sessionId;
  }, [pendingHistoryContext, refineResult]);

  useEffect(() => {
    if (!batchResult?.candidates.length || selectedBatchCandidateId) return;

    const preferredCandidateId =
      batchResult.candidates.find((candidate) => candidate.status === "completed")?.sessionId ||
      batchResult.candidates.find((candidate) => candidate.status === "completed")
        ?.candidateId?.toString();

    if (preferredCandidateId) {
      setSelectedBatchCandidateId(preferredCandidateId);
    }
  }, [batchResult, selectedBatchCandidateId]);

  const batchCandidates = (batchResult?.candidates || []).map((candidate) => {
    const candidateId = candidate.sessionId || `${batchResult?.batchId || "batch"}-${candidate.candidateId}`;

    return {
      id: candidateId,
      index: candidate.candidateId,
      status: candidate.status,
      artifacts: (candidate.artifacts || []).map((artifact) =>
        toArtifactPreview({
          kind: artifact.kind,
          mimeType: artifact.mimeType,
          summary: artifact.summary,
          data: artifact.data,
          assetId: artifact.assetId || artifact.id,
          projectId: artifact.projectId,
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

  // Provider edit/new page
  if (currentPage === "provider-new" || currentPage === "provider-edit") {
    return (
      <ErrorBoundary>
        <ProviderEditPage
          providerId={editingProvider}
          isNew={currentPage === "provider-new"}
          onBack={() => {
            setEditingProvider(undefined);
            setCurrentPage("main");
            setIsSettingsOpen(true);
          }}
        />
        <Toast toasts={toasts} onRemove={removeToast} />
      </ErrorBoundary>
    );
  }

  // Main page
  return (
    <ErrorBoundary>
      {/* History Panel - Sliding from left */}
      {isHistoryPanelOpen && (
        <HistoryPanel
          isOpen={isHistoryPanelOpen}
          onClose={() => setIsHistoryPanelOpen(false)}
          onSelectSession={handleSelectSession}
          selectedSessionId={selectedSessionId}
        />
      )}

      {isSettingsOpen && (
        <SettingsDrawer
          isOpen={isSettingsOpen}
          onClose={() => setIsSettingsOpen(false)}
        />
      )}

      <Layout
        header={
          <Header
            onSettingsClick={() => setIsSettingsOpen(true)}
            onHistoryClick={() => setIsHistoryPanelOpen(true)}
            isHistoryOpen={isHistoryPanelOpen}
            isSettingsOpen={isSettingsOpen}
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
                    onNavigateToSettings={() => setIsSettingsOpen(true)}
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
                  reset();
                  resetBatch();
                  resetRefine();
                  setSelectedBatchCandidateId(null);
                  setRefineSeedImageData(null);
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
              />
            </div>
          </section>
        </div>

        <ExportModal
          isOpen={showExport}
          onClose={() => setShowExport(false)}
          imageData={exportArtifact?.data}
        />

        <Toast toasts={toasts} onRemove={removeToast} />
      </Layout>
    </ErrorBoundary>
  );
}
