import * as React from "react";
import type { ReactNode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";

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
  useGenerate: () => ({
    isGenerating: false,
    stages: [],
    result: null,
    error: null,
    errorDetail: null,
    generate: vi.fn(),
    reset: vi.fn(),
    restore: vi.fn(),
  }),
  useBatchGeneration: () => ({
    isGenerating: false,
    progress: null,
    result: null,
    error: null,
    startBatch: vi.fn(),
    resetBatch: vi.fn(),
    restoreBatch: vi.fn(),
  }),
  useRefine: () => {
    const [result, setResult] = React.useState<{
      image: { data: string; mimeType: string };
    } | null>(null);

    return {
      isRefining: false,
      result,
      error: null,
      refine: vi.fn(),
      reset: vi.fn(() => setResult(null)),
      restore: vi.fn((value) => {
        restoreRefineMock(value);
        setResult(value);
      }),
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
