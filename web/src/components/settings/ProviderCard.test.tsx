import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ProviderCard } from './ProviderCard';
import type { Provider } from '../../hooks/useProviders';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('ProviderCard', () => {
  const mockProvider: Provider = {
    id: 'test-provider',
    type: 'openai',
    name: 'openai',
    display_name: 'OpenAI',
    query_model: 'gpt-4',
    gen_model: 'gpt-4',
    status: 'configured',
    enabled: true,
    is_system: true,
    is_default: false,
    timeout: '30s',
  };

  it('renders provider display name', () => {
    render(<ProviderCard provider={mockProvider} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('OpenAI')).toBeInTheDocument();
  });

  it('shows models array when present', () => {
    const providerWithModels = {
      ...mockProvider,
      models: [
        { id: 'gpt-4', name: 'GPT-4', enabled: true },
        { id: 'gpt-3.5-turbo', name: 'GPT-3.5 Turbo', enabled: true },
      ],
    };
    render(<ProviderCard provider={providerWithModels} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('gpt-4')).toBeInTheDocument();
  });

  it('shows legacy query_model when models array is empty', () => {
    const providerWithLegacy = {
      ...mockProvider,
      models: [],
    };
    render(<ProviderCard provider={providerWithLegacy} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('gpt-4')).toBeInTheDocument();
  });

  it('calls onEdit when edit button clicked', async () => {
    const mockOnEdit = vi.fn();
    const user = (await import('@testing-library/user-event')).userEvent.setup();
    render(<ProviderCard provider={mockProvider} onEdit={mockOnEdit} onDelete={vi.fn()} />);

    const editButton = screen.getByRole('button', { name: /edit/i });
    await user.click(editButton);

    expect(mockOnEdit).toHaveBeenCalled();
  });

  it('shows default badge when provider is default', () => {
    const defaultProvider = { ...mockProvider, is_default: true };
    render(<ProviderCard provider={defaultProvider} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('settings.default')).toBeInTheDocument();
  });

  it('shows configured status badge', () => {
    render(<ProviderCard provider={mockProvider} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('settings.status.configured')).toBeInTheDocument();
  });

  it('shows no_keys status badge', () => {
    const noKeysProvider = { ...mockProvider, status: 'no_keys' as const };
    render(<ProviderCard provider={noKeysProvider} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('settings.status.no_keys')).toBeInTheDocument();
  });

  it('shows invalid status badge', () => {
    const invalidProvider = { ...mockProvider, status: 'invalid' as const };
    render(<ProviderCard provider={invalidProvider} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('settings.status.invalid')).toBeInTheDocument();
  });

  it('hides set default action for providers without keys', () => {
    const noKeysProvider = { ...mockProvider, status: 'no_keys' as const };
    render(
      <ProviderCard
        provider={noKeysProvider}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onSetDefault={vi.fn()}
      />
    );

    expect(screen.queryByRole('button', { name: 'settings.setAsDefault' })).not.toBeInTheDocument();
  });

  it('shows remove configuration action for enabled system provider without keys', () => {
    const noKeysProvider = { ...mockProvider, status: 'no_keys' as const, enabled: true };
    render(
      <ProviderCard
        provider={noKeysProvider}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onRemoveKeys={vi.fn()}
      />
    );

    expect(screen.getByRole('button', { name: 'settings.removeKeys' })).toBeInTheDocument();
  });
});
