import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { HistoryPanel } from './HistoryPanel';

// Mock the hooks
vi.mock('../../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
    language: 'en',
  }),
  useHistory: () => ({
    sessions: [
      {
        id: 'session-1',
        projectId: 'project-1',
        createdAt: '2026-03-24T10:00:00Z',
        status: 'completed',
        prompt: 'Generate a scatter plot showing correlation between variables',
      },
      {
        id: 'session-2',
        projectId: 'project-1',
        createdAt: '2026-03-24T09:00:00Z',
        status: 'failed',
        prompt: 'Create a bar chart',
      },
    ],
    isLoading: false,
    error: null,
    count: 2,
    refresh: vi.fn(),
    restoreSession: vi.fn(),
  }),
}));

describe('HistoryPanel', () => {
  it('renders closed state correctly', () => {
    const { container } = render(
      <HistoryPanel
        isOpen={false}
        onClose={vi.fn()}
        onSelectSession={vi.fn()}
      />
    );

    expect(container).toBeEmptyDOMElement();
  });

  it('renders open state correctly', () => {
    const { container } = render(
      <HistoryPanel
        isOpen={true}
        onClose={vi.fn()}
        onSelectSession={vi.fn()}
      />
    );

    // Panel should be visible
    const panel = container.querySelector('[role="dialog"]');
    expect(panel).toHaveClass('translate-x-0');
  });

  it('displays history title', () => {
    render(
      <HistoryPanel
        isOpen={true}
        onClose={vi.fn()}
        onSelectSession={vi.fn()}
      />
    );

    expect(screen.getByText('history.title')).toBeInTheDocument();
  });

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn();
    render(
      <HistoryPanel
        isOpen={true}
        onClose={onClose}
        onSelectSession={vi.fn()}
      />
    );

    const closeButton = screen.getByLabelText('common.close');
    fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalled();
  });

  it('calls onSelectSession when history item is clicked', () => {
    const onSelectSession = vi.fn();
    render(
      <HistoryPanel
        isOpen={true}
        onClose={vi.fn()}
        onSelectSession={onSelectSession}
      />
    );

    // Find and click the first history item
    const items = screen.getAllByText(/Scatter plot|Bar chart/i);
    if (items.length > 0) {
      fireEvent.click(items[0].closest('button')!);
      expect(onSelectSession).toHaveBeenCalledWith('session-1');
    }
  });

  it('renders sessions with semantic summaries', () => {
    render(
      <HistoryPanel
        isOpen={true}
        onClose={vi.fn()}
        onSelectSession={vi.fn()}
      />
    );

    // Should show semantic summaries (shortened titles)
    expect(screen.getAllByText(/Scatter plot/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Bar chart/i).length).toBeGreaterThan(0);
  });

  it('closes panel after selecting a session', () => {
    const onClose = vi.fn();
    const onSelectSession = vi.fn();
    render(
      <HistoryPanel
        isOpen={true}
        onClose={onClose}
        onSelectSession={onSelectSession}
      />
    );

    // Find and click the first history item
    const items = screen.getAllByText(/Scatter plot|Bar chart/i);
    if (items.length > 0) {
      fireEvent.click(items[0].closest('button')!);
      expect(onClose).toHaveBeenCalled();
    }
  });
});
