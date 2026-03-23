import type {
  RefineApiRequest,
  RefineRequest,
  RefineResponse,
  RefineResult,
} from "../types/api";
import { ApiError } from "./api";

const REFINE_ENDPOINT = "/api/v1/refine";

async function fileToDataUrl(file: File): Promise<string> {
  return await new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = () => reject(reader.error ?? new Error("Failed to read refine image file"));
    reader.readAsDataURL(file);
  });
}

async function toRefineApiRequest(request: RefineRequest): Promise<RefineApiRequest> {
  return {
    session_id: request.session_id,
    image_data: await fileToDataUrl(request.image.file),
    instructions: request.instructions,
    resolution: request.resolution,
    model: request.model,
    provider_id: request.provider_id,
    max_iterations: request.max_iterations,
    enable_iteration: request.enable_iteration,
  };
}

export async function refineImage(request: RefineRequest): Promise<RefineResult> {
  const apiRequest = await toRefineApiRequest(request);

  const response = await fetch(REFINE_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(apiRequest),
  });

  let payload: RefineResponse | undefined;
  try {
    payload = (await response.json()) as RefineResponse;
  } catch {
    if (!response.ok) {
      throw new ApiError(
        response.status,
        response.statusText,
        `HTTP ${response.status}: ${response.statusText}`
      );
    }
    throw new Error("Invalid refine response payload");
  }

  if (!response.ok) {
    throw new ApiError(
      response.status,
      response.statusText,
      payload?.error || `HTTP ${response.status}: ${response.statusText}`
    );
  }

  if (payload.error) {
    throw new Error(payload.error);
  }

  const imageData = payload.image?.data || payload.image_data;
  if (!imageData) {
    throw new Error("Refine response did not include image data");
  }

  return {
    sessionId: payload.session_id,
    status: payload.status,
    content: payload.content,
    metadata: payload.metadata,
    image: {
      data: imageData,
      mimeType: payload.image?.mime_type || "image/png",
      metadata: payload.image?.metadata,
    },
  };
}
