/**
 * Spinner Component
 * Loading spinner with size variants
 */

import React from 'react';
import { useReducedMotion } from './hooks/useReducedMotion';

export interface SpinnerProps {
  /** Spinner size */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  /** Custom class name */
  className?: string;
  /** Color variant */
  color?: 'primary' | 'muted' | 'white';
  /** Label for screen readers */
  label?: string;
}

const sizeClasses: Record<string, string> = {
  xs: 'w-3 h-3',
  sm: 'w-4 h-4',
  md: 'w-6 h-6',
  lg: 'w-8 h-8',
  xl: 'w-12 h-12',
};

const colorClasses: Record<string, string> = {
  primary: 'text-[var(--theme-primary,oklch(0.46_0.1_40))]',
  muted: 'text-[var(--theme-muted-foreground,oklch(0.48_0.03_60))]',
  white: 'text-white',
};

export const Spinner: React.FC<SpinnerProps> = ({
  size = 'md',
  color = 'primary',
  className = '',
  label = 'Loading...',
}) => {
  const { prefersReducedMotion } = useReducedMotion();

  if (prefersReducedMotion) {
    return (
      <span className={`${sizeClasses[size]} ${colorClasses[color]} ${className}`} role="status">
        <span className="sr-only">{label}</span>
        <svg
          className="w-full h-full"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
          />
        </svg>
      </span>
    );
  }

  return (
    <span className={`${sizeClasses[size]} ${colorClasses[color]} ${className}`} role="status">
      <span className="sr-only">{label}</span>
      <svg
        className="w-full h-full animate-spin"
        fill="none"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="4"
        />
        <path
          className="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
        />
      </svg>
    </span>
  );
};

export default Spinner;
