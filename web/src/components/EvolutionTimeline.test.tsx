import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EvolutionTimeline, type EvolutionStage } from './EvolutionTimeline';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
}));

describe('EvolutionTimeline', () => {
  const mockStages: EvolutionStage[] = [
    { name: 'Planner', status: 'complete', description: 'Planning complete' },
    { name: 'Stylist', status: 'running', description: 'Applying styles' },
    { name: 'Critic Round 1', status: 'pending', description: 'Waiting to start' },
  ];

  it('renders all stages', () => {
    render(<EvolutionTimeline stages={mockStages} />);
    expect(screen.getByText('Planner')).toBeInTheDocument();
    expect(screen.getByText('Stylist')).toBeInTheDocument();
    expect(screen.getByText('Critic Round 1')).toBeInTheDocument();
  });

  it('shows correct status icons', () => {
    render(<EvolutionTimeline stages={mockStages} />);
    expect(screen.getByText('✅')).toBeInTheDocument(); // complete
    expect(screen.getByText('🔄')).toBeInTheDocument(); // running
    expect(screen.getByText('○')).toBeInTheDocument();  // pending
  });

  it('shows running status text for running stages', () => {
    render(<EvolutionTimeline stages={mockStages} />);
    expect(screen.getByText('generate.running')).toBeInTheDocument();
  });

  it('expands stage on click', () => {
    render(<EvolutionTimeline stages={mockStages} />);
    const plannerButton = screen.getByText('Planner').closest('button');
    fireEvent.click(plannerButton!);
    expect(screen.getByText('Planning complete')).toBeInTheDocument();
  });

  it('collapses when clicking again', () => {
    render(<EvolutionTimeline stages={mockStages} />);
    const plannerButton = screen.getByText('Planner').closest('button');
    fireEvent.click(plannerButton!);
    expect(screen.getByText('Planning complete')).toBeInTheDocument();
    fireEvent.click(plannerButton!);
    expect(screen.queryByText('Planning complete')).not.toBeInTheDocument();
  });

  it('renders nothing when stages array is empty', () => {
    const { container } = render(<EvolutionTimeline stages={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('shows error icon for error status', () => {
    const errorStages: EvolutionStage[] = [
      { name: 'Failed Stage', status: 'error', description: 'Something went wrong' },
    ];
    render(<EvolutionTimeline stages={errorStages} />);
    expect(screen.getByText('❌')).toBeInTheDocument();
  });

  it('accepts currentStage prop', () => {
    render(<EvolutionTimeline stages={mockStages} currentStage={1} />);
    expect(screen.getByText('Stylist')).toBeInTheDocument();
  });

  it('accepts pipelineMode prop', () => {
    render(<EvolutionTimeline stages={mockStages} pipelineMode="planner-critic" />);
    expect(screen.getByText('Planner')).toBeInTheDocument();
  });
});
