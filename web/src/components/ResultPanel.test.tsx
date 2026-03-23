import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ResultPanel } from './ResultPanel';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
}));

describe('ResultPanel', () => {
  const mockArtifacts = [
    { kind: 'visualization', mimeType: 'image/png', summary: 'Test chart' },
  ];

  it('renders session ID', () => {
    render(<ResultPanel sessionId="test-123" artifacts={mockArtifacts} />);
    expect(screen.getByText('test-123')).toBeInTheDocument();
  });

  it('renders artifacts', () => {
    render(<ResultPanel sessionId="test-123" artifacts={mockArtifacts} />);
    expect(screen.getByText('Test chart')).toBeInTheDocument();
  });

  it('calls onNewGeneration when button clicked', () => {
    const mockOnNew = vi.fn();
    render(
      <ResultPanel
        sessionId="test-123"
        artifacts={mockArtifacts}
        onNewGeneration={mockOnNew}
      />
    );
    screen.getByText('generate.new').click();
    expect(mockOnNew).toHaveBeenCalled();
  });
});
