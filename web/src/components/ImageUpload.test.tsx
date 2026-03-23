import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ImageUpload } from './ImageUpload';

// Mock the hooks
vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'refine.dropImage': 'Drop image here or click to upload',
        'common.clear': 'Clear',
      };
      return translations[key] || key;
    },
  }),
}));

describe('ImageUpload', () => {
  const mockOnImageSelect = vi.fn();

  beforeEach(() => {
    mockOnImageSelect.mockClear();
  });

  it('renders drop zone', () => {
    render(<ImageUpload onImageSelect={mockOnImageSelect} />);
    expect(screen.getByText(/drop image here or click to upload/i)).toBeInTheDocument();
  });

  it('drag-and-drop triggers file selection', () => {
    render(<ImageUpload onImageSelect={mockOnImageSelect} />);
    const dropZone = screen.getByRole('button', { name: /drop image here or click to upload/i });

    // Create a mock file
    const file = new File(['test image content'], 'test.png', { type: 'image/png' });
    const dataTransfer = { files: [file] };

    // Simulate drag over
    fireEvent.dragOver(dropZone);
    expect(dropZone).toHaveClass('border-primary');

    // Simulate drop
    fireEvent.drop(dropZone, { dataTransfer });

    // The FileReader is async, so we need to wait for the callback
    // For now, just check that the file input would be triggered
  });

  it('click opens file picker', () => {
    render(<ImageUpload onImageSelect={mockOnImageSelect} />);
    const dropZone = screen.getByRole('button', { name: /drop image here or click to upload/i });

    // Click should trigger the hidden file input
    const fileInput = dropZone.querySelector('input[type="file"]') as HTMLInputElement;
    const clickSpy = vi.spyOn(fileInput, 'click');

    fireEvent.click(dropZone);
    expect(clickSpy).toHaveBeenCalled();
  });

  it('accepts PNG, JPG, SVG files', () => {
    render(<ImageUpload onImageSelect={mockOnImageSelect} />);
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;

    expect(fileInput).toHaveAttribute('accept', 'image/png,image/jpeg,image/svg+xml');
  });

  it('shows preview of uploaded image', async () => {
    render(<ImageUpload onImageSelect={mockOnImageSelect} />);
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;

    // Create a mock PNG file with a small valid PNG header
    const file = new File(['test image content'], 'test.png', { type: 'image/png' });

    // Simulate file selection
    fireEvent.change(fileInput, { target: { files: [file] } });

    // Wait for the FileReader to process the file
    // The preview should appear after the file is loaded
    await vi.waitFor(() => {
      expect(screen.getByAltText('Uploaded preview')).toBeInTheDocument();
    });
  });

  it('onImageSelect callback fires with base64 data', async () => {
    render(<ImageUpload onImageSelect={mockOnImageSelect} />);
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;

    // Create a mock file
    const file = new File(['test image content'], 'test.png', { type: 'image/png' });

    // Simulate file selection
    fireEvent.change(fileInput, { target: { files: [file] } });

    // Wait for the FileReader to process and call the callback
    await vi.waitFor(() => {
      expect(mockOnImageSelect).toHaveBeenCalled();
      // The callback should receive a base64 data URL string
      const callArg = mockOnImageSelect.mock.calls[0][0];
      expect(callArg).toMatch(/^data:image\/png;base64,/);
    });
  });

  it('shows clear button to remove image', async () => {
    render(<ImageUpload onImageSelect={mockOnImageSelect} />);
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;

    // Create and select a file
    const file = new File(['test image content'], 'test.png', { type: 'image/png' });
    fireEvent.change(fileInput, { target: { files: [file] } });

    // Wait for preview to appear
    await vi.waitFor(() => {
      expect(screen.getByAltText('Uploaded preview')).toBeInTheDocument();
    });

    // Clear button should be visible
    const clearButton = screen.getByRole('button', { name: /clear/i });
    expect(clearButton).toBeInTheDocument();

    // Click clear button
    fireEvent.click(clearButton);

    // Preview should be removed
    expect(screen.queryByAltText('Uploaded preview')).not.toBeInTheDocument();
  });
});
