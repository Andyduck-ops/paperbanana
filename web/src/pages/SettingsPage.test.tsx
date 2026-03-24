import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { SettingsPage } from './SettingsPage';

const mockUseModelConfig = vi.fn();

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (_key: string, fallback?: string) => fallback || _key,
  }),
}));

vi.mock('../context/ModelConfigContext', () => ({
  ModelConfigProvider: ({ children }: { children: ReactNode }) => children,
  useModelConfig: () => mockUseModelConfig(),
}));

vi.mock('../components', () => ({
  ChannelManager: ({ channels }: { channels: Array<{ id: string }> }) => (
    <div>channels:{channels.length}</div>
  ),
  ModelSelector: ({ channels }: { channels: Array<{ id: string }> }) => (
    <div>models:{channels.length}</div>
  ),
  RoleMapping: ({ channels }: { channels: Array<{ id: string }> }) => (
    <div>roles:{channels.length}</div>
  ),
}));

describe('SettingsPage', () => {
  beforeEach(() => {
    mockUseModelConfig.mockReset();
  });

  it('keeps custom draft channels visible while still hiding untouched system presets', () => {
    mockUseModelConfig.mockReturnValue({
      channels: [
        {
          id: 'system-openai',
          name: 'openai',
          type: 'openai',
          display_name: 'OpenAI',
          timeout: 60,
          enabled: true,
          models: [],
          status: 'no_keys',
          is_system: true,
        },
        {
          id: 'lab-proxy',
          name: 'lab-proxy',
          type: 'custom',
          display_name: 'Lab Proxy',
          timeout: 60,
          enabled: true,
          models: [],
          status: 'no_keys',
          is_system: false,
        },
      ],
      role_assignments: {
        image_generation: null,
        retrieval_reasoning: null,
      },
      loading: false,
      addChannel: vi.fn(),
      updateChannel: vi.fn(),
      deleteChannel: vi.fn(),
      fetchChannelModels: vi.fn(),
      assignRole: vi.fn(),
      clearRole: vi.fn(),
    });

    render(<SettingsPage onBack={vi.fn()} />);

    expect(screen.getByText('channels:1')).toBeInTheDocument();
    expect(screen.getByText('No models available')).toBeInTheDocument();
    expect(screen.getByText('Fetch models from a channel to get started')).toBeInTheDocument();
  });

  it('shows configured channels in models and role mapping sections', () => {
    mockUseModelConfig.mockReturnValue({
      channels: [
        {
          id: 'configured-openai',
          name: 'openai',
          type: 'openai',
          display_name: 'OpenAI',
          timeout: 60,
          enabled: true,
          models: [{ id: 'gpt-4.1', name: 'gpt-4.1', enabled: true }],
          status: 'configured',
          is_system: true,
        },
      ],
      role_assignments: {
        image_generation: null,
        retrieval_reasoning: null,
      },
      loading: false,
      addChannel: vi.fn(),
      updateChannel: vi.fn(),
      deleteChannel: vi.fn(),
      fetchChannelModels: vi.fn(),
      assignRole: vi.fn(),
      clearRole: vi.fn(),
    });

    render(<SettingsPage onBack={vi.fn()} />);

    expect(screen.getByText('channels:1')).toBeInTheDocument();
    expect(screen.getByText('models:1')).toBeInTheDocument();
    expect(screen.getByText('roles:1')).toBeInTheDocument();
  });
});
