import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useBatchGeneration } from './useBatchGeneration';

describe('useBatchGeneration', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('reduces stream events into progress and result state', async () => {
    const encoder = new TextEncoder();
    const chunks = [
      'data: {"type":"batch_start","batch_id":"batch-1","timing":{"started_at":"2026-03-18T00:00:00Z"}}\n',
      'data: {"type":"candidate_start","candidate_id":0}\n',
      'data: {"type":"candidate_complete","candidate_id":0,"status":"completed"}\n',
      'data: {"batch_id":"batch-1","results":[{"candidate_id":0,"status":"completed","artifacts":[{"id":"a1","kind":"figure","mimeType":"image/png"}]},{"candidate_id":1,"status":"failed","error":{"message":"boom"}}],"successful":1,"failed":1,"timing":{"completed_at":"2026-03-18T00:01:00Z"}}\n',
    ];

    const reader = {
      read: vi
        .fn()
        .mockResolvedValueOnce({ done: false, value: encoder.encode(chunks[0]) })
        .mockResolvedValueOnce({ done: false, value: encoder.encode(chunks[1]) })
        .mockResolvedValueOnce({ done: false, value: encoder.encode(chunks[2]) })
        .mockResolvedValueOnce({ done: false, value: encoder.encode(chunks[3]) })
        .mockResolvedValueOnce({ done: true, value: undefined }),
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      body: {
        getReader: () => reader,
      },
    } as unknown as Response);

    const { result } = renderHook(() => useBatchGeneration());

    await act(async () => {
      await result.current.startBatch('prompt', 2, { visualizerNode: 'viz-1' });
    });

    expect(globalThis.fetch).toHaveBeenCalledWith('/api/v1/generate/batch', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prompt: 'prompt',
        num_candidates: 2,
        visualizer_node: 'viz-1',
      }),
      signal: expect.any(AbortSignal),
    });

    await waitFor(() => {
      expect(result.current.isGenerating).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.progress).toMatchObject({
        batchId: 'batch-1',
        status: 'completed',
        successful: 1,
        failed: 1,
        completedAt: '2026-03-18T00:01:00Z',
      });
      expect(result.current.result).toMatchObject({
        batchId: 'batch-1',
        status: 'completed',
        successful: 1,
        failed: 1,
      });
      expect(result.current.result?.candidates[1].error).toBe('boom');
    });
  });

  it('captures request errors and can reset state', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
    } as unknown as Response);

    const { result } = renderHook(() => useBatchGeneration());

    await act(async () => {
      await result.current.startBatch('broken', 3);
    });

    await waitFor(() => {
      expect(result.current.isGenerating).toBe(false);
      expect(result.current.error).toBe('HTTP 500');
      expect(result.current.progress).toBeNull();
      expect(result.current.result).toBeNull();
    });

    act(() => {
      result.current.resetBatch();
    });

    expect(result.current).toMatchObject({
      isGenerating: false,
      progress: null,
      result: null,
      error: null,
    });
  });

  it('aborts an active stream on reset without surfacing an error', async () => {
    let resolveRead: ((value: { done: boolean; value?: Uint8Array }) => void) | null = null;
    const abortListener = vi.fn();

    globalThis.fetch = vi.fn().mockImplementation((_url, init?: RequestInit) => {
      init?.signal?.addEventListener('abort', abortListener);

      return Promise.resolve({
        ok: true,
        body: {
          getReader: () => ({
            read: vi.fn().mockImplementation(
              () =>
                new Promise<{ done: boolean; value?: Uint8Array }>((resolve) => {
                  resolveRead = resolve;
                })
            ),
          }),
        },
      } as unknown as Response);
    });

    const { result } = renderHook(() => useBatchGeneration());

    let batchPromise: Promise<void> | undefined;
    await act(async () => {
      batchPromise = result.current.startBatch('prompt', 2);
    });

    await waitFor(() => {
      expect(result.current.isGenerating).toBe(true);
    });

    act(() => {
      result.current.resetBatch();
    });

    expect(abortListener).toHaveBeenCalledTimes(1);
    expect(result.current).toMatchObject({
      isGenerating: false,
      progress: null,
      result: null,
      error: null,
    });

    await act(async () => {
      resolveRead?.({ done: true, value: undefined });
      await batchPromise;
    });

    expect(result.current).toMatchObject({
      isGenerating: false,
      progress: null,
      result: null,
      error: null,
    });
  });

  it('ignores stale stream updates after a newer batch starts', async () => {
    const encoder = new TextEncoder();
    let firstReadResolve: ((value: { done: boolean; value?: Uint8Array }) => void) | null = null;

    const firstReader = {
      read: vi.fn().mockImplementation(
        () =>
          new Promise<{ done: boolean; value?: Uint8Array }>((resolve) => {
            firstReadResolve = resolve;
          })
      ),
    };

    const secondReader = {
      read: vi
        .fn()
        .mockResolvedValueOnce({
          done: false,
          value: encoder.encode(
            'data: {"type":"batch_start","batch_id":"batch-2","timing":{"started_at":"2026-03-18T00:00:00Z"}}\n'
          ),
        })
        .mockResolvedValueOnce({
          done: false,
          value: encoder.encode(
            'data: {"batch_id":"batch-2","results":[{"candidate_id":0,"status":"completed","artifacts":[{"id":"a2","kind":"figure","mimeType":"image/png"}]}],"successful":1,"failed":0,"timing":{"completed_at":"2026-03-18T00:02:00Z"}}\n'
          ),
        })
        .mockResolvedValueOnce({ done: true, value: undefined }),
    };

    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        body: { getReader: () => firstReader },
      } as unknown as Response)
      .mockResolvedValueOnce({
        ok: true,
        body: { getReader: () => secondReader },
      } as unknown as Response);

    globalThis.fetch = fetchMock;

    const { result } = renderHook(() => useBatchGeneration());

    let firstBatchPromise: Promise<void> | undefined;
    await act(async () => {
      firstBatchPromise = result.current.startBatch('first', 1);
    });

    await waitFor(() => {
      expect(result.current.isGenerating).toBe(true);
    });

    await act(async () => {
      await result.current.startBatch('second', 1);
    });

    await act(async () => {
      firstReadResolve?.({
        done: false,
        value: encoder.encode(
          'data: {"type":"batch_start","batch_id":"batch-stale","timing":{"started_at":"2026-03-18T00:00:00Z"}}\n'
        ),
      });
      await firstBatchPromise;
    });

    await waitFor(() => {
      expect(result.current.isGenerating).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.result).toMatchObject({
        batchId: 'batch-2',
        successful: 1,
        failed: 0,
      });
    });

    expect(result.current.progress?.batchId).toBe('batch-2');
    expect(result.current.result?.batchId).toBe('batch-2');
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});
