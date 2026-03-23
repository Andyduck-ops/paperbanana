import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ConfigPanel, GenerationConfig } from './ConfigPanel';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'generate.advancedSettings': 'Advanced Settings',
        'generate.aspectRatio': 'Aspect Ratio',
        'generate.criticRounds': 'Critic Rounds',
        'generate.retrievalMode': 'Retrieval Mode',
        'generate.pipelineMode': 'Pipeline Mode',
        'generate.queryModel': 'Query Model',
        'generate.genModel': 'Generation Model',
        'generate.configureProviders': 'Configure providers in Settings',
        'generate.modes.auto': 'Auto',
        'generate.modes.manual': 'Manual',
        'generate.modes.random': 'Random',
        'generate.modes.none': 'None',
        'generate.pipelines.full': 'Full Pipeline',
        'generate.pipelines.planner-critic': 'Planner + Critic Only',
        'generate.pipelines.vanilla': 'Vanilla (Direct)',
        'settings.default': 'Default',
      };
      return translations[key] || key;
    },
  }),
}));

describe('ConfigPanel', () => {
  const defaultConfig: GenerationConfig = {
    aspectRatio: '16:9',
    criticRounds: 3,
    retrievalMode: 'auto',
    pipelineMode: 'full',
    queryModel: undefined,
    genModel: undefined,
  };

  const mockOnChange = vi.fn();
  const mockProviders = [
    {
      id: 'provider-1',
      type: 'test',
      name: 'test-provider',
      display_name: 'Test Provider',
      query_model: 'model-vision',
      gen_model: 'model-gen',
      timeout: '30s',
      status: 'configured' as const,
      enabled: true,
      is_system: false,
      is_default: true,
      models: [
        { id: 'model-vision', name: 'Vision Model', supports_vision: true, enabled: true },
        { id: 'model-gen', name: 'Gen Model', supports_vision: false, enabled: true },
      ],
    },
  ];

  it('renders collapsed by default with "Advanced Settings" header', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} />);

    expect(screen.getByText('Advanced Settings')).toBeInTheDocument();
    // Config options should not be visible when collapsed
    expect(screen.queryByLabelText('Aspect Ratio')).not.toBeInTheDocument();
  });

  it('expands on click to show all 5 config options', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} providers={mockProviders} />);

    // Click to expand
    fireEvent.click(screen.getByText('Advanced Settings'));

    // Now config options should be visible
    expect(screen.getByLabelText('Aspect Ratio')).toBeInTheDocument();
    expect(screen.getByLabelText('Critic Rounds')).toBeInTheDocument();
    expect(screen.getByLabelText('Retrieval Mode')).toBeInTheDocument();
    expect(screen.getByLabelText('Pipeline Mode')).toBeInTheDocument();
    expect(screen.getByLabelText('Query Model')).toBeInTheDocument();
    expect(screen.getByLabelText('Generation Model')).toBeInTheDocument();
  });

  it('Aspect ratio select has options 21:9, 16:9, 3:2', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    const select = screen.getByLabelText('Aspect Ratio') as HTMLSelectElement;
    const options = Array.from(select.options).map((opt) => opt.value);

    expect(options).toContain('21:9');
    expect(options).toContain('16:9');
    expect(options).toContain('3:2');
  });

  it('Critic rounds input accepts values 1-5', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    const input = screen.getByLabelText('Critic Rounds') as HTMLInputElement;

    expect(input.type).toBe('number');
    expect(input.min).toBe('1');
    expect(input.max).toBe('5');
  });

  it('Retrieval mode select has options auto/manual/random/none', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    const select = screen.getByLabelText('Retrieval Mode') as HTMLSelectElement;
    const options = Array.from(select.options).map((opt) => opt.value);

    expect(options).toContain('auto');
    expect(options).toContain('manual');
    expect(options).toContain('random');
    expect(options).toContain('none');
  });

  it('Pipeline mode select has options full/planner-critic/vanilla', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    const select = screen.getByLabelText('Pipeline Mode') as HTMLSelectElement;
    const options = Array.from(select.options).map((opt) => opt.value);

    expect(options).toContain('full');
    expect(options).toContain('planner-critic');
    expect(options).toContain('vanilla');
  });

  it('Pipeline mode default is "full"', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    const select = screen.getByLabelText('Pipeline Mode') as HTMLSelectElement;
    expect(select.value).toBe('full');
  });

  it('Pipeline mode onChange fires when selection changes', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    const select = screen.getByLabelText('Pipeline Mode');
    fireEvent.change(select, { target: { value: 'vanilla' } });

    expect(mockOnChange).toHaveBeenCalledWith({
      ...defaultConfig,
      pipelineMode: 'vanilla',
    });
  });

  it('Model dropdowns populate from providers prop', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} providers={mockProviders} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    const querySelect = screen.getByLabelText('Query Model') as HTMLSelectElement;
    const genSelect = screen.getByLabelText('Generation Model') as HTMLSelectElement;

    // Query model should show vision models with provider:model format
    const queryOptions = Array.from(querySelect.options).map((opt) => opt.value);
    expect(queryOptions).toContain('test-provider:model-vision');

    // Gen model should show all models with provider:model format
    const genOptions = Array.from(genSelect.options).map((opt) => opt.value);
    expect(genOptions).toContain('test-provider:model-vision');
    expect(genOptions).toContain('test-provider:model-gen');
  });

  it('onChange callback fires when any config changes', () => {
    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    // Change aspect ratio
    const aspectSelect = screen.getByLabelText('Aspect Ratio');
    fireEvent.change(aspectSelect, { target: { value: '21:9' } });

    expect(mockOnChange).toHaveBeenCalledWith({
      ...defaultConfig,
      aspectRatio: '21:9',
    });

    mockOnChange.mockClear();

    // Change critic rounds
    const roundsInput = screen.getByLabelText('Critic Rounds');
    fireEvent.change(roundsInput, { target: { value: '5' } });

    expect(mockOnChange).toHaveBeenCalledWith({
      ...defaultConfig,
      criticRounds: 5,
    });
  });

  it('Disabled state applies to all inputs', () => {
    const { rerender } = render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} disabled={true} />);

    // When disabled, the toggle button should be disabled
    const toggleButton = screen.getByText('Advanced Settings').closest('button');
    expect(toggleButton).toBeDisabled();

    // Clicking should not expand the panel because it's disabled
    fireEvent.click(screen.getByText('Advanced Settings'));
    expect(screen.queryByLabelText('Aspect Ratio')).not.toBeInTheDocument();

    // Now test with expanded state first, then disable
    rerender(<ConfigPanel config={defaultConfig} onChange={mockOnChange} disabled={false} />);
    fireEvent.click(screen.getByText('Advanced Settings'));

    // Now it should be expanded
    expect(screen.getByLabelText('Aspect Ratio')).toBeInTheDocument();

    // Now disable while expanded
    rerender(<ConfigPanel config={defaultConfig} onChange={mockOnChange} disabled={true} providers={mockProviders} />);

    // Check that inputs are disabled
    expect(screen.getByLabelText('Aspect Ratio')).toBeDisabled();
    expect(screen.getByLabelText('Critic Rounds')).toBeDisabled();
    expect(screen.getByLabelText('Retrieval Mode')).toBeDisabled();
    expect(screen.getByLabelText('Pipeline Mode')).toBeDisabled();
    expect(screen.getByLabelText('Query Model')).toBeDisabled();
    expect(screen.getByLabelText('Generation Model')).toBeDisabled();
  });

  it('groups models by provider using optgroup', () => {
    const mockProvidersWithStatus = [
      {
        id: 'provider-1',
        type: 'openai',
        name: 'openai',
        display_name: 'OpenAI',
        query_model: 'gpt-4o',
        gen_model: 'gpt-4o',
        timeout: '30s',
        status: 'configured' as const,
        enabled: true,
        is_system: false,
        is_default: true,
        models: [
          { id: 'gpt-4o', name: 'GPT-4o', supports_vision: true, enabled: true },
        ],
      },
      {
        id: 'provider-2',
        type: 'anthropic',
        name: 'anthropic',
        display_name: 'Anthropic',
        query_model: 'claude-3',
        gen_model: 'claude-3',
        timeout: '30s',
        status: 'configured' as const,
        enabled: true,
        is_system: false,
        is_default: false,
        models: [
          { id: 'claude-3', name: 'Claude 3', supports_vision: true, enabled: true },
        ],
      },
    ];

    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} providers={mockProvidersWithStatus} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    // Should have optgroup elements with provider display names
    const optgroups = document.querySelectorAll('optgroup');
    expect(optgroups.length).toBeGreaterThan(0);
    expect(Array.from(optgroups).some(og => og.label === 'OpenAI')).toBe(true);
    expect(Array.from(optgroups).some(og => og.label === 'Anthropic')).toBe(true);
  });

  it('filters to only configured providers', () => {
    const mockProvidersWithStatus = [
      {
        id: 'provider-1',
        type: 'openai',
        name: 'openai',
        display_name: 'OpenAI',
        query_model: 'gpt-4o',
        gen_model: 'gpt-4o',
        timeout: '30s',
        status: 'configured' as const,
        enabled: true,
        is_system: false,
        is_default: true,
        models: [
          { id: 'gpt-4o', name: 'GPT-4o', supports_vision: true, enabled: true },
        ],
      },
      {
        id: 'provider-2',
        type: 'anthropic',
        name: 'anthropic',
        display_name: 'Anthropic',
        query_model: 'claude-3',
        gen_model: 'claude-3',
        timeout: '30s',
        status: 'no_keys' as const,
        enabled: true,
        is_system: false,
        is_default: false,
        models: [
          { id: 'claude-3', name: 'Claude 3', supports_vision: true, enabled: true },
        ],
      },
    ];

    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} providers={mockProvidersWithStatus} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    // Should only have OpenAI optgroup, not Anthropic
    // Note: There are 2 optgroups (one for Query Model, one for Generation Model), both showing OpenAI
    const optgroups = document.querySelectorAll('optgroup');
    expect(optgroups.length).toBe(2);
    expect(Array.from(optgroups).every(og => og.label === 'OpenAI')).toBe(true);
    // Verify Anthropic is not present
    expect(Array.from(optgroups).some(og => og.label === 'Anthropic')).toBe(false);
  });

  it('uses provider:model value format', () => {
    const mockProvidersWithStatus = [
      {
        id: 'provider-1',
        type: 'openai',
        name: 'openai',
        display_name: 'OpenAI',
        query_model: 'gpt-4o',
        gen_model: 'gpt-4o',
        timeout: '30s',
        status: 'configured' as const,
        enabled: true,
        is_system: false,
        is_default: true,
        models: [
          { id: 'gpt-4o', name: 'GPT-4o', supports_vision: true, enabled: true },
        ],
      },
    ];

    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} providers={mockProvidersWithStatus} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    const select = screen.getByLabelText('Query Model') as HTMLSelectElement;
    const options = Array.from(select.options);

    // Find the gpt-4o option and check its value
    const gptOption = options.find(opt => opt.textContent === 'GPT-4o');
    expect(gptOption?.value).toBe('openai:gpt-4o');
  });

  it('shows empty state link when no configured providers', () => {
    const mockProvidersNoConfig = [
      {
        id: 'provider-1',
        type: 'openai',
        name: 'openai',
        display_name: 'OpenAI',
        query_model: '',
        gen_model: '',
        timeout: '30s',
        status: 'no_keys' as const,
        enabled: true,
        is_system: false,
        is_default: true,
        models: [],
      },
    ];

    render(<ConfigPanel config={defaultConfig} onChange={mockOnChange} providers={mockProvidersNoConfig} />);

    fireEvent.click(screen.getByText('Advanced Settings'));

    // Should show link to settings instead of select (appears for both model dropdowns)
    const emptyStateButtons = screen.getAllByText('Configure providers in Settings');
    expect(emptyStateButtons.length).toBe(2); // One for Query Model, one for Generation Model
    expect(screen.queryByLabelText('Query Model')).not.toBeInTheDocument();
  });
});
