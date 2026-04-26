// @ts-nocheck - Test file with unused variables
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DualInputPanel } from './DualInputPanel';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'generate.methodSection': 'Method Section',
        'generate.figureCaption': 'Figure Caption',
        'generate.referenceImage': 'Reference Image',
        'generate.methodPlaceholder': 'Paste method section content...',
        'generate.captionPlaceholder': 'Enter figure caption...',
        'generate.loadExample': 'Load Example',
        'generate.previewMarkdown': 'Preview Markdown',
        'refine.dropImage': 'Drop image here or click to upload',
        'common.clear': 'Clear',
      };
      return translations[key] || key;
    },
  }),
}));

describe('DualInputPanel', () => {
  const mockOnMethodChange = vi.fn();
  const mockOnCaptionChange = vi.fn();

  beforeEach(() => {
    mockOnMethodChange.mockClear();
    mockOnCaptionChange.mockClear();
  });

  it('renders two textareas', () => {
    render(
      <DualInputPanel
        methodContent=""
        caption=""
        onMethodChange={mockOnMethodChange}
        onCaptionChange={mockOnCaptionChange}
      />
    );

    const textareas = screen.getAllByRole('textbox');
    expect(textareas).toHaveLength(2);
  });

  it('method section textarea has correct placeholder', () => {
    const { container } = render(
      <DualInputPanel
        methodContent=""
        caption=""
        onMethodChange={mockOnMethodChange}
        onCaptionChange={mockOnCaptionChange}
      />
    );

    const methodTextarea = screen.getByPlaceholderText('Paste method section content...');
    expect(methodTextarea).toBeInTheDocument();
  });

  it('figure caption textarea has correct placeholder', () => {
    render(
      <DualInputPanel
        methodContent=""
        caption=""
        onMethodChange={mockOnMethodChange}
        onCaptionChange={mockOnCaptionChange}
      />
    );

    const captionTextarea = screen.getByPlaceholderText('Enter figure caption...');
    expect(captionTextarea).toBeInTheDocument();
  });

  it('onChange callbacks fire for both inputs', () => {
    render(
      <DualInputPanel
        methodContent=""
        caption=""
        onMethodChange={mockOnMethodChange}
        onCaptionChange={mockOnCaptionChange}
      />
    );

    const textareas = screen.getAllByRole('textbox');

    // Test method change
    fireEvent.change(textareas[0], { target: { value: 'Test method content' } });
    expect(mockOnMethodChange).toHaveBeenCalledWith('Test method content');

    // Test caption change
    fireEvent.change(textareas[1], { target: { value: 'Test caption' } });
    expect(mockOnCaptionChange).toHaveBeenCalledWith('Test caption');
  });

  it('disabled state applies to both textareas', () => {
    render(
      <DualInputPanel
        methodContent=""
        caption=""
        onMethodChange={mockOnMethodChange}
        onCaptionChange={mockOnCaptionChange}
        disabled={true}
      />
    );

    const textareas = screen.getAllByRole('textbox');
    expect(textareas[0]).toBeDisabled();
    expect(textareas[1]).toBeDisabled();
  });

  it('renders reference image upload when handler is provided', () => {
    const mockOnReferenceImageChange = vi.fn();

    render(
      <DualInputPanel
        methodContent=""
        caption=""
        onMethodChange={mockOnMethodChange}
        onCaptionChange={mockOnCaptionChange}
        onReferenceImageChange={mockOnReferenceImageChange}
      />
    );

    expect(screen.getByText('Reference Image')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /drop image here or click to upload/i })).toBeInTheDocument();
  });
});
