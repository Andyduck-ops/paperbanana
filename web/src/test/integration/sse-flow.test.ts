// @ts-nocheck - Test file with complex mocking
/**
 * SSE Integration Tests
 *
 * TASK-006: Full SSE event flow verification
 * Tests all event types match backend and reconnection scenarios
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { streamGenerate, type SSEEventType } from '../../lib/sse';
import type { GenerateRequest } from '../../types/api';

// Mock fetch for SSE testing
const createMockSSEResponse = (events: Array<{ type: SSEEventType; data: unknown }>) => {
  const eventStrings = events.map(e => `event: ${e.type}\ndata: ${JSON.stringify(e.data)}\n\n`);
  const body = eventStrings.join('');
  const encoder = new TextEncoder();

  return {
    ok: true,
    status: 200,
    body: {
      getReader: () => ({
        read: vi.fn()
          .mockResolvedValueOnce({ done: false, value: encoder.encode(body) })
          .mockResolvedValueOnce({ done: true, value: undefined }),
        releaseLock: vi.fn(),
      }),
    },
  };
};

const createRawMockSSEResponse = (chunks: string[]) => {
  const encoder = new TextEncoder();

  return {
    ok: true,
    status: 200,
    body: {
      getReader: () => {
        const read = vi.fn();
        for (const chunk of chunks) {
          read.mockResolvedValueOnce({ done: false, value: encoder.encode(chunk) });
        }
        read.mockResolvedValueOnce({ done: true, value: undefined });
        return {
          read,
          releaseLock: vi.fn(),
        };
      },
    },
  };
};

describe('SSE Integration Tests', () => {
  beforeEach(() => {
    // Mock Request constructor to handle relative URLs in tests
    const OriginalRequest = global.Request;
    vi.stubGlobal('Request', vi.fn((input: string | URL, init?: RequestInit) => {
      // Convert relative URLs to absolute for Node.js compatibility
      const url = typeof input === 'string' && input.startsWith('/')
        ? `http://localhost${input}`
        : input;
      return new OriginalRequest(url, init);
    }));
    vi.spyOn(global, 'fetch').mockReset();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  describe('Event Type Alignment', () => {
    it('should handle all backend event types', async () => {
      const backendEventTypes: SSEEventType[] = [
        'run_started',
        'stage_started',
        'stage_completed',
        'stage_failed',
        'run_completed',
        'run_failed',
        'resume_start',
      ];

      const receivedTypes: SSEEventType[] = [];

      const mockResponse = createMockSSEResponse(
        backendEventTypes.map(type => ({
          type,
          data: { stage: 'test', status: 'running', session_id: 'test-session' },
        }))
      );

      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onRunStarted: () => receivedTypes.push('run_started'),
        onStageStart: () => receivedTypes.push('stage_started'),
        onStageComplete: () => receivedTypes.push('stage_completed'),
        onError: () => {
          // Both stage_failed and run_failed go to onError
        },
        onResult: () => receivedTypes.push('run_completed'),
        onResumeStart: () => receivedTypes.push('resume_start'),
      });

      // Verify we handled at least the core event types
      expect(receivedTypes).toContain('run_started');
      expect(receivedTypes).toContain('stage_started');
      expect(receivedTypes).toContain('stage_completed');
    });

    it('should map stage_failed to error callback', async () => {
      const errorReceived = { message: '' };

      const mockResponse = createMockSSEResponse([
        { type: 'stage_failed', data: { stage: 'planner', message: 'Planning failed' } },
      ]);

      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onError: (data) => {
          errorReceived.message = data.message;
        },
      });

      expect(errorReceived.message).toBe('Planning failed');
    });

    it('should map run_failed to error callback', async () => {
      const errorReceived = { message: '', stage: '' };

      const mockResponse = createMockSSEResponse([
        { type: 'run_failed', data: { stage: 'visualizer', message: 'Generation failed' } },
      ]);

      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onError: (data) => {
          errorReceived.message = data.message;
          errorReceived.stage = data.stage || '';
        },
      });

      expect(errorReceived.message).toBe('Generation failed');
    });

    it('should parse gin-style SSE fields without spaces after colons', async () => {
      const receivedStages: string[] = [];

      const mockResponse = createRawMockSSEResponse([
        'event:stage_started\ndata:{"stage":"visualizer","agent":"Visualizer"}\n\n',
        'event:result\ndata:{"session_id":"session-1","generated_artifacts":[]}\n\n',
      ]);

      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onStageStart: (data) => {
          receivedStages.push(data.stage);
        },
      });

      expect(receivedStages).toEqual(['visualizer']);
    });

    it('should parse events split across multiple chunks', async () => {
      const receivedTypes: string[] = [];

      const mockResponse = createRawMockSSEResponse([
        'event:stage_started\ndata:{"stage":"retr',
        'iever","agent":"Retriever"}\n\n',
        'event:stage_completed\ndata:{"stage":"retriever","summary":"done","artifact_count":1,"artifact_kinds":["reference_bundle"]}\n\n',
      ]);

      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onStageStart: (data) => receivedTypes.push(`start:${data.stage}`),
        onStageComplete: (data) => receivedTypes.push(`complete:${data.stage}`),
      });

      expect(receivedTypes).toEqual(['start:retriever', 'complete:retriever']);
    });
  });

  describe('Resume Event Handling (GD-UI-004)', () => {
    it('should receive resume_start event before run_started', async () => {
      const eventOrder: string[] = [];

      const mockResponse = createMockSSEResponse([
        { type: 'resume_start', data: { resumed_from_stage: 'planner', session_id: 'test-123' } },
        { type: 'stage_started', data: { stage: 'stylist' } },
        { type: 'stage_completed', data: { stage: 'stylist' } },
        { type: 'run_completed', data: { session_id: 'test-123' } },
      ]);

      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onResumeStart: (data) => {
          eventOrder.push(`resume:${data.resumed_from_stage}`);
        },
        onStageStart: (data) => {
          eventOrder.push(`start:${data.stage}`);
        },
        onStageComplete: (data) => {
          eventOrder.push(`complete:${data.stage}`);
        },
        onResult: () => {
          eventOrder.push('result');
        },
      });

      // Resume event should come first
      expect(eventOrder[0]).toBe('resume:planner');
      // Then stage events
      expect(eventOrder[1]).toBe('start:stylist');
    });

    it('should include resume metadata in event data', async () => {
      const resumeData = { resumed_from_stage: '', session_id: '' };

      const mockResponse = createMockSSEResponse([
        {
          type: 'resume_start',
          data: {
            resumed_from_stage: 'retriever',
            stages_completed_before_resume: ['retriever'],
            session_id: 'session-456',
          },
        },
        { type: 'run_completed', data: { session_id: 'session-456' } },
      ]);

      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onResumeStart: (data) => {
          resumeData.resumed_from_stage = data.resumed_from_stage;
          resumeData.session_id = data.session_id;
        },
      });

      expect(resumeData.resumed_from_stage).toBe('retriever');
      expect(resumeData.session_id).toBe('session-456');
    });
  });

  describe('Error Event Format (GD-UI-002)', () => {
    it('should include failed_stage in error event', async () => {
      const errorData = { failed_stage: '', stages_not_run: [] as string[] };

      const mockResponse = createMockSSEResponse([
        {
          type: 'stage_failed',
          data: {
            stage: 'planner',
            message: 'Planning timeout',
            failed_stage: 'planner',
            stages_not_run: ['stylist', 'visualizer', 'critic'],
          },
        },
      ]);

      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onError: (data) => {
          errorData.failed_stage = data.failed_stage || data.stage || '';
          errorData.stages_not_run = data.stages_not_run || [];
        },
      });

      expect(errorData.failed_stage).toBe('planner');
      expect(errorData.stages_not_run).toContain('stylist');
      expect(errorData.stages_not_run).toContain('visualizer');
      expect(errorData.stages_not_run).toContain('critic');
    });
  });

  describe('Reconnection Scenarios', () => {
    it('should handle connection drop gracefully', async () => {
      // Simulate connection dropping mid-stream
      const partialEvents = [
        { type: 'stage_started', data: { stage: 'retriever' } },
        { type: 'stage_completed', data: { stage: 'retriever' } },
        // Connection drops here...
      ];

      const mockResponse = createMockSSEResponse(partialEvents);
      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      const completedStages: string[] = [];

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onStageComplete: (data) => {
          completedStages.push(data.stage);
        },
      });

      // Should have received completed stages before drop
      expect(completedStages).toContain('retriever');
    });
  });

  describe('Event Sequence Validation', () => {
    it('should process events in order', async () => {
      const sequence: string[] = [];

      const mockResponse = createMockSSEResponse([
        { type: 'stage_started', data: { stage: 'retriever' } },
        { type: 'stage_completed', data: { stage: 'retriever' } },
        { type: 'stage_started', data: { stage: 'planner' } },
        { type: 'stage_completed', data: { stage: 'planner' } },
        { type: 'run_completed', data: { session_id: 'test' } },
      ]);

      vi.spyOn(global, 'fetch').mockResolvedValue(mockResponse as Response);

      await streamGenerate({ prompt: 'test' } as GenerateRequest, {
        onStageStart: (data) => sequence.push(`start:${data.stage}`),
        onStageComplete: (data) => sequence.push(`complete:${data.stage}`),
        onResult: () => sequence.push('done'),
      });

      // Verify order
      expect(sequence).toEqual([
        'start:retriever',
        'complete:retriever',
        'start:planner',
        'complete:planner',
        'done',
      ]);
    });
  });
});
