import { describe, expect, it, vi, beforeEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";

import type { RefineRequest, RefineResult } from "../types/api";
import { refineImage } from "../lib/refine";
import { useRefine } from "./useRefine";

vi.mock("../lib/refine", () => ({
  refineImage: vi.fn(),
}));

describe("useRefine", () => {
  const file = new File(["image-bytes"], "refine.png", { type: "image/png" });
  const request: RefineRequest = {
    image: {
      file,
      previewUrl: "blob:refine-preview",
    },
    instructions: "Sharpen labels",
    resolution: "2K",
    enable_iteration: false,
    max_iterations: 1,
  };
  const result: RefineResult = {
    sessionId: "session-1",
    status: "completed",
    content: "done",
    metadata: { iterations: "2" },
    image: {
      data: "refined-image",
      mimeType: "image/webp",
    },
  };

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("tracks refine lifecycle and stores the final result", async () => {
    const onSuccess = vi.fn();
    let resolveRefine!: (value: RefineResult) => void;
    const pendingRefine = new Promise<RefineResult>((resolve) => {
      resolveRefine = resolve;
    });
    vi.mocked(refineImage).mockReturnValue(pendingRefine);

    const { result: hook } = renderHook(() => useRefine({ onSuccess }));

    let refinePromise!: Promise<RefineResult>;
    act(() => {
      refinePromise = hook.current.refine(request);
    });

    await waitFor(() => {
      expect(hook.current.isRefining).toBe(true);
      expect(hook.current.error).toBeNull();
    });

    await act(async () => {
      resolveRefine(result);
      await expect(refinePromise).resolves.toEqual(result);
    });

    await waitFor(() => {
      expect(hook.current.isRefining).toBe(false);
      expect(hook.current.result).toEqual(result);
    });

    expect(onSuccess).toHaveBeenCalledWith(result);
  });

  it("captures errors and preserves reset semantics", async () => {
    const onError = vi.fn();
    const failure = new Error("Refinement failed");
    vi.mocked(refineImage).mockRejectedValue(failure);

    const { result: hook } = renderHook(() => useRefine({ onError }));

    await act(async () => {
      await expect(hook.current.refine(request)).rejects.toThrow("Refinement failed");
    });

    await waitFor(() => {
      expect(hook.current.isRefining).toBe(false);
      expect(hook.current.error).toEqual(failure);
      expect(hook.current.result).toBeNull();
    });

    act(() => {
      hook.current.reset();
    });

    expect(hook.current.isRefining).toBe(false);
    expect(hook.current.error).toBeNull();
    expect(hook.current.result).toBeNull();
    expect(onError).toHaveBeenCalledWith(failure);
  });
});
