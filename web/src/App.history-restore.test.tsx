import * as React from "react";
import type { ReactNode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";
import { useAppStore } from "./stores";

const restoreSessionMock = vi.fn();
const restoreRefineMock = vi.fn();

vi.mock("./hooks", () => ({
  useLanguage: () => ({
    t: (key: string) =>
      ({
        "history.restore": "恢复",
        "history.restoreFailed": "恢复失败",
        "app.tabGenerate": "生成",
        "app.tabRefine": "精修",
      })[key] || key,
  }),
  useGenerationFlow: () => {
    const setMainTab = useAppStore((state) => state.setMainTab);
    const [refineArtifact, setRefineArtifact] =
      React.useState<{ kind: string; mimeType: string; summary: string; data: string } | null>(null);

    const handleSelectSession = React.useCallback(
      async (sessionId: string) => {
        const restored = await restoreSessionMock(sessionId);
        if (!restored) return;
        if (restored.mode === "refine") {
          setMainTab("refine");
          const payload = {
            sessionId: restored.id,
            status: restored.status === "completed" ? "completed" : "failed",
            content: restored.prompt,
            image: {
              data: restored.artifacts[0]?.data || "",
              mimeType: restored.artifacts[0]?.mimeType || "image/png",
            },
          };
          restoreRefineMock(payload);
          setRefineArtifact({
            kind: "image",
            mimeType: payload.image.mimeType,
            summary: "Refined image",
            data: payload.image.data,
          });
        }
      },
      [setMainTab],
    );

    return {
      isGenerating: false,
      isBatchGenerating: false,
      isRefining: false,
      stages: [],
      result: null,
      batchResult: null,
      batchProgress: null,
      batchError: null,
      refineResult: refineArtifact ? { image: { data: refineArtifact.data, mimeType: refineArtifact.mimeType } } : null,
      refineArtifact,
      refineError: null,
      error: null,
      handleGenerate: vi.fn(),
      handleRefine: vi.fn(),
      handleExport: vi.fn(),
      handleSelectSession,
      resetGenerate: vi.fn(),
      resetBatch: vi.fn(),
      resetRefine: vi.fn(),
      cancel: vi.fn(),
    };
  },
  useToast: () => ({
    toasts: [],
    addToast: vi.fn(),
    removeToast: vi.fn(),
  }),
  useKeyboardShortcuts: vi.fn(),
  useHistory: () => ({
    count: 1,
    restoreSession: restoreSessionMock,
    sessions: [],
    isLoading: false,
    error: null,
    refresh: vi.fn(),
  }),
  useLocalWorkRecords: () => ({
    records: [],
    addRecord: vi.fn(),
    removeRecord: vi.fn(),
    clearRecords: vi.fn(),
  }),
  useFocusTrap: () => ({ current: null }),
}));

vi.mock("./pages/ProviderEditPage", () => ({
  ProviderEditPage: () => <div>provider edit</div>,
}));

vi.mock("./components", () => ({
  Layout: ({ header, footer, children }: { header: ReactNode; footer: ReactNode; children: ReactNode }) => (
    <div>
      {header}
      {children}
      {footer}
    </div>
  ),
  Header: ({ onHistoryClick }: { onHistoryClick: () => void }) => (
    <button type="button" onClick={onHistoryClick}>
      打开历史
    </button>
  ),
  Footer: () => <div>footer</div>,
  Toast: () => null,
  ErrorBoundary: ({ children }: { children: ReactNode }) => <>{children}</>,
  SettingsDrawer: () => null,
  GeneratePanel: () => <div>generate panel</div>,
  HistoryPanel: ({ isOpen, onSelectSession }: { isOpen: boolean; onSelectSession: (sessionId: string) => void }) =>
    isOpen ? (
      <button type="button" onClick={() => onSelectSession("refine-session-1")}>
        选择精修记录
      </button>
    ) : null,
  Workspace: ({ mode, refineResult }: { mode: string; refineResult?: { data?: string } | null }) => (
    <div data-testid="workspace-state">
      {mode}:{refineResult?.data || "none"}
    </div>
  ),
  ExportModal: () => null,
  RefinePanel: () => <div>refine panel</div>,
  ShortcutsHelpPanel: () => null,
  WelcomeWizard: () => null,
  isWizardCompleted: () => true,
}));

describe("App refine history restore", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    restoreSessionMock.mockResolvedValue({
      mode: "refine",
      id: "refine-session-1",
      projectId: "default",
      status: "completed",
      currentStage: "polish",
      prompt: "Sharpen labels",
      artifacts: [
        {
          kind: "polished_image",
          mimeType: "image/png",
          summary: "Refined image",
          data: "restored-image",
        },
      ],
    });
  });

  it("switches to refine mode and hydrates the restored artifact", async () => {
    render(<App />);

    expect(screen.getByTestId("workspace-state")).toHaveTextContent("generate:none");

    fireEvent.click(screen.getByRole("button", { name: "打开历史" }));
    fireEvent.click(screen.getByRole("button", { name: "选择精修记录" }));

    await waitFor(() => {
      expect(restoreSessionMock).toHaveBeenCalledWith("refine-session-1");
      expect(restoreRefineMock).toHaveBeenCalledWith({
        sessionId: "refine-session-1",
        status: "completed",
        content: "Sharpen labels",
        image: {
          data: "restored-image",
          mimeType: "image/png",
        },
      });
      expect(screen.getByTestId("workspace-state")).toHaveTextContent("refine:restored-image");
    });
  });
});
