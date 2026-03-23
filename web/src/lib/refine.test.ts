import { describe, expect, it, vi, beforeEach } from "vitest";

import type { RefineRequest, RefineResponse } from "../types/api";
import { refineImage } from "./refine";

describe("refineImage", () => {
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

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("returns normalized final image payload", async () => {
    const payload: RefineResponse = {
      session_id: "session-1",
      status: "completed",
      content: "done",
      metadata: { iterations: "2", stop_reason: "accepted" },
      image: {
        data: "refined-image",
        mime_type: "image/webp",
      },
    };

    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify(payload), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      })
    );

    const result = await refineImage(request);

    expect(result.sessionId).toBe("session-1");
    expect(result.image.data).toBe("refined-image");
    expect(result.image.mimeType).toBe("image/webp");
    expect(result.metadata?.iterations).toBe("2");
  });

  it("falls back to deprecated image_data field when needed", async () => {
    const payload: RefineResponse = {
      session_id: "session-2",
      status: "completed",
      image_data: "legacy-image",
    };

    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify(payload), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      })
    );

    const result = await refineImage(request);

    expect(result.image.data).toBe("legacy-image");
    expect(result.image.mimeType).toBe("image/png");
  });
});
