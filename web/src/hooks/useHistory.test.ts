// @ts-nocheck - Test file with simplified mock data
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useHistory, type RestoredSession } from "./useHistory";
import { apiClient } from "../lib/api";

vi.mock("../lib/api", () => ({
  apiClient: {
    listHistory: vi.fn(),
    getSession: vi.fn(),
  },
}));

describe("useHistory", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    vi.mocked(apiClient.listHistory).mockResolvedValue({
      sessions: [],
    });
  });

  it("restores refine sessions from final_output artifacts", async () => {
    // @ts-expect-error - Mock data simplified for testing
    vi.mocked(apiClient.getSession).mockResolvedValue({
      id: "refine-session-1",
      project_id: "default",
      status: "completed",
      current_stage: "polish",
      schema_version: "agent-session/v1",
      created_at: "2026-03-25T00:00:00Z",
      updated_at: "2026-03-25T00:00:10Z",
      snapshot: {
        initial_input: {
          content: "Sharpen labels",
        },
        stage_states: [
          {
            stage: "polish",
            status: "completed",
          },
        ],
        final_output: {
          generated_artifacts: [
            {
              kind: "polished_image",
              mime_type: "image/png",
              bytes: "refined-final",
              metadata: { summary: "refined image artifact" },
            },
          ],
        },
      },
    });

    const { result } = renderHook(() => useHistory());

    await waitFor(() => {
      expect(apiClient.listHistory).toHaveBeenCalled();
    });

    let restored: RestoredSession | null = null;
    await act(async () => {
      restored = await result.current.restoreSession("refine-session-1");
    });

    expect(restored).toMatchObject({
      mode: "refine",
      id: "refine-session-1",
      projectId: "default",
      visualizationId: undefined,
      status: "completed",
      currentStage: "polish",
      prompt: "Sharpen labels",
      artifacts: [
        {
          kind: "polished_image",
          mimeType: "image/png",
          summary: "refined image artifact",
          data: "refined-final",
          assetId: undefined,
        },
      ],
      error: undefined,
    });
  });

  it("falls back to the last polish-stage artifact when final_output is empty", async () => {
    // @ts-expect-error - Mock data simplified for testing
    vi.mocked(apiClient.getSession).mockResolvedValue({
      id: "refine-session-2",
      project_id: "default",
      status: "completed",
      current_stage: "polish",
      schema_version: "agent-session/v1",
      created_at: "2026-03-25T00:00:00Z",
      updated_at: "2026-03-25T00:00:10Z",
      snapshot: {
        initial_input: {
          content: "Reduce clutter",
        },
        stage_states: [
          {
            stage: "polish",
            status: "completed",
            output: {
              generated_artifacts: [
                {
                  kind: "polished_image",
                  mime_type: "image/webp",
                  bytes: "stage-fallback",
                  metadata: { summary: "last polish artifact" },
                },
              ],
            },
          },
        ],
        final_output: {},
      },
    });

    const { result } = renderHook(() => useHistory());

    await waitFor(() => {
      expect(apiClient.listHistory).toHaveBeenCalled();
    });

    let restored: RestoredSession | null = null;
    await act(async () => {
      restored = await result.current.restoreSession("refine-session-2");
    });

    expect(restored).not.toBeNull();
    if (!restored) {
      throw new Error("expected refine restore");
    }
    const restoredRefine = restored as Extract<RestoredSession, { mode: "refine" }>;
    expect(restoredRefine.artifacts[0]).toEqual({
      kind: "polished_image",
      mimeType: "image/webp",
      summary: "last polish artifact",
      data: "stage-fallback",
      assetId: undefined,
    });
  });

  it("restores batch sessions from server history metadata instead of local storage", async () => {
    // @ts-expect-error - Mock data simplified for testing
    vi.mocked(apiClient.listHistory).mockResolvedValue({
      sessions: [
        {
          id: "batch-root-001",
          project_id: "default",
          status: "completed",
          current_stage: "batch",
          schema_version: "agent-session/v1",
          created_at: "2026-03-25T00:00:00Z",
          updated_at: "2026-03-25T00:00:10Z",
          completed_at: "2026-03-25T00:00:12Z",
          prompt: "Generate three figure variants",
          summary: "Generate three figure variants",
          mode: "batch",
          batch_id: "batch-root-001",
          candidate_session_ids: ["batch-root-001-candidate-0", "batch-root-001-candidate-1"],
        },
      ],
    });
    // @ts-expect-error - Mock data simplified for testing
    vi.mocked(apiClient.getSession)
      .mockResolvedValueOnce({
        id: "batch-root-001",
        project_id: "default",
        status: "completed",
        current_stage: "batch",
        schema_version: "agent-session/v1",
        created_at: "2026-03-25T00:00:00Z",
        updated_at: "2026-03-25T00:00:10Z",
        completed_at: "2026-03-25T00:00:12Z",
        snapshot: {
          initial_input: {
            content: "Generate three figure variants",
          },
          metadata: {
            "batch.group_id": "batch-root-001",
            "batch.session_ids": "batch-root-001-candidate-0,batch-root-001-candidate-1",
            "history.mode": "batch",
          },
          stage_states: [],
          final_output: {},
          started_at: "2026-03-25T00:00:00Z",
          completed_at: "2026-03-25T00:00:12Z",
        },
      })
      .mockResolvedValueOnce({
        id: "batch-root-001-candidate-0",
        project_id: "default",
        status: "completed",
        current_stage: "critic",
        schema_version: "agent-session/v1",
        created_at: "2026-03-25T00:00:00Z",
        updated_at: "2026-03-25T00:00:10Z",
        snapshot: {
          initial_input: {
            content: "Generate three figure variants",
          },
          stage_states: [
            {
              stage: "visualizer",
              status: "completed",
              output: {
                generated_artifacts: [
                  {
                    kind: "rendered_figure",
                    mime_type: "image/png",
                    bytes: "candidate-0-image",
                    metadata: { summary: "candidate 0" },
                  },
                ],
              },
            },
          ],
          final_output: {
            generated_artifacts: [
              {
                kind: "rendered_figure",
                mime_type: "image/png",
                bytes: "candidate-0-image",
                metadata: { summary: "candidate 0" },
              },
            ],
          },
        },
      // @ts-expect-error - Mock data simplified for testing
      })
      .mockResolvedValueOnce({
        id: "batch-root-001-candidate-1",
        project_id: "default",
        status: "failed",
        current_stage: "visualizer",
        schema_version: "agent-session/v1",
        created_at: "2026-03-25T00:00:00Z",
        updated_at: "2026-03-25T00:00:10Z",
        snapshot: {
          initial_input: {
            content: "Generate three figure variants",
          },
          stage_states: [
            {
              stage: "visualizer",
              status: "failed",
              error: {
                message: "upstream error",
              },
            },
          ],
          final_output: {},
          error: {
            message: "upstream error",
          },
        },
      });

    const { result } = renderHook(() => useHistory());

    await waitFor(() => {
      expect(apiClient.listHistory).toHaveBeenCalled();
      expect(result.current.sessions[0]?.mode).toBe("batch");
    });

    let restored: RestoredSession | null = null;
    await act(async () => {
      restored = await result.current.restoreSession("batch-root-001");
    });

    expect(restored).toMatchObject({
      mode: "batch",
      id: "batch-root-001",
      projectId: "default",
      visualizationId: undefined,
      status: "completed",
      currentStage: "batch",
      prompt: "Generate three figure variants",
      batchId: "batch-root-001",
      startedAt: "2026-03-25T00:00:00Z",
      completedAt: "2026-03-25T00:00:12Z",
      successful: 1,
      failed: 1,
      candidates: [
        {
          candidateId: 0,
          sessionId: "batch-root-001-candidate-0",
          status: "completed",
          artifacts: [
            {
              kind: "rendered_figure",
              mimeType: "image/png",
              summary: "candidate 0",
              data: "candidate-0-image",
              assetId: undefined,
            },
          ],
          error: undefined,
        },
        {
          candidateId: 1,
          sessionId: "batch-root-001-candidate-1",
          status: "failed",
          artifacts: [],
          error: "upstream error",
        },
      ],
    });
  });
});
