import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
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
});
