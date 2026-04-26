import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { GeneratePanel } from './GeneratePanel';

// Mock the hooks
vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'generate.descriptionLabel': 'Description',
        'generate.descriptionPlaceholder': 'Describe your visualization',
        'generate.visualizerNode': 'Visualizer',
        'generate.defaultVisualizer': 'Default',
        'generate.submit': 'Generate',
        'generate.generating': 'Generating...',
        'generate.methodSection': 'Method Section',
        'generate.figureCaption': 'Figure Caption',
        'generate.methodPlaceholder': 'Paste method section content...',
        'generate.captionPlaceholder': 'Enter figure caption...',
        'generate.loadExample': 'Load Example',
        'generate.previewMarkdown': 'Preview Markdown',
        'generate.batchMode': 'Batch Mode',
        'generate.numCandidatesHint': 'Generate multiple variants in parallel',
        'generate.advancedSettings': 'Advanced Settings',
        'generate.aspectRatio': 'Aspect Ratio',
        'generate.criticRounds': 'Critic Rounds',
        'generate.retrievalMode': 'Retrieval Mode',
        'generate.pipelineMode': 'Pipeline Mode',
        'generate.queryModel': 'Query Model',
        'generate.genModel': 'Generation Model',
        'generate.modes.auto': 'Auto',
        'generate.modes.manual': 'Manual',
        'generate.modes.random': 'Random',
        'generate.modes.none': 'None',
        'generate.pipelines.full': 'Full Pipeline',
        'generate.pipelines.planner-critic': 'Planner + Critic Only',
        'settings.default': 'Default',
      };
      return translations[key] || key;
    },
  }),
  usePromptTemplates: () => ({
    templates: [],
    isLoading: false,
    addTemplate: vi.fn(),
    updateTemplate: vi.fn(),
    deleteTemplate: vi.fn(),
    getTemplate: vi.fn(),
  }),
}));

vi.mock('../hooks/useProviders', () => ({
  useProviders: () => ({
    providers: [],
    loading: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

describe('GeneratePanel', () => {
  const mockOnGenerate = vi.fn();

  beforeEach(() => {
    mockOnGenerate.mockClear();
  });

  it('renders DualInputPanel with two textareas', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);
    expect(screen.getByPlaceholderText('Paste method section content...')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter figure caption...')).toBeInTheDocument();
  });

  it('disables submit when both fields are empty', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);
    const button = screen.getByRole('button', { name: 'Generate' });
    expect(button).toBeDisabled();
  });

  it('enables submit when method section has content', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);
    const textareas = screen.getAllByRole('textbox');
    fireEvent.change(textareas[0], { target: { value: 'Test method content' } });
    const button = screen.getByRole('button', { name: 'Generate' });
    expect(button).not.toBeDisabled();
  });

  it('enables submit when caption has content', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);
    const textareas = screen.getAllByRole('textbox');
    fireEvent.change(textareas[1], { target: { value: 'Test caption' } });
    const button = screen.getByRole('button', { name: 'Generate' });
    expect(button).not.toBeDisabled();
  });

  it('calls onGenerate with combined prompt on submit', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);
    const textareas = screen.getAllByRole('textbox');

    // Fill both fields
    fireEvent.change(textareas[0], { target: { value: 'This is the method section.' } });
    fireEvent.change(textareas[1], { target: { value: 'Figure 1: Results' } });

    const button = screen.getByRole('button', { name: 'Generate' });
    fireEvent.click(button);

    // The new interface uses options object with content/visualIntent split
    expect(mockOnGenerate).toHaveBeenCalledWith(
      'Paper Context & References:\nThis is the method section.\n\nTarget Figure Brief:\nFigure 1: Results',
      expect.objectContaining({
        content: 'This is the method section.',
        visualIntent: 'Figure 1: Results',
        visualizerNode: undefined,
        numCandidates: undefined,
        config: expect.any(Object),
      })
    );
  });

  it('calls onGenerate with only method content when caption is empty', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);
    const textareas = screen.getAllByRole('textbox');

    fireEvent.change(textareas[0], { target: { value: 'Method only content' } });

    const button = screen.getByRole('button', { name: 'Generate' });
    fireEvent.click(button);

    expect(mockOnGenerate).toHaveBeenCalledWith(
      'Method only content',
      expect.objectContaining({
        visualizerNode: undefined,
        numCandidates: undefined,
      })
    );
  });

  it('calls onGenerate with only caption when method is empty', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);
    const textareas = screen.getAllByRole('textbox');

    fireEvent.change(textareas[1], { target: { value: 'Caption only content' } });

    const button = screen.getByRole('button', { name: 'Generate' });
    fireEvent.click(button);

    expect(mockOnGenerate).toHaveBeenCalledWith(
      'Caption only content',
      expect.objectContaining({
        visualizerNode: undefined,
        numCandidates: undefined,
      })
    );
  });

  it('shows visualizer node selector when options provided', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} visualizerNodes={['node-a', 'node-b']} />);
    fireEvent.click(screen.getByText('Advanced Settings'));
    expect(screen.getByLabelText('Visualizer')).toBeInTheDocument();
  });

  it('batch mode still works with dual inputs', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);

    // Advanced settings collapsed by default; expand to access batch mode + num candidates
    fireEvent.click(screen.getByText('Advanced Settings'));

    // Enable batch mode
    const batchCheckbox = screen.getByRole('checkbox');
    fireEvent.click(batchCheckbox);

    // Fill method content
    const textareas = screen.getAllByRole('textbox');
    fireEvent.change(textareas[0], { target: { value: 'Test method' } });

    // Find number input for candidates
    const numInput = screen.getByRole('spinbutton');
    expect(numInput).toBeInTheDocument();
    expect(numInput).toHaveValue(3); // default value
  });

  it('disabled state collapses DualInputPanel into loading skeleton', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} isGenerating={true} />);

    // While generating, DualInputPanel collapses; textareas are not rendered
    expect(screen.queryAllByRole('textbox')).toHaveLength(0);
    // Submit button shows the generating label and is disabled
    const submitButton = screen.getByRole('button', { name: /Generating/ });
    expect(submitButton).toBeDisabled();
  });

  it('ConfigPanel renders below input fields', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);
    expect(screen.getByText('Advanced Settings')).toBeInTheDocument();
  });

  it('default config values are set correctly', () => {
    render(<GeneratePanel onGenerate={mockOnGenerate} />);
    const textareas = screen.getAllByRole('textbox');
    fireEvent.change(textareas[0], { target: { value: 'Test' } });

    const button = screen.getByRole('button', { name: 'Generate' });
    fireEvent.click(button);

    const callArgs = mockOnGenerate.mock.calls[0];
    expect(callArgs[1].config.aspectRatio).toBe('16:9');
    expect(callArgs[1].config.criticRounds).toBe(3);
    expect(callArgs[1].config.retrievalMode).toBe('auto');
    expect(callArgs[1].config.pipelineMode).toBe('full');
  });
});
