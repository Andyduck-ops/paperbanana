import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { VersionTimeline } from './VersionTimeline';
import type { Version } from '../../hooks/useVersions';

vi.mock('../../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
}));

describe('VersionTimeline', () => {
  const mockVersions: Version[] = [
    {
      id: 'version-1',
      visualization_id: 'viz-1',
      version: 3,
      created_at: '2024-01-03T12:00:00.000Z',
      artifacts: [
        { id: 'art-1', kind: 'image', mime_type: 'image/png', summary: 'Chart v3' },
      ],
    },
    {
      id: 'version-2',
      visualization_id: 'viz-1',
      version: 2,
      created_at: '2024-01-02T10:00:00.000Z',
      artifacts: [
        { id: 'art-2', kind: 'image', mime_type: 'image/png', summary: 'Chart v2' },
      ],
    },
    {
      id: 'version-3',
      visualization_id: 'viz-1',
      version: 1,
      created_at: '2024-01-01T08:00:00.000Z',
      artifacts: [],
    },
  ];

  const defaultProps = {
    versions: mockVersions,
    isLoading: false,
    onRestore: vi.fn(),
    onPreview: vi.fn(),
  };

  it('renders correctly with versions', () => {
    const { container } = render(<VersionTimeline {...defaultProps} />);
    expect(container).toMatchSnapshot();
  });

  it('renders correctly when loading', () => {
    const { container } = render(<VersionTimeline {...defaultProps} isLoading={true} />);
    expect(container).toMatchSnapshot();
  });

  it('renders correctly with empty versions', () => {
    const { container } = render(<VersionTimeline {...defaultProps} versions={[]} />);
    expect(container).toMatchSnapshot();
  });

  it('displays version count', () => {
    render(<VersionTimeline {...defaultProps} />);
    
    expect(screen.getByText(/version.versions/)).toBeInTheDocument();
  });

  it('marks latest version correctly', () => {
    render(<VersionTimeline {...defaultProps} />);
    
    const latestBadges = screen.getAllByText('version.latest');
    expect(latestBadges.length).toBe(1);
  });

  it('calls onRestore when restore button is clicked', () => {
    const mockOnRestore = vi.fn();
    render(<VersionTimeline {...defaultProps} onRestore={mockOnRestore} />);
    
    const restoreButtons = screen.getAllByText('version.restore');
    fireEvent.click(restoreButtons[1]); // Click on version 2
    
    expect(mockOnRestore).toHaveBeenCalledWith('version-2');
  });

  it('calls onPreview when preview button is clicked', () => {
    const mockOnPreview = vi.fn();
    render(<VersionTimeline {...defaultProps} onPreview={mockOnPreview} />);
    
    const previewButtons = screen.getAllByText('version.preview');
    fireEvent.click(previewButtons[0]);
    
    expect(mockOnPreview).toHaveBeenCalledWith(mockVersions[0]);
  });

  it('renders timeline with visual connector', () => {
    const { container } = render(<VersionTimeline {...defaultProps} />);
    
    // Check for timeline markers (dots)
    const dots = container.querySelectorAll('.version-timeline-dot');
    expect(dots.length).toBe(3);
  });
});
