import type { GenerateRequest } from '../types/api';

const API_BASE = '/api/v1';

// SSE event types from backend runner lifecycle
export type SSEEventType =
  | 'stage_start'
  | 'stage_complete'
  | 'result'
  | 'error';

export interface SSEEvent {
  type: SSEEventType;
  data: unknown;
}

export interface StageStartEvent {
  stage: string;
  agent: string;
}

export interface StageCompleteEvent {
  stage: string;
  summary: string;
  artifact_count: number;
  artifact_kinds: string[];
}

export interface ResultEvent {
  session_id: string;
  generated_artifacts: Array<{
    kind: string;
    mime_type: string;
    summary: string;
    data?: string;
  }>;
}

export interface ErrorEvent {
  message: string;
  stage?: string;
  error?: string;
}

export interface SSEOptions {
  onStageStart?: (data: StageStartEvent) => void;
  onStageComplete?: (data: StageCompleteEvent) => void;
  onResult?: (data: ResultEvent) => void;
  onError?: (data: ErrorEvent) => void;
  onOpen?: () => void;
  onClose?: () => void;
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
  const request = createSSERequest(data);
  const response = await fetch(request);

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

  options.onOpen?.();

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (line.startsWith('event: ')) {
          const eventType = line.slice(7).trim() as SSEEventType;
          const dataLine = lines[i + 1];
          if (dataLine?.startsWith('data: ')) {
            const eventData = JSON.parse(dataLine.slice(6));

            switch (eventType) {
              case 'stage_start':
                options.onStageStart?.(eventData as StageStartEvent);
                break;
              case 'stage_complete':
                options.onStageComplete?.(eventData as StageCompleteEvent);
                break;
              case 'result':
                options.onResult?.(eventData as ResultEvent);
                break;
              case 'error': {
                const errorData = eventData as ErrorEvent;
                options.onError?.({
                  ...errorData,
                  message: errorData.message || errorData.error || 'Unknown error',
                });
                break;
              }
            }
          }
        }
      }
    }
  } finally {
    reader.releaseLock();
    options.onClose?.();
  }
}
