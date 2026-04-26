import type { Provider } from '../stores';

export interface DefaultConfigResult {
  success: boolean;
  provider?: Provider;
  error?: string;
}

const DEFAULT_OPENAI_PROVIDER = {
  type: 'openai' as const,
  name: 'openai',
  display_name: 'OpenAI',
  enabled: true,
};

const DEFAULT_ANTHROPIC_PROVIDER = {
  type: 'anthropic' as const,
  name: 'anthropic',
  display_name: 'Anthropic',
  enabled: true,
};

export async function ensureDefaultConfig(): Promise<DefaultConfigResult> {
  try {
    const response = await fetch('/api/v1/providers');
    if (!response.ok) {
      throw new Error('Failed to fetch providers');
    }
    const data = await response.json();
    const providers = data.providers || [];

    const hasConfiguredProvider = providers.some(
      (p: { status: string }) => p.status === 'configured'
    );

    if (hasConfiguredProvider) {
      return { success: true };
    }

    const createResponse = await fetch('/api/v1/providers', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(DEFAULT_OPENAI_PROVIDER),
    });

    if (!createResponse.ok) {
      const anthropicResponse = await fetch('/api/v1/providers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(DEFAULT_ANTHROPIC_PROVIDER),
      });

      if (!anthropicResponse.ok) {
        throw new Error('Failed to create default provider');
      }

      const anthropicData = await anthropicResponse.json();
      return {
        success: true,
        provider: anthropicData.provider,
      };
    }

    const createData = await createResponse.json();
    return {
      success: true,
      provider: createData.provider,
    };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error',
    };
  }
}

export async function hasValidConfig(): Promise<boolean> {
  try {
    const response = await fetch('/api/v1/providers');
    if (!response.ok) return false;

    const data = await response.json();
    const providers = data.providers || [];

    return providers.some(
      (p: { status: string; enabled: boolean }) =>
        p.status === 'configured' && p.enabled
    );
  } catch {
    return false;
  }
}

export async function getDefaultModelSelection(): Promise<{
  queryModel?: string;
  genModel?: string;
}> {
  try {
    const response = await fetch('/api/v1/providers');
    if (!response.ok) return {};

    const data = await response.json();
    const providers = data.providers || [];

    const configuredProvider = providers.find(
      (p: { status: string; enabled: boolean; models?: { enabled: boolean; id: string }[] }) =>
        p.status === 'configured' && p.enabled && p.models?.some((m) => m.enabled)
    );

    if (!configuredProvider) return {};

    const enabledModels = configuredProvider.models?.filter((m: { enabled: boolean }) => m.enabled) || [];
    const defaultModel = enabledModels[0];

    if (!defaultModel) return {};

    return {
      queryModel: `${configuredProvider.name}:${defaultModel.id}`,
      genModel: `${configuredProvider.name}:${defaultModel.id}`,
    };
  } catch {
    return {};
  }
}
