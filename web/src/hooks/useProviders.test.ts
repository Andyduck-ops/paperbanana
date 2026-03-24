import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { useProviders } from './useProviders';

class MockEventSource {
  static instances: MockEventSource[] = [];

  private listeners = new Map<string, Array<(event: MessageEvent<string>) => void>>();
  close = vi.fn();

  constructor(public url: string) {
    MockEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: (event: MessageEvent<string>) => void) {
    const existing = this.listeners.get(type) || [];
    existing.push(listener);
    this.listeners.set(type, existing);
  }

  emit(type: string, payload: unknown) {
    const listeners = this.listeners.get(type) || [];
    const event = { data: JSON.stringify(payload) } as MessageEvent<string>;
    listeners.forEach((listener) => listener(event));
  }
}

describe('useProviders', () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    MockEventSource.instances = [];
    fetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({
        providers: [
          {
            id: 'provider-1',
            type: 'openai',
            name: 'openai',
            display_name: 'OpenAI',
            query_model: 'gpt-4o',
            gen_model: 'gpt-4o',
            timeout: '30s',
            status: 'configured',
            enabled: true,
            is_system: true,
            is_default: true,
            models: [],
          },
        ],
      }),
    });

    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('EventSource', MockEventSource as unknown as typeof EventSource);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('refreshes providers when the config stream emits a change', async () => {
    const { result, unmount } = renderHook(() => useProviders());

    await waitFor(() => {
      expect(result.current.providers).toHaveLength(1);
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(MockEventSource.instances).toHaveLength(1);

    act(() => {
      MockEventSource.instances[0].emit('config_changed', {
        type: 'provider_updated',
        provider_id: 'provider-1',
      });
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    unmount();
    expect(MockEventSource.instances[0].close).toHaveBeenCalled();
  });
});
