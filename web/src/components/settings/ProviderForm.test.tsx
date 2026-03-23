import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ProviderForm } from './ProviderForm';

// Mock i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('ProviderForm', () => {
  const mockOnSave = vi.fn();
  const mockOnCancel = vi.fn();
  const mockPresets = [
    {
      type: 'openai',
      display_name: 'OpenAI',
      api_host: 'https://api.openai.com/v1',
      default_models: [{ id: 'gpt-4', name: 'GPT-4', enabled: true }],
      supports_vision: true,
      docs_url: 'https://platform.openai.com',
      api_key_url: 'https://platform.openai.com/api-keys',
    },
    {
      type: 'anthropic',
      display_name: 'Anthropic',
      api_host: 'https://api.anthropic.com/v1',
      default_models: [{ id: 'claude-3', name: 'Claude 3', enabled: true }],
      supports_vision: true,
      docs_url: 'https://docs.anthropic.com',
      api_key_url: 'https://console.anthropic.com',
    },
  ];

  beforeEach(() => {
    mockOnSave.mockClear();
    mockOnCancel.mockClear();
  });

  it('renders form fields for new provider', () => {
    render(
      <ProviderForm
        presets={mockPresets}
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );
    // Check that the provider type select exists
    const select = screen.getByRole('combobox');
    expect(select).toBeInTheDocument();
  });

  it('calls onCancel when cancel button clicked', async () => {
    const user = userEvent.setup();
    render(
      <ProviderForm
        presets={mockPresets}
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );
    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelButton);
    expect(mockOnCancel).toHaveBeenCalled();
  });

  it('shows validation errors for empty required fields', async () => {
    const user = userEvent.setup();
    render(
      <ProviderForm
        presets={mockPresets}
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );
    const submitButton = screen.getByRole('button', { name: /save/i });
    await user.click(submitButton);
    // Form should show validation errors
    await waitFor(() => {
      expect(mockOnSave).not.toHaveBeenCalled();
    });
  });

  it('auto-fills fields when provider type is selected', async () => {
    const user = userEvent.setup();
    render(
      <ProviderForm
        presets={mockPresets}
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    // Select OpenAI from dropdown
    const typeSelect = screen.getByRole('combobox');
    await user.selectOptions(typeSelect, 'openai');

    // Display name should be auto-filled - find input by value
    const inputs = screen.getAllByRole('textbox');
    const displayNameInput = inputs.find(input => input.getAttribute('placeholder')?.includes('displayName'));
    expect(displayNameInput).toHaveValue('OpenAI');
  });

  it('toggles API key visibility', async () => {
    const user = userEvent.setup();
    render(
      <ProviderForm
        presets={mockPresets}
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    // Find password input by type
    const passwordInput = screen.getByPlaceholderText('sk-...');
    expect(passwordInput).toHaveAttribute('type', 'password');

    const toggleButton = screen.getByRole('button', { name: /show/i });
    await user.click(toggleButton);
    expect(passwordInput).toHaveAttribute('type', 'text');
  });
});
