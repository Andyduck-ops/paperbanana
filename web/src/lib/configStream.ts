export interface ConfigChangeEvent {
  type: string;
  provider_id?: string;
  key_id?: string;
  timestamp?: string;
}

export function subscribeToConfigChanges(
  onChange: (event: ConfigChangeEvent) => void
): () => void {
  if (typeof window === 'undefined' || typeof EventSource === 'undefined') {
    return () => {};
  }

  const source = new EventSource('/api/v1/config/stream');

  source.addEventListener('config_changed', (event) => {
    try {
      const payload = JSON.parse((event as MessageEvent<string>).data) as ConfigChangeEvent;
      onChange(payload);
    } catch {
      onChange({ type: 'config_changed' });
    }
  });

  return () => {
    source.close();
  };
}
