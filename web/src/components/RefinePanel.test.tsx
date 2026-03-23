import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RefinePanel, type RefineRequest } from './RefinePanel';

// Mock the hooks
vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'refine.title': 'Refine Image',
        'refine.dropImage': 'Drop image here or click to upload',
        'refine.instructions': 'Refinement Instructions',
        'refine.instructionsPlaceholder': 'Describe how you want to improve the image...',
        'refine.resolution': 'Target Resolution',
        'refine.resolution2K': '2K (2560x1440)',
        'refine.resolution4K': '4K (3840x2160)',
        'refine.refineButton': 'Refine Image',
        'refine.refining': 'Refining...',
        'common.clear': 'Clear',
      };
      return translations[key] || key;
    },
  }),
}));

describe('RefinePanel', () => {
  const mockOnRefine = vi.fn();

  beforeEach(() => {
    mockOnRefine.mockClear();
  });

  it('renders ImageUpload component', () => {
    render(<RefinePanel onRefine={mockOnRefine} />);
    expect(screen.getByText(/drop image here or click to upload/i)).toBeInTheDocument();
  });

  it('renders instructions textarea', () => {
    render(<RefinePanel onRefine={mockOnRefine} />);
    const textarea = screen.getByPlaceholderText(/describe how you want to improve the image/i);
    expect(textarea).toBeInTheDocument();
  });

  it('renders resolution selector (2K, 4K)', () => {
    render(<RefinePanel onRefine={mockOnRefine} />);
    expect(screen.getByText('Target Resolution')).toBeInTheDocument();
    expect(screen.getByLabelText(/2K/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/4K/i)).toBeInTheDocument();
  });

  it('submit button disabled until image uploaded', () => {
    render(<RefinePanel onRefine={mockOnRefine} />);
    const submitButton = screen.getByRole('button', { name: /refine image/i });
    expect(submitButton).toBeDisabled();
  });

  it('onRefine callback fires with image data and instructions', async () => {
    render(<RefinePanel onRefine={mockOnRefine} />);

    // Upload an image via the file input
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['test image content'], 'test.png', { type: 'image/png' });
    fireEvent.change(fileInput, { target: { files: [file] } });

    // Wait for image to be uploaded
    await waitFor(() => {
      expect(screen.getByAltText('Uploaded preview')).toBeInTheDocument();
    });

    // Fill in instructions
    const textarea = screen.getByPlaceholderText(/describe how you want to improve the image/i);
    fireEvent.change(textarea, { target: { value: 'Make the colors more vibrant' } });

    // Select 4K resolution
    const radio4K = screen.getByLabelText(/4K/i);
    fireEvent.click(radio4K);

    // Submit button should now be enabled
    const submitButton = screen.getByRole('button', { name: /refine image/i });
    expect(submitButton).not.toBeDisabled();

    // Click submit
    fireEvent.click(submitButton);

    // Verify callback
    expect(mockOnRefine).toHaveBeenCalled();
    const callArg: RefineRequest = mockOnRefine.mock.calls[0][0];
    expect(callArg.imageData).toMatch(/^data:image\/png;base64,/);
    expect(callArg.instructions).toBe('Make the colors more vibrant');
    expect(callArg.resolution).toBe('4K');
  });

  it('shows loading state during refinement', () => {
    render(<RefinePanel onRefine={mockOnRefine} isRefining={true} />);

    // Should show "Refining..." text
    expect(screen.getByText('Refining...')).toBeInTheDocument();

    // The button should be disabled
    const submitButton = screen.getByRole('button', { name: /refining/i });
    expect(submitButton).toBeDisabled();
  });

  it('default resolution is 2K', () => {
    render(<RefinePanel onRefine={mockOnRefine} />);
    const radio2K = screen.getByLabelText(/2K/i) as HTMLInputElement;
    expect(radio2K).toBeChecked();
  });
});
