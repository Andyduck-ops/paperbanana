import type { GenerateRequest } from '../types/api';

const API_BASE = '/api/v1';

// SSE event types from backend runner lifecycle
// Must match internal/domain/agent/events.go EventType constants
export type SSEEventType =
  | 'run_started'
  | 'stage_started'
  | 'stage_completed'
  | 'stage_failed'
  | 'run_completed'
  | 'run_failed'
  | 'run_canceled'
  | 'result'
  | 'error'
  | 'resume_start'
  | 'batch_start'
  | 'candidate_start'
  | 'candidate_complete'
  | 'batch_complete';

export interface SSEEvent {
  type: SSEEventType;
  data: unknown;
}

// Backend Event structure from internal/domain/agent/events.go
// Fields are sent as-is from backend, with metadata containing additional info
export interface RunStartedEvent {
  sequence: number;
  session_id: string;
  request_id: string;
  type: 'run_started';
  status: string;
  occurred_at: string;
  metadata?: Record<string, string>;
}

export interface StageStartEvent {
  sequence: number;
  session_id: string;
  stage: string;
  type: 'stage_started';
  status: string;
  occurred_at: string;
  timing?: { started_at?: string };
  metadata?: Record<string, string> & { agent?: string };
}

export interface StageCompleteEvent {
  sequence: number;
  session_id: string;
  stage: string;
  type: 'stage_completed';
  status: string;
  occurred_at: string;
  timing?: { started_at?: string; completed_at?: string };
  metadata?: {
    summary?: string;
    artifact_count?: string;
    artifact_kinds?: string;
  };
}

export interface ResultEvent {
  session_id: string;
  project_id?: string;
  generated_artifacts: Array<{
    kind: string;
    mime_type: string;
    summary: string;
    data?: string;
    asset_id?: string;
    project_id?: string;
  }>;
}

// GD-UI-002: Error event includes failed stage and stages that won't run
export interface ErrorEvent {
  message: string;
  stage?: string;
  error?: string;
  failed_stage?: string;
  stages_not_run?: string[];
}

// GD-UI-004: Resume metadata for resumed tasks
export interface ResumeStartEvent {
  resumed_from_stage: string;
  stages_completed_before_resume: string[];
  session_id: string;
}

export interface SSEOptions {
  signal?: AbortSignal;
  onRunStarted?: (data: RunStartedEvent) => void;
  onStageStart?: (data: StageStartEvent) => void;
  onStageComplete?: (data: StageCompleteEvent) => void;
  onResult?: (data: ResultEvent) => void;
  onError?: (data: ErrorEvent) => void;
  onResumeStart?: (data: ResumeStartEvent) => void;
  onRunCanceled?: (data: { session_id: string; stage?: string }) => void;
  onOpen?: () => void;
  onClose?: () => void;
}

function stripSSEFieldPrefix(line: string, field: 'event' | 'data'): string | null {
  if (!line.startsWith(`${field}:`)) {
    return null;
  }

  const rawValue = line.slice(field.length + 1);
  return rawValue.startsWith(' ') ? rawValue.slice(1) : rawValue;
}

function dispatchEvent(
  eventType: SSEEventType,
  payload: unknown,
  options: SSEOptions
) {
  switch (eventType) {
    case 'run_started':
      options.onRunStarted?.(payload as RunStartedEvent);
      break;
    case 'stage_started':
      options.onStageStart?.(payload as StageStartEvent);
      break;
    case 'stage_completed':
      options.onStageComplete?.(payload as StageCompleteEvent);
      break;
    case 'stage_failed': {
      const errorData = payload as ErrorEvent & { stage?: string; error?: string };
      options.onError?.({
        message: errorData.message || errorData.error || 'Stage failed',
        stage: errorData.stage,
        error: errorData.error,
        failed_stage: errorData.failed_stage || errorData.stage,
        stages_not_run: errorData.stages_not_run,
      });
      break;
    }
    case 'run_completed':
    case 'result':
      options.onResult?.(payload as ResultEvent);
      break;
    case 'run_failed':
    case 'error': {
      const errorData = payload as ErrorEvent;
      options.onError?.({
        ...errorData,
        message: errorData.message || errorData.error || 'Unknown error',
      });
      break;
    }
    case 'run_canceled':
      options.onRunCanceled?.(payload as { session_id: string; stage?: string });
      break;
    case 'resume_start':
      options.onResumeStart?.(payload as ResumeStartEvent);
      break;
    // Batch events - currently not handled with specific callbacks
    // but recognized as valid event types
    case 'batch_start':
    case 'candidate_start':
    case 'candidate_complete':
    case 'batch_complete':
      // These events are recognized but don't have specific handlers yet
      break;
  }
}

function parseEventChunk(chunk: string, options: SSEOptions) {
  const lines = chunk.split('\n');
  let eventType: SSEEventType | null = null;
  const dataLines: string[] = [];

  for (const rawLine of lines) {
    const line = rawLine.replace(/\r$/, '');
    if (!line || line.startsWith(':')) {
      continue;
    }

    const parsedEventType = stripSSEFieldPrefix(line, 'event');
    if (parsedEventType) {
      eventType = parsedEventType.trim() as SSEEventType;
      continue;
    }

    const parsedData = stripSSEFieldPrefix(line, 'data');
    if (parsedData !== null) {
      dataLines.push(parsedData);
    }
  }

  if (!eventType || dataLines.length === 0) {
    return;
  }

  try {
    const payload = JSON.parse(dataLines.join('\n'));
    dispatchEvent(eventType, payload, options);
  } catch {
    // Skip malformed payloads and keep the stream alive.
  }
}

export function createSSERequest(data: GenerateRequest): Request {
  return new Request(`${API_BASE}/generate/stream`, {
    method: 'POST',
    headers: {
      'Accept': 'text/event-stream',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
}

export async function streamGenerate(
  data: GenerateRequest,
  options: SSEOptions = {}
): Promise<void> {
  const { signal, ...sseCallbacks } = options;
  const request = createSSERequest(data);
  const response = await fetch(request, { signal });

  if (!response.ok) {
    const payload = await response.json().catch(() => null) as { error?: string } | null;
    throw new Error(payload?.error || `HTTP ${response.status}: ${response.statusText}`);
  }

  const reader = response.body?.getReader();
  if (!reader) {
    throw new Error('No response body');
  }

  const decoder = new TextDecoder();
  let buffer = '';

  sseCallbacks.onOpen?.();

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const chunks = buffer.split('\n\n');
      buffer = chunks.pop() || '';

      for (const chunk of chunks) {
        parseEventChunk(chunk, sseCallbacks);
      }
    }

    if (buffer.trim()) {
      parseEventChunk(buffer, sseCallbacks);
    }
  } finally {
    reader.releaseLock();
    sseCallbacks.onClose?.();
  }
}
