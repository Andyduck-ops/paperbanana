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
}
