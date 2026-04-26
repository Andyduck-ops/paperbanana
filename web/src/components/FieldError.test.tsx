import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { FieldError } from './FieldError';

describe('FieldError', () => {
  it('renders nothing when error is undefined', () => {
    const { container } = render(<FieldError />);
    expect(container.firstChild).toBeNull();
  });

  it('renders nothing when error is empty string', () => {
    const { container } = render(<FieldError error="" />);
    expect(container.firstChild).toBeNull();
  });

  it('renders error message when provided', () => {
    render(<FieldError error="This field is required" />);
    expect(screen.getByText('This field is required')).toBeInTheDocument();
  });

  it('has role="alert" for accessibility', () => {
    render(<FieldError error="Invalid input" />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });

  it('applies custom className', () => {
    const { container } = render(<FieldError error="Error" className="mt-2" />);
    expect(container.firstChild).toHaveClass('mt-2');
  });

  it('applies default error styling classes', () => {
    const { container } = render(<FieldError error="Error message" />);
    const element = container.firstChild as HTMLElement;
    expect(element.className).toContain('text-red-600');
    expect(element.className).toContain('dark:text-red-400');
    expect(element.className).toContain('text-sm');
    expect(element.className).toContain('mt-1');
  });
});
