import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { HistoryItem } from './HistoryItem';
import type { HistorySession } from '../../hooks';

// Mock the hooks
vi.mock('../../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
    language: 'en',
  }),
}));

describe('HistoryItem', () => {
  const mockSession: HistorySession = {
    id: 'session-1',
    projectId: 'project-1',
    createdAt: '2026-03-24T10:00:00Z',
    status: 'completed',
    prompt: 'Generate a scatter plot showing correlation between temperature and sales',
  };

  it('renders session with semantic summary', () => {
    const { container } = render(<HistoryItem session={mockSession} />);
    const title = container.querySelector('.history-item__title');

    expect(title).toBeInTheDocument();
    expect(title?.textContent).toMatch(/scatter plot/i);
  });

  it('renders timestamp', () => {
    render(<HistoryItem session={mockSession} />);

    // Should show time
    expect(screen.getByText(/10:00|AM|PM/i)).toBeInTheDocument();
  });

  it('shows selected state', () => {
    const { container } = render(
      <HistoryItem session={mockSession} isSelected={true} />
    );

    const button = container.querySelector('button');
    expect(button).toHaveClass('history-item--selected');
  });

  it('calls onClick when clicked', () => {
    const onClick = vi.fn();
    const { container } = render(
      <HistoryItem session={mockSession} onClick={onClick} />
    );

    const button = container.querySelector('button');
    fireEvent.click(button!);

    expect(onClick).toHaveBeenCalled();
  });

  it('renders different status indicators', () => {
    const statuses: Array<HistorySession['status']> = ['completed', 'failed', 'running', 'pending'];

    statuses.forEach((status) => {
      const { container } = render(
        <HistoryItem session={{ ...mockSession, status }} />
      );

      // Should render without errors
      expect(container.querySelector('button')).toBeInTheDocument();
    });
  });

  it('renders mode indicator based on prompt', () => {
    const { container } = render(<HistoryItem session={mockSession} />);

    // Should show generate mode indicator
    expect(container.textContent?.toLowerCase()).toContain('generate');
  });

  it('handles empty prompt gracefully', () => {
    const emptySession: HistorySession = {
      ...mockSession,
      prompt: undefined,
    };

    render(<HistoryItem session={emptySession} />);

    // Should show default title
    expect(screen.getByText('Recent workspace task')).toBeInTheDocument();
  });
});
