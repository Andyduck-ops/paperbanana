import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CriticSuggestions } from './CriticSuggestions';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (key === 'generate.criticSuggestions' && params?.round) {
        return `Critic Suggestions (Round ${params.round})`;
      }
      return key;
    },
  }),
}));

describe('CriticSuggestions', () => {
  it('renders nothing when no suggestions', () => {
    const { container } = render(
      <CriticSuggestions suggestions="" roundNumber={1} hasChanges={true} />
    );
    expect(container.firstChild).toBeNull();
  });

  it('shows suggestions button', () => {
    render(
      <CriticSuggestions suggestions="Add more contrast" roundNumber={1} hasChanges={true} />
    );
    expect(screen.getByText(/Critic Suggestions \(Round 1\)/)).toBeInTheDocument();
  });

  it('expands on click', () => {
    render(
      <CriticSuggestions suggestions="Add more contrast" roundNumber={1} hasChanges={true} />
    );
    const button = screen.getByRole('button');
    fireEvent.click(button);
    expect(screen.getByText('Add more contrast')).toBeInTheDocument();
  });

  it('shows "no changes needed" message when hasChanges is false', () => {
    render(
      <CriticSuggestions suggestions="No changes needed" roundNumber={1} hasChanges={false} />
    );
    const button = screen.getByRole('button');
    fireEvent.click(button);
    expect(screen.getByText('generate.noChangesNeeded')).toBeInTheDocument();
  });

  it('shows actual suggestions when hasChanges is true', () => {
    render(
      <CriticSuggestions suggestions="Improve color scheme" roundNumber={2} hasChanges={true} />
    );
    const button = screen.getByRole('button');
    fireEvent.click(button);
    expect(screen.getByText('Improve color scheme')).toBeInTheDocument();
  });

  it('collapses when clicking again', () => {
    render(
      <CriticSuggestions suggestions="Add labels" roundNumber={1} hasChanges={true} />
    );
    const button = screen.getByRole('button');
    fireEvent.click(button);
    expect(screen.getByText('Add labels')).toBeInTheDocument();
    fireEvent.click(button);
    expect(screen.queryByText('Add labels')).not.toBeInTheDocument();
  });

  it('displays correct round number', () => {
    render(
      <CriticSuggestions suggestions="Test" roundNumber={3} hasChanges={true} />
    );
    expect(screen.getByText(/Round 3/)).toBeInTheDocument();
  });

  it('shows lightbulb icon', () => {
    render(
      <CriticSuggestions suggestions="Test" roundNumber={1} hasChanges={true} />
    );
    expect(screen.getByText('💡')).toBeInTheDocument();
  });
});
