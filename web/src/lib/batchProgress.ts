import type { BatchCandidate, BatchProgress } from '../types/batch';

interface BatchStartEvent {
  type?: string;
  batch_id: string;
  timing?: {
    started_at?: string;
    completed_at?: string;
  };
}

interface BatchCandidateEvent {
  type: 'candidate_start' | 'candidate_complete';
  candidate_id: number;
  status?: 'completed' | 'failed';
  error?: {
    message?: string;
  };
}

interface BatchResultCandidate {
  candidate_id: number;
  status: BatchCandidate['status'];
  artifacts?: BatchCandidate['artifacts'];
  error?: {
    message?: string;
  };
}

interface BatchResultEvent {
  batch_id: string;
  results: BatchResultCandidate[];
  successful?: number;
  failed?: number;
  timing?: {
    started_at?: string;
    completed_at?: string;
  };
}

interface BatchCompleteEvent {
  type: 'batch_complete';
  timing?: {
    completed_at?: string;
  };
}

export type BatchStreamEvent =
  | BatchStartEvent
  | BatchCandidateEvent
  | BatchResultEvent
  | BatchCompleteEvent;

function createPendingCandidates(count: number): BatchCandidate[] {
  return Array.from({ length: count }, (_, candidateId) => ({
    candidateId,
    status: 'pending',
  }));
}

function cloneCandidate(candidate: BatchCandidate): BatchCandidate {
  return {
    ...candidate,
    artifacts: candidate.artifacts ? [...candidate.artifacts] : undefined,
  };
}

function applyCandidatePatch(
  progress: BatchProgress,
  candidateId: number,
  updater: (candidate: BatchCandidate) => BatchCandidate
): BatchProgress {
  const current = progress.candidates[candidateId];
  if (!current) {
    return progress;
  }

  const nextCandidate = updater(cloneCandidate(current));
  if (nextCandidate === current) {
    return progress;
  }

  const candidates = [...progress.candidates];
  const previousStatus = current.status;
  const nextStatus = nextCandidate.status;
  let successful = progress.successful;
  let failed = progress.failed;

  if (previousStatus === 'completed') successful -= 1;
  if (previousStatus === 'failed') failed -= 1;
  if (nextStatus === 'completed') successful += 1;
  if (nextStatus === 'failed') failed += 1;

  candidates[candidateId] = nextCandidate;

  return {
    ...progress,
    candidates,
    successful,
    failed,
  };
}

export function createBatchProgress(batchId: string, numCandidates: number, startedAt?: string): BatchProgress {
  return {
    batchId,
    status: 'running',
    candidates: createPendingCandidates(numCandidates),
    successful: 0,
    failed: 0,
    startedAt: startedAt || new Date().toISOString(),
  };
}

export function reduceBatchStreamEvent(
  current: BatchProgress | null,
  event: BatchStreamEvent,
  numCandidates: number
): BatchProgress | null {
  if ('results' in event) {
    if (!current) {
      return null;
    }

    let next = current;
    for (const result of event.results) {
      next = applyCandidatePatch(next, result.candidate_id, (candidate) => ({
        ...candidate,
        status: result.status,
        artifacts: result.artifacts,
        error: result.error?.message,
      }));
    }

    return {
      ...next,
      batchId: event.batch_id,
      status: 'completed',
      startedAt: event.timing?.started_at || next.startedAt,
      completedAt: event.timing?.completed_at || next.completedAt,
      successful: event.successful ?? next.successful,
      failed: event.failed ?? next.failed,
    };
  }

  if ('candidate_id' in event) {
    if (!current) {
      return null;
    }

    if (event.type === 'candidate_start') {
      return applyCandidatePatch(current, event.candidate_id, (candidate) => ({
        ...candidate,
        status: 'running',
        error: undefined,
      }));
    }

    return applyCandidatePatch(current, event.candidate_id, (candidate) => ({
      ...candidate,
      status: event.status === 'completed' ? 'completed' : 'failed',
      error: event.error?.message,
    }));
  }

  if ('batch_id' in event) {
    return createBatchProgress(event.batch_id, numCandidates, event.timing?.started_at);
  }

  if (!current) {
    return null;
  }

  return {
    ...current,
    status: 'completed',
    completedAt: event.timing?.completed_at || new Date().toISOString(),
  };
}
