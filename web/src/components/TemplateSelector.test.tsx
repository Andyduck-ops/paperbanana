import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { TemplateSelector } from './TemplateSelector';
import type { PromptTemplate } from '../hooks/usePromptTemplates';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
}));

describe('TemplateSelector', () => {
  const mockTemplates: PromptTemplate[] = [
    {
      id: 'template-1',
      name: 'Test Template 1',
      methodContent: 'Method content 1',
      caption: 'Caption 1',
      createdAt: '2024-01-01T00:00:00.000Z',
    },
    {
      id: 'template-2',
      name: 'Test Template 2',
      methodContent: 'Method content 2',
      caption: 'Caption 2',
      createdAt: '2024-01-02T00:00:00.000Z',
    },
  ];

  it('renders correctly with default props', () => {
    const { container } = render(
      <TemplateSelector
        templates={mockTemplates}
        onSelect={vi.fn()}
      />
    );
    expect(container).toMatchSnapshot();
    expect(screen.getByText('templates.select')).toBeInTheDocument();
  });

  it('renders correctly with empty templates', () => {
    const { container } = render(
      <TemplateSelector
        templates={[]}
        onSelect={vi.fn()}
      />
    );
    expect(container).toMatchSnapshot();
  });

  it('renders correctly when disabled', () => {
    const { container } = render(
      <TemplateSelector
        templates={mockTemplates}
        onSelect={vi.fn()}
        disabled={true}
      />
    );
    expect(container).toMatchSnapshot();
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('renders correctly with save capability', () => {
    const { container } = render(
      <TemplateSelector
        templates={mockTemplates}
        onSelect={vi.fn()}
        onSave={vi.fn()}
        onDelete={vi.fn()}
        currentMethodContent="Test method"
        currentCaption="Test caption"
      />
    );
    expect(container).toMatchSnapshot();
  });

  it('opens dropdown when clicked', () => {
    render(
      <TemplateSelector
        templates={mockTemplates}
        onSelect={vi.fn()}
      />
    );
    
    const button = screen.getByRole('button');
    fireEvent.click(button);
    
    expect(screen.getByText('Test Template 1')).toBeInTheDocument();
    expect(screen.getByText('Test Template 2')).toBeInTheDocument();
  });

  it('calls onSelect when template is clicked', () => {
    const mockOnSelect = vi.fn();
    render(
      <TemplateSelector
        templates={mockTemplates}
        onSelect={mockOnSelect}
      />
    );
    
    fireEvent.click(screen.getByRole('button'));
    fireEvent.click(screen.getByText('Test Template 1'));
    
    expect(mockOnSelect).toHaveBeenCalledWith(mockTemplates[0]);
  });

  it('calls onDelete when delete button is clicked', () => {
    const mockOnDelete = vi.fn();
    render(
      <TemplateSelector
        templates={mockTemplates}
        onSelect={vi.fn()}
        onDelete={mockOnDelete}
      />
    );
    
    fireEvent.click(screen.getByRole('button'));
    const deleteButtons = screen.getAllByTitle('templates.delete');
    fireEvent.click(deleteButtons[0]);
    
    expect(mockOnDelete).toHaveBeenCalledWith('template-1');
  });

  it('shows save dialog when save button is clicked', () => {
    render(
      <TemplateSelector
        templates={mockTemplates}
        onSelect={vi.fn()}
        onSave={vi.fn()}
        currentMethodContent="Test content"
      />
    );
    
    fireEvent.click(screen.getByRole('button'));
    fireEvent.click(screen.getByText('templates.save'));
    
    expect(screen.getByPlaceholderText('templates.namePlaceholder')).toBeInTheDocument();
  });

  it('calls onSave when save dialog is submitted', () => {
    const mockOnSave = vi.fn();
    render(
      <TemplateSelector
        templates={mockTemplates}
        onSelect={vi.fn()}
        onSave={mockOnSave}
        currentMethodContent="Test content"
      />
    );
    
    fireEvent.click(screen.getByRole('button'));
    fireEvent.click(screen.getByText('templates.save'));
    
    const input = screen.getByPlaceholderText('templates.namePlaceholder');
    fireEvent.change(input, { target: { value: 'New Template' } });
    fireEvent.click(screen.getByText('common.save'));
    
    expect(mockOnSave).toHaveBeenCalledWith('New Template');
  });
});
