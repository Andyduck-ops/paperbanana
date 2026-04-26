import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { TemplateManager } from './TemplateManager';
import type { Template } from '../../hooks/useTemplates';

vi.mock('../../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
}));

// Mock window.confirm
Object.defineProperty(window, 'confirm', {
  writable: true,
  value: vi.fn(),
});

describe('TemplateManager', () => {
  const mockTemplates: Template[] = [
    {
      id: 'template-1',
      name: 'Test Template 1',
      description: 'Description 1',
      category: 'general',
      created_at: '2024-01-01T00:00:00.000Z',
    },
    {
      id: 'template-2',
      name: 'Test Template 2',
      description: 'Description 2',
      category: 'custom',
      created_at: '2024-01-02T00:00:00.000Z',
    },
  ];

  const defaultProps = {
    templates: mockTemplates,
    isLoading: false,
    onCreate: vi.fn(),
    onUpdate: vi.fn(),
    onDelete: vi.fn(),
  };

  it('renders correctly with templates', () => {
    const { container } = render(<TemplateManager {...defaultProps} />);
    expect(container).toMatchSnapshot();
  });

  it('renders correctly when loading', () => {
    const { container } = render(
      <TemplateManager {...defaultProps} isLoading={true} />
    );
    expect(container).toMatchSnapshot();
  });

  it('renders correctly with empty templates', () => {
    const { container } = render(
      <TemplateManager {...defaultProps} templates={[]} />
    );
    expect(container).toMatchSnapshot();
  });

  it('opens create modal when new button is clicked', () => {
    render(<TemplateManager {...defaultProps} />);
    
    fireEvent.click(screen.getByText('template.new'));
    
    expect(screen.getByText('template.create')).toBeInTheDocument();
  });

  it('opens edit modal when edit button is clicked', () => {
    render(<TemplateManager {...defaultProps} />);
    
    const editButtons = screen.getAllByTitle('common.edit');
    fireEvent.click(editButtons[0]);
    
    expect(screen.getByText('template.edit')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Test Template 1')).toBeInTheDocument();
  });

  it('calls onDelete when delete is confirmed', () => {
    const mockOnDelete = vi.fn();
    (window.confirm as ReturnType<typeof vi.fn>).mockReturnValue(true);
    
    render(<TemplateManager {...defaultProps} onDelete={mockOnDelete} />);
    
    const deleteButtons = screen.getAllByTitle('common.delete');
    fireEvent.click(deleteButtons[0]);
    
    expect(window.confirm).toHaveBeenCalled();
    expect(mockOnDelete).toHaveBeenCalledWith('template-1');
  });
});
