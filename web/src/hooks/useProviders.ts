import { useState, useEffect, useCallback } from 'react';

export interface ModelInfo {
  id: string;
  name: string;
  max_tokens?: number;
  supports_vision?: boolean;
  enabled: boolean;
}

export interface ProviderPreset {
  type: string;
  display_name: string;
  api_host: string;
  default_models: ModelInfo[];
  supports_vision: boolean;
  docs_url: string;
  api_key_url: string;
}

export interface Provider {
  id: string;
  type: string;
  name: string;
  display_name: string;
  query_model: string;
  gen_model: string;
  base_url?: string;
  timeout: string;
  status: 'configured' | 'no_keys' | 'invalid';
  enabled: boolean;
  is_system: boolean;
  is_default: boolean;
  models?: ModelInfo[];
}

interface ProvidersResponse {
  providers: Provider[];
}

interface PresetsResponse {
  presets: ProviderPreset[];
}

export function useProviders() {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProviders = useCallback(() => {
    setLoading(true);
    fetch('/api/v1/providers')
      .then(res => {
        if (!res.ok) throw new Error('Failed to fetch providers');
        return res.json();
      })
      .then((data: ProvidersResponse) => {
        setProviders(data.providers || []);
        setError(null);
      })
      .catch(err => {
        setError(err.message);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    fetchProviders();
  }, [fetchProviders]);

  return { providers, loading, error, refetch: fetchProviders };
}

export function usePresets() {
  const [presets, setPresets] = useState<ProviderPreset[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    fetch('/api/v1/providers/presets')
      .then(res => {
        if (!res.ok) throw new Error('Failed to fetch presets');
        return res.json();
      })
      .then((data: PresetsResponse) => {
        setPresets(data.presets || []);
        setError(null);
      })
      .catch(err => {
        setError(err.message);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  return { presets, loading, error };
}

export function useProvider(name: string) {
  const [provider, setProvider] = useState<Provider | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!name) {
      setLoading(false);
      return;
    }

    setLoading(true);
    fetch(`/api/v1/providers/${name}`)
      .then(res => {
        if (!res.ok) throw new Error('Provider not found');
        return res.json();
      })
      .then(data => {
        setProvider(data.provider);
        setError(null);
      })
      .catch(err => {
        setError(err.message);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [name]);

  return { provider, loading, error };
}
