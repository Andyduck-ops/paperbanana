import "./themes/base.css";
import "./themes/pop-art.css";
import "./themes/pop-art-dark.css";
import "./themes/classical-chinese.css";
import "./themes/minimalist-bw.css";
import { useState } from "react";
import { Layout, Header, Footer, Toast, ErrorBoundary } from "./components";
import {
  GeneratePanel,
  ProgressPanel,
  ResultPanel,
  HistorySidebar,
  ExportModal,
  BatchProgressPanel,
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
} from "./hooks";
import { copyImageToClipboard } from "./lib/clipboard";
import { SettingsPage } from "./pages/SettingsPage";
import { ProviderEditPage } from "./pages/ProviderEditPage";

type Page = "main" | "settings" | "provider-new" | "provider-edit";
type MainTab = "generate" | "refine";

export function App() {
  const [currentPage, setCurrentPage] = useState<Page>("main");
  const [editingProvider, setEditingProvider] = useState<string>();
  const [mainTab, setMainTab] = useState<MainTab>("generate");
  const { t } = useLanguage();

  const { isGenerating, stages, result, error, generate, reset } =
    useGenerate();
  const { toasts, addToast, removeToast } = useToast();
  const [selectedSessionId, setSelectedSessionId] = useState<string>();
  const [exportArtifact, setExportArtifact] = useState<Artifact>();
  const [showExport, setShowExport] = useState(false);

  const {
    isGenerating: isBatchGenerating,
    progress: batchProgress,
    result: batchResult,
    error: batchError,
    startBatch,
    resetBatch,
  } = useBatchGeneration();
  const {
    isRefining,
    result: refineResult,
    refine,
    reset: resetRefine,
  } = useRefine({
    onSuccess: () => {
      addToast("Image refined successfully", "success");
    },
    onError: (refineError) => {
      addToast(refineError.message || "Refinement failed", "error");
    },
  });

  const handleGenerate = async (prompt: string, options?: GenerateOptions) => {
    if (options?.numCandidates && options.numCandidates > 1) {
      // Batch generation
      await handleBatchGenerate(
        prompt,
        options.numCandidates,
        options.visualizerNode,
        options.config
      );
    } else {
      // Single generation
      await generate(prompt, {
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
    config?: GenerateOptions['config']
  ) => {
    await startBatch(prompt, numCandidates, {
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

  const handleRefine = async (request: Parameters<typeof refine>[0]) => {
    await refine(request);
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

  useKeyboardShortcuts({
    onNewGeneration: () => {
      if (!isGenerating && !isBatchGenerating) {
        reset();
        resetBatch();
      }
    },
    onExport: () => {
      if (result?.artifacts.length) {
        handleExport(result.artifacts[0]);
      }
    },
    onEscape: () => {
      setShowExport(false);
    },
  });

  // Settings page
  if (currentPage === "settings") {
    return (
      <ErrorBoundary>
        <SettingsPage
          onBack={() => setCurrentPage("main")}
          onAddProvider={() => setCurrentPage("provider-new")}
          onEditProvider={(name) => {
            setEditingProvider(name);
            setCurrentPage("provider-edit");
          }}
        />
        <Toast toasts={toasts} onRemove={removeToast} />
      </ErrorBoundary>
    );
  }

  // Provider edit/new page
  if (currentPage === "provider-new" || currentPage === "provider-edit") {
    return (
      <ErrorBoundary>
        <ProviderEditPage
          providerId={editingProvider}
          isNew={currentPage === "provider-new"}
          onBack={() => {
            setEditingProvider(undefined);
            setCurrentPage("settings");
          }}
        />
        <Toast toasts={toasts} onRemove={removeToast} />
      </ErrorBoundary>
    );
  }

  // Main page
  return (
    <ErrorBoundary>
      <Layout
        header={<Header onSettingsClick={() => setCurrentPage("settings")} />}
        footer={<Footer />}
        sidebar={
          <div className="flex h-full flex-col gap-6">
            <HistorySidebar
              selectedSessionId={selectedSessionId}
              onSelectSession={setSelectedSessionId}
            />
          </div>
        }
      >
        <div className="workspace-shell">
          <section className="workspace-stage">
            {(error || batchError) && (
              <div className="workspace-alert rounded-[24px] border border-red-500/25 bg-red-500/10 p-4 text-red-700">
                {error || batchError}
              </div>
            )}

            <div className="workspace-stage__surface">
              {/* Tab Switcher */}
              <div className="flex gap-2 mb-4">
                <button
                  onClick={() => { setMainTab("generate"); reset(); resetBatch(); resetRefine(); }}
                  className={`px-4 py-2 rounded-lg font-medium transition-all ${
                    mainTab === "generate"
                      ? "bg-primary text-background"
                      : "bg-muted text-muted-foreground hover:bg-muted/80"
                  }`}
                >
                  {t('app.tabGenerate')}
                </button>
                <button
                  onClick={() => { setMainTab("refine"); reset(); resetBatch(); resetRefine(); }}
                  className={`px-4 py-2 rounded-lg font-medium transition-all ${
                    mainTab === "refine"
                      ? "bg-primary text-background"
                      : "bg-muted text-muted-foreground hover:bg-muted/80"
                  }`}
                >
                  {t('app.tabRefine')}
                </button>
              </div>

              {/* Generate Tab */}
              {mainTab === "generate" && (
                <>
                  {!isGenerating && !result && !isBatchGenerating && !batchProgress && (
                    <GeneratePanel
                      onGenerate={handleGenerate}
                      isGenerating={isGenerating || isBatchGenerating}
                      onNavigateToSettings={() => setCurrentPage("settings")}
                    />
                  )}

                  {isGenerating && stages.length > 0 && (
                    <ProgressPanel stages={stages} isVisible={isGenerating} />
                  )}

                  {result && (
                    <ResultPanel
                      sessionId={result.sessionId}
                      artifacts={result.artifacts}
                      onExport={handleExport}
                      onCopy={handleCopy}
                      onNewGeneration={() => {
                        reset();
                        resetBatch();
                      }}
                    />
                  )}

                  {(isBatchGenerating || batchProgress) && !batchResult && (
                    <BatchProgressPanel
                      batchId={batchProgress?.batchId || null}
                      initialProgress={batchProgress || undefined}
                    />
                  )}

                  {batchResult && (
                    <div className="space-y-4">
                      <div className="rounded-[24px] border border-border/70 bg-card/80 p-4 shadow-[0_18px_60px_rgba(32,24,20,0.08)] backdrop-blur-sm">
                        <h3 className="text-lg font-medium text-foreground">
                          Batch Complete
                        </h3>
                        <p className="mt-1 text-sm text-muted-foreground">
                          {batchResult.successful} successful, {batchResult.failed} failed
                        </p>
                      </div>
                      {batchResult.candidates
                        .filter((c) => c.status === "completed" && c.artifacts?.length)
                        .map((candidate) => (
                          <div key={candidate.candidateId} className="space-y-2">
                            <h4 className="font-medium text-foreground">
                              Candidate {candidate.candidateId + 1}
                            </h4>
                            {candidate.artifacts && candidate.artifacts.length > 0 && (
                              <ResultPanel
                                sessionId={`${batchResult.batchId}-${candidate.candidateId}`}
                                artifacts={candidate.artifacts.map((a) => ({
                                  kind: a.kind || "",
                                  mimeType: a.mimeType || "",
                                  summary: "",
                                  data: "",
                                }))}
                                onExport={handleExport}
                                onCopy={handleCopy}
                                onNewGeneration={() => {}}
                              />
                            )}
                          </div>
                        ))}
                      <button
                        onClick={() => {
                          reset();
                          resetBatch();
                        }}
                        className="w-full rounded-full bg-primary px-6 py-3 font-medium text-background transition-opacity hover:opacity-90"
                      >
                        New Generation
                      </button>
                    </div>
                  )}
                </>
              )}

              {mainTab === "refine" && (
                <>
                  {!refineArtifact && (
                    <RefinePanel
                      onRefine={handleRefine}
                      isRefining={isRefining}
                    />
                  )}

                  {refineArtifact && (
                    <div className="space-y-4">
                      <h3 className="text-lg font-medium text-foreground">
                        Refined Image
                      </h3>
                      <ResultPanel
                        sessionId="refine-result"
                        artifacts={[refineArtifact]}
                        onExport={handleExport}
                        onCopy={handleCopy}
                        onNewGeneration={() => {
                          resetRefine();
                        }}
                      />
                      <button
                        onClick={() => resetRefine()}
                        className="w-full rounded-full bg-primary px-6 py-3 font-medium text-background transition-opacity hover:opacity-90"
                      >
                        Refine Another
                      </button>
                    </div>
                  )}
                </>
              )}

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
