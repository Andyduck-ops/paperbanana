import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { StageCard } from './StageCard';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'generate.artifacts': 'artifacts',
        'generate.failed': 'Failed',
        'common.retry': 'Retry',
        'common.loading': 'Loading...',
      };
      return translations[key] || key;
    },
  }),
}));

describe('StageCard', () => {
  it('renders with pending status', () => {
    render(<StageCard stage="retriever" agent="Retriever" status="pending" />);
    expect(screen.getByText('Retriever')).toBeInTheDocument();
    expect(screen.getByText('retriever')).toBeInTheDocument();
    expect(screen.getByText('○')).toBeInTheDocument();
  });

  it('renders with running status and shows animation class', () => {
    const { container } = render(<StageCard stage="planner" agent="Planner" status="running" />);
    expect(screen.getByText('Planner')).toBeInTheDocument();
    expect(screen.getByText('◆')).toBeInTheDocument();
    expect(container.querySelector('.animate-pulse')).toBeInTheDocument();
  });

  it('renders with complete status and shows checkmark', () => {
    render(<StageCard stage="visualizer" agent="Visualizer" status="complete" />);
    expect(screen.getByText('✓')).toBeInTheDocument();
  });

  it('renders with error status and shows X icon', () => {
    render(<StageCard stage="stylist" agent="Stylist" status="error" />);
    expect(screen.getByText('✗')).toBeInTheDocument();
  });

  it('shows summary when complete', () => {
    render(
      <StageCard
        stage="retriever"
        agent="Retriever"
        status="complete"
        summary="Found 5 references"
      />
    );
    expect(screen.getByText('Found 5 references')).toBeInTheDocument();
  });

  it('does not show summary when not complete', () => {
    render(
      <StageCard
        stage="retriever"
        agent="Retriever"
        status="running"
        summary="Found 5 references"
      />
    );
    expect(screen.queryByText('Found 5 references')).not.toBeInTheDocument();
  });

  it('shows artifact count badge when provided', () => {
    render(
      <StageCard
        stage="retriever"
        agent="Retriever"
        status="complete"
        artifactCount={3}
      />
    );
    expect(screen.getByText('3 artifacts')).toBeInTheDocument();
  });

  it('does not show artifact count when 0', () => {
    render(
      <StageCard
        stage="retriever"
        agent="Retriever"
        status="complete"
        artifactCount={0}
      />
    );
    expect(screen.queryByText('0 artifacts')).not.toBeInTheDocument();
  });

  it('shows error message when status is error', () => {
    render(
      <StageCard
        stage="planner"
        agent="Planner"
        status="error"
        error="Network timeout"
      />
    );
    expect(screen.getByText('Failed')).toBeInTheDocument();
    expect(screen.getByText('Network timeout')).toBeInTheDocument();
  });

  it('shows retry button when onRetry is provided and status is error', () => {
    const mockOnRetry = vi.fn();
    render(
      <StageCard
        stage="planner"
        agent="Planner"
        status="error"
        error="Network timeout"
        onRetry={mockOnRetry}
      />
    );
    const retryButton = screen.getByRole('button', { name: 'Retry' });
    expect(retryButton).toBeInTheDocument();
    fireEvent.click(retryButton);
    expect(mockOnRetry).toHaveBeenCalled();
  });

  it('disables retry button when isRetrying is true', () => {
    const mockOnRetry = vi.fn();
    render(
      <StageCard
        stage="planner"
        agent="Planner"
        status="error"
        error="Network timeout"
        onRetry={mockOnRetry}
        isRetrying={true}
      />
    );
    const retryButton = screen.getByRole('button', { name: /Retry|Loading/i });
    expect(retryButton).toBeDisabled();
    expect(retryButton).toHaveTextContent('Loading...');
  });

  it('does not show retry button when status is not error', () => {
    const mockOnRetry = vi.fn();
    render(
      <StageCard
        stage="planner"
        agent="Planner"
        status="complete"
        onRetry={mockOnRetry}
      />
    );
    expect(screen.queryByRole('button', { name: 'Retry' })).not.toBeInTheDocument();
  });

  it('shows duration when provided and status is complete', () => {
    render(
      <StageCard
        stage="retriever"
        agent="Retriever"
        status="complete"
        duration={5000}
      />
    );
    expect(screen.getByText('5.0s')).toBeInTheDocument();
  });

  it('formats duration in minutes when over 60 seconds', () => {
    render(
      <StageCard
        stage="retriever"
        agent="Retriever"
        status="complete"
        duration={125000}
      />
    );
    expect(screen.getByText('2m 5s')).toBeInTheDocument();
  });

  it('formats duration in ms when under 1 second', () => {
    render(
      <StageCard
        stage="retriever"
        agent="Retriever"
        status="complete"
        duration={500}
      />
    );
    expect(screen.getByText('500ms')).toBeInTheDocument();
  });

  it('does not show duration when status is not complete', () => {
    render(
      <StageCard
        stage="retriever"
        agent="Retriever"
        status="running"
        duration={5000}
      />
    );
    expect(screen.queryByText('5.0s')).not.toBeInTheDocument();
  });

  it('renders with not_run status', () => {
    render(<StageCard stage="critic" agent="Critic" status="not_run" />);
    expect(screen.getByText('—')).toBeInTheDocument();
  });
});
