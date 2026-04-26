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

  it('renders two textareas with correct column layout', () => {
    render(
      <DualInputPanel
        methodContent=""
        caption=""
        onMethodChange={mockOnMethodChange}
        onCaptionChange={mockOnCaptionChange}
      />
    );

    // Check for both textareas
    const textareas = screen.getAllByRole('textbox');
    expect(textareas).toHaveLength(2);

    const panel = textareas[0].closest('.dual-input-panel');
    expect(panel).toHaveClass('md:grid-cols-5');

    const methodSection = textareas[0].closest('.dual-input-panel__context');
    expect(methodSection).toBeInTheDocument();

    const captionSection = textareas[1].closest('.dual-input-panel__brief');
    expect(captionSection).toBeInTheDocument();
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

  it('markdown preview toggle shows/hides preview for each field', () => {
    render(
      <DualInputPanel
        methodContent="**Bold text**"
        caption=""
        onMethodChange={mockOnMethodChange}
        onCaptionChange={mockOnCaptionChange}
      />
    );

    // Find preview toggle buttons
    const previewButtons = screen.getAllByRole('button', { name: /preview/i });
    expect(previewButtons.length).toBeGreaterThan(0);

    // Click first preview toggle
    fireEvent.click(previewButtons[0]);

    // Should show preview element (check for preview content)
    // The preview should render markdown
    const previewElements = document.querySelectorAll('.prose, .markdown-preview');
    expect(previewElements.length).toBeGreaterThan(0);
  });

  it('example dropdown populates both textareas from a shared selector', () => {
    const examples = [
      { method: 'Example method 1', caption: 'Example caption 1' },
      { method: 'Example method 2', caption: 'Example caption 2' },
    ];

    render(
      <DualInputPanel
        methodContent=""
        caption=""
        onMethodChange={mockOnMethodChange}
        onCaptionChange={mockOnCaptionChange}
        examples={examples}
      />
    );

    const select = screen.getByRole('combobox');
    fireEvent.change(select, { target: { value: '0' } });
    expect(mockOnMethodChange).toHaveBeenCalledWith('Example method 1');
    expect(mockOnCaptionChange).toHaveBeenCalledWith('Example caption 1');
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
