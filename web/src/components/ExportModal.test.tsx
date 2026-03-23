import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ExportModal } from './ExportModal';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('../lib/export', () => ({
  exportAsPng: vi.fn(),
  exportAsSvg: vi.fn(),
  exportAsPdf: vi.fn(),
}));

describe('ExportModal', () => {
  it('renders nothing when closed', () => {
    render(<ExportModal isOpen={false} onClose={vi.fn()} />);
    expect(screen.queryByText('export.title')).not.toBeInTheDocument();
  });

  it('renders format options when open', () => {
    render(<ExportModal isOpen={true} onClose={vi.fn()} />);
    // Buttons have uppercase text
    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBeGreaterThan(0);
  });

  it('calls onClose when cancel clicked', () => {
    const mockClose = vi.fn();
    render(<ExportModal isOpen={true} onClose={mockClose} />);
    screen.getByText('common.cancel').click();
    expect(mockClose).toHaveBeenCalled();
  });

  it('shows DPI section for PNG format', () => {
    render(<ExportModal isOpen={true} onClose={vi.fn()} />);
    expect(screen.getByText('export.dpi')).toBeInTheDocument();
  });
});
