import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ModelTagList } from './ModelTagList';

describe('ModelTagList', () => {
  it('renders empty state with add button', () => {
    const onChange = vi.fn();
    render(<ModelTagList models={[]} onChange={onChange} />);
    expect(screen.getByRole('button', { name: '+' })).toBeInTheDocument();
  });

  it('renders existing model tags', () => {
    const onChange = vi.fn();
    const models = [
      { id: 'gpt-4', name: 'GPT-4' },
      { id: 'claude-3', name: 'Claude 3' },
    ];
    render(<ModelTagList models={models} onChange={onChange} />);
    expect(screen.getByText('gpt-4')).toBeInTheDocument();
    expect(screen.getByText('claude-3')).toBeInTheDocument();
  });

  it('shows remove button on hover (in DOM)', () => {
    const onChange = vi.fn();
    const models = [{ id: 'gpt-4', name: 'GPT-4' }];
    render(<ModelTagList models={models} onChange={onChange} />);
    // Remove button exists but is hidden via CSS
    const removeButtons = screen.getAllByRole('button');
    expect(removeButtons.length).toBeGreaterThan(1); // + button + remove button
  });

  it('calls onChange when remove button clicked', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const models = [{ id: 'gpt-4', name: 'GPT-4' }];
    render(<ModelTagList models={models} onChange={onChange} />);

    // Find and click the remove button (x button)
    const removeButton = screen.getByRole('button', { name: /remove/i });
    await user.click(removeButton);

    expect(onChange).toHaveBeenCalledWith([]);
  });

  it('shows add input when + button clicked', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ModelTagList models={[]} onChange={onChange} />);

    const addButton = screen.getByRole('button', { name: '+' });
    await user.click(addButton);

    // Should show manual input field
    expect(screen.getByPlaceholderText('model-id')).toBeInTheDocument();
  });

  it('adds a model via manual input', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ModelTagList models={[]} onChange={onChange} />);

    // Open add panel
    const addButton = screen.getByRole('button', { name: '+' });
    await user.click(addButton);

    // Type model id
    const input = screen.getByPlaceholderText('model-id');
    await user.type(input, 'new-model');

    // Click add button in the panel
    const confirmButton = screen.getByRole('button', { name: /add/i });
    await user.click(confirmButton);

    expect(onChange).toHaveBeenCalledWith([
      { id: 'new-model', name: 'new-model' }
    ]);
  });
});
