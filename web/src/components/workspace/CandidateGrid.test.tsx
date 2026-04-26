import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CandidateGrid } from './CandidateGrid';
import type { Candidate } from './CandidateGrid';
import type { Artifact } from '../ArtifactPreview';

vi.mock('../../hooks', () => ({
  useLanguage: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (params) {
        return `${key} ${JSON.stringify(params)}`;
      }
      return key;
    },
  }),
}));

describe('CandidateGrid', () => {
  const mockArtifacts: Artifact[] = [
    {
      kind: 'image',
      mimeType: 'image/png',
      summary: 'Test image',
      data: 'base64data',
    },
  ];

  const mockCandidates: Candidate[] = [
    {
      id: 'candidate-1',
      index: 0,
      status: 'completed',
      artifacts: mockArtifacts,
      metadata: { model: 'gpt-4', duration: 5000, seed: 12345 },
    },
    {
      id: 'candidate-2',
      index: 1,
      status: 'running',
      metadata: { model: 'gpt-4' },
    },
    {
      id: 'candidate-3',
      index: 2,
      status: 'failed',
      error: 'Generation failed',
    },
    {
      id: 'candidate-4',
      index: 3,
      status: 'pending',
    },
  ];

  const defaultProps = {
    candidates: mockCandidates,
    onSelect: vi.fn(),
    onRefine: vi.fn(),
    onDelete: vi.fn(),
    onExport: vi.fn(),
  };

  it('renders correctly in grid view', () => {
    const { container } = render(<CandidateGrid {...defaultProps} />);
    expect(container).toMatchSnapshot();
  });

  it('renders correctly in list view', () => {
    const { container } = render(
      <CandidateGrid {...defaultProps} viewMode="list" />
    );
    expect(container).toMatchSnapshot();
  });

  it('renders correctly with empty candidates', () => {
    const { container } = render(
      <CandidateGrid {...defaultProps} candidates={[]} />
    );
    expect(container).toMatchSnapshot();
  });

  it('renders correctly in multi-select mode', () => {
    const { container } = render(
      <CandidateGrid {...defaultProps} multiSelect={true} />
    );
    expect(container).toMatchSnapshot();
  });

  it('renders correctly with selected candidate', () => {
    const { container } = render(
      <CandidateGrid {...defaultProps} selectedId="candidate-1" />
    );
    expect(container).toMatchSnapshot();
  });

  it('calls onRefine when refine button is clicked', () => {
    const mockOnRefine = vi.fn();
    render(<CandidateGrid {...defaultProps} onRefine={mockOnRefine} />);
    
    const refineButtons = screen.getAllByTitle('refine.title');
    fireEvent.click(refineButtons[0]);
    
    expect(mockOnRefine).toHaveBeenCalledWith('candidate-1');
  });

  it('calls onDelete when delete button is clicked', () => {
    const mockOnDelete = vi.fn();
    render(<CandidateGrid {...defaultProps} onDelete={mockOnDelete} />);
    
    const deleteButtons = screen.getAllByTitle('common.delete');
    fireEvent.click(deleteButtons[0]);
    
    expect(mockOnDelete).toHaveBeenCalledWith('candidate-1');
  });

  it('displays error messages for failed candidates', () => {
    render(<CandidateGrid {...defaultProps} />);
    
    expect(screen.getByText('Generation failed')).toBeInTheDocument();
  });
});
