import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DangerZone } from './DangerZone';

// Mock useTranslation hook
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const translations: Record<string, string> = {
        'settings.dangerZone': 'Danger Zone',
        'settings.dangerZoneHint': 'Irreversible actions that affect system configuration.',
        'settings.resetProviders': 'Reset System Providers',
        'settings.resetConfirmTitle': 'Reset All System Providers',
        'settings.resetConfirmHint': 'This will clear all API keys for system providers.',
        'settings.resetWarning': 'Warning: Any in-progress generations may fail.',
        'settings.typeReset': "Type 'RESET' to confirm",
        'settings.confirmReset': 'Reset',
        'settings.resetSuccess': `Successfully cleared ${params?.count ?? 0} API key(s).`,
        'common.cancel': 'Cancel',
        'common.loading': 'Loading...',
        'common.error': 'An error occurred',
        'error.networkError': 'Network error',
      };
      return translations[key] || key;
    },
  }),
}));

// Mock fetch
const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

describe('DangerZone', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it('renders danger zone section', () => {
    render(<DangerZone />);
    expect(screen.getByText('Danger Zone')).toBeInTheDocument();
  });

  it('shows warning icon', () => {
    render(<DangerZone />);
    expect(screen.getByText('!')).toBeInTheDocument();
  });

  it('shows "Reset System Providers" button', () => {
    render(<DangerZone />);
    expect(screen.getByText('Reset System Providers')).toBeInTheDocument();
  });

  it('clicking button opens confirmation modal', () => {
    render(<DangerZone />);
    const button = screen.getByText('Reset System Providers');
    fireEvent.click(button);
    expect(screen.getByText('Reset All System Providers')).toBeInTheDocument();
  });

  it('modal requires typing "RESET" to enable confirm button', () => {
    render(<DangerZone />);
    fireEvent.click(screen.getByText('Reset System Providers'));

    const confirmButton = screen.getByTestId('reset-confirm-button');
    expect(confirmButton).toBeDisabled();

    const input = screen.getByTestId('reset-confirm-input');
    fireEvent.change(input, { target: { value: 'RESET' } });

    expect(confirmButton).not.toBeDisabled();
  });

  it('confirm button calls API with confirm: "RESET"', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ keys_cleared: 3 }),
    });

    render(<DangerZone />);
    fireEvent.click(screen.getByText('Reset System Providers'));

    const input = screen.getByTestId('reset-confirm-input');
    fireEvent.change(input, { target: { value: 'RESET' } });

    const confirmButton = screen.getByTestId('reset-confirm-button');
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/v1/providers/reset', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ confirm: 'RESET' }),
      });
    });
  });

  it('success shows confirmation message', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ keys_cleared: 3 }),
    });

    render(<DangerZone />);
    fireEvent.click(screen.getByText('Reset System Providers'));

    const input = screen.getByTestId('reset-confirm-input');
    fireEvent.change(input, { target: { value: 'RESET' } });

    const confirmButton = screen.getByTestId('reset-confirm-button');
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(screen.getByText('Successfully cleared 3 API key(s).')).toBeInTheDocument();
    });
  });

  it('error shows error message', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      json: () => Promise.resolve({ error: 'Reset failed' }),
    });

    render(<DangerZone />);
    fireEvent.click(screen.getByText('Reset System Providers'));

    const input = screen.getByTestId('reset-confirm-input');
    fireEvent.change(input, { target: { value: 'RESET' } });

    const confirmButton = screen.getByTestId('reset-confirm-button');
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(screen.getByText('Reset failed')).toBeInTheDocument();
    });
  });

  it('cancel button closes modal', () => {
    render(<DangerZone />);
    fireEvent.click(screen.getByText('Reset System Providers'));
    expect(screen.getByText('Reset All System Providers')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Cancel'));
    expect(screen.queryByText('Reset All System Providers')).not.toBeInTheDocument();
  });

  it('typing wrong text keeps confirm button disabled', () => {
    render(<DangerZone />);
    fireEvent.click(screen.getByText('Reset System Providers'));

    const input = screen.getByTestId('reset-confirm-input');
    fireEvent.change(input, { target: { value: 'reset' } });

    const confirmButton = screen.getByTestId('reset-confirm-button');
    expect(confirmButton).toBeDisabled();
  });
});
