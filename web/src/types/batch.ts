export interface BatchArtifact {
  id: string;
  kind: string;
  mimeType: string;
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
