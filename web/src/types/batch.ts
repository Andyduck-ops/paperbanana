// BatchArtifact mirrors the backend Artifact struct (internal/domain/agent/types.go)
// for consistent type alignment between frontend and backend.
export interface BatchArtifact {
  id: string;
  kind: string;
  mime_type: string;
  uri: string;
  content?: string;
  data?: string; // Base64-encoded binary data (deprecated in backend, kept for compatibility)
  asset_id?: string;
  metadata?: Record<string, string>;
}

export interface BatchCandidate {
  candidateId: number;
  status: 'pending' | 'running' | 'completed' | 'failed';
  artifacts?: BatchArtifact[];
  error?: string;
}

export interface BatchProgress {
  batchId: string;
  status: 'running' | 'completed' | 'failed';
  candidates: BatchCandidate[];
  successful: number;
  failed: number;
  startedAt: string;
  completedAt?: string;
  downloadUrl?: string;
}

// UIArtifact is the frontend UI representation of an artifact with camelCase fields.
// Used by components like ArtifactPreview for display purposes.
export interface UIArtifact {
  id: string;
  kind: string;
  mimeType: string;
  summary?: string;
  data?: string;
  assetId?: string;
  projectId?: string;
  uri?: string;
}

// toUIArtifact converts a BatchArtifact (snake_case from backend) to UIArtifact (camelCase for UI).
export function toUIArtifact(artifact: BatchArtifact): UIArtifact {
  return {
    id: artifact.id,
    kind: artifact.kind,
    mimeType: artifact.mime_type,
    summary: artifact.kind, // Default summary to kind if not provided
    data: artifact.data || artifact.content,
    assetId: artifact.asset_id,
    projectId: artifact.metadata?.project_id,
    uri: artifact.uri,
  };
}

export async function downloadBatchArchive(batchId: string): Promise<void> {
  const response = await fetch(`/api/v1/batches/${batchId}/download`);
  if (!response.ok) {
    throw new Error(`Download failed: HTTP ${response.status}`);
  }
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `batch-${batchId}.zip`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
