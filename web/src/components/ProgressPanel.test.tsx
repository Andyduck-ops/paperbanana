import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ProgressPanel } from './ProgressPanel';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
}));

describe('ProgressPanel', () => {
  it('renders nothing when no stages', () => {
    const { container } = render(<ProgressPanel stages={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders nothing when isVisible is false', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'complete' as const },
    ];
    const { container } = render(<ProgressPanel stages={stages} isVisible={false} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders all stages', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'complete' as const },
      { stage: 'planner', agent: 'Planner', status: 'running' as const },
    ];
    render(<ProgressPanel stages={stages} />);
    expect(screen.getByText('Retriever')).toBeInTheDocument();
    expect(screen.getByText('Planner')).toBeInTheDocument();
  });

  it('shows completed count', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'complete' as const },
      { stage: 'planner', agent: 'Planner', status: 'pending' as const },
    ];
    render(<ProgressPanel stages={stages} />);
    expect(screen.getByText('1/2')).toBeInTheDocument();
  });

  it('shows cancel button when isGenerating and onCancel provided', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'running' as const },
    ];
    const mockOnCancel = vi.fn();
    render(<ProgressPanel stages={stages} isGenerating={true} onCancel={mockOnCancel} />);
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
  });

  it('does not show cancel button when not generating', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'running' as const },
    ];
    const mockOnCancel = vi.fn();
    render(<ProgressPanel stages={stages} isGenerating={false} onCancel={mockOnCancel} />);
    expect(screen.queryByRole('button', { name: 'Cancel' })).not.toBeInTheDocument();
  });

  it('does not show cancel button when onCancel not provided', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'running' as const },
    ];
    render(<ProgressPanel stages={stages} isGenerating={true} />);
    expect(screen.queryByRole('button', { name: 'Cancel' })).not.toBeInTheDocument();
  });

  it('calls onCancel when cancel button is clicked', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'running' as const },
    ];
    const mockOnCancel = vi.fn();
    render(<ProgressPanel stages={stages} isGenerating={true} onCancel={mockOnCancel} />);
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(mockOnCancel).toHaveBeenCalledTimes(1);
  });

  it('shows estimated time when provided', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'running' as const },
    ];
    render(<ProgressPanel stages={stages} estimatedTime={45} />);
    expect(screen.getByText(/Est\.\s*~45s/)).toBeInTheDocument();
  });

  it('formats estimated time in minutes when over 60 seconds', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'running' as const },
    ];
    render(<ProgressPanel stages={stages} estimatedTime={125} />);
    expect(screen.getByText(/~2m 5s/)).toBeInTheDocument();
  });

  it('does not show estimated time when not provided', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'running' as const },
    ];
    render(<ProgressPanel stages={stages} />);
    expect(screen.queryByText(/Est\./)).not.toBeInTheDocument();
  });

  it('does not show estimated time when 0', () => {
    const stages = [
      { stage: 'retriever', agent: 'Retriever', status: 'running' as const },
    ];
    render(<ProgressPanel stages={stages} estimatedTime={0} />);
    expect(screen.queryByText(/Est\./)).not.toBeInTheDocument();
  });
});
