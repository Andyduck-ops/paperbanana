import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { HistorySidebar } from './HistorySidebar';

vi.mock('../hooks', () => ({
  useLanguage: () => ({ t: (k: string) => k, language: 'en' }),
  useHistory: () => ({
    sessions: [
      { id: 's1', projectId: 'p1', createdAt: '2026-03-17T10:00:00Z', status: 'complete' },
    ],
    isLoading: false,
    error: null,
  }),
}));

describe('HistorySidebar', () => {
  it('renders history title', () => {
    render(<HistorySidebar />);
    expect(screen.getByText('history.title')).toBeInTheDocument();
  });

  it('renders sessions', () => {
    render(<HistorySidebar />);
    expect(screen.getByText('history.untitled')).toBeInTheDocument();
  });

  it('calls onSelectSession when item clicked', () => {
    const mockSelect = vi.fn();
    render(<HistorySidebar onSelectSession={mockSelect} />);
    screen.getByText('history.untitled').click();
    expect(mockSelect).toHaveBeenCalledWith('s1');
  });
});
