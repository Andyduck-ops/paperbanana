import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useGenerate } from './useGenerate';

vi.mock('../lib/sse', () => ({
  streamGenerate: vi.fn(async (_data: unknown, options: {
    onStageStart?: (data: { stage: string; agent: string }) => void;
    onStageComplete?: (data: { stage: string; summary: string; artifact_count: number; artifact_kinds: string[] }) => void;
    onResult?: (data: { session_id: string; generated_artifacts: unknown[] }) => void;
  }) => {
    // Simulate stage events
    options.onStageStart?.({ stage: 'retriever', agent: 'Retriever' });
    options.onStageComplete?.({
      stage: 'retriever',
      summary: 'Found references',
      artifact_count: 1,
      artifact_kinds: ['reference_bundle'],
    });
    options.onResult?.({
      session_id: 'test-session',
      generated_artifacts: [],
    });
  }),
}));

describe('useGenerate', () => {
  it('starts with idle state', () => {
    const { result } = renderHook(() => useGenerate());
    expect(result.current.isGenerating).toBe(false);
    expect(result.current.stages).toHaveLength(0);
  });

  it('sets isGenerating when generate called', async () => {
    const { result } = renderHook(() => useGenerate());
    await act(async () => {
      result.current.generate('test prompt');
    });
    // Check final state after streaming completes
    expect(result.current.result).not.toBeNull();
  });

  it('resets state', async () => {
    const { result } = renderHook(() => useGenerate());
    await act(async () => {
      result.current.generate('test prompt');
    });
    act(() => {
      result.current.reset();
    });
    expect(result.current.isGenerating).toBe(false);
    expect(result.current.stages).toHaveLength(0);
  });
});
