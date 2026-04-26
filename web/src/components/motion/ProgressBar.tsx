/**
 * ProgressBar Component
 * Animated progress indicator with variants
 */

import React from 'react';
import { useReducedMotion } from './hooks/useReducedMotion';
import type { ProgressBarProps } from './types';

const sizeClasses: Record<string, string> = {
  sm: 'h-1',
  md: 'h-2',
  lg: 'h-4',
};

const colorClasses: Record<string, string> = {
  primary: 'bg-[var(--theme-primary,oklch(0.46_0.1_40))]',
  success: 'bg-[var(--color-status-success,oklch(0.61_0.16_145))]',
  warning: 'bg-[var(--color-status-warning,oklch(0.75_0.16_85))]',
  error: 'bg-[var(--color-status-error,oklch(0.59_0.19_25))]',
};

export const ProgressBar: React.FC<ProgressBarProps> = ({
  value = 0,
  indeterminate = false,
  striped = false,
  size = 'md',
  color = 'primary',
  className = '',
}) => {
  const { prefersReducedMotion } = useReducedMotion();
  
  const clampedValue = Math.min(Math.max(value, 0), 100);
  
  const fillStyles: React.CSSProperties = indeterminate
    ? {
        width: '50%',
        animation: prefersReducedMotion ? 'none' : 'indeterminate 1.5s ease-in-out infinite',
      }
    : {
        width: `${clampedValue}%`,
        transition: prefersReducedMotion ? 'none' : 'width 0.3s ease-in',
      };

  const stripedStyles: React.CSSProperties = striped && !indeterminate
    ? {
        backgroundImage: `linear-gradient(
          45deg,
          rgba(255, 255, 255, 0.15) 25%,
          transparent 25%,
          transparent 50%,
          rgba(255, 255, 255, 0.15) 50%,
          rgba(255, 255, 255, 0.15) 75%,
          transparent 75%,
          transparent
        )`,
        backgroundSize: '1rem 1rem',
        animation: prefersReducedMotion ? 'none' : 'progress-stripes 1s linear infinite',
      }
    : {};

  return (
    <div
      className={`w-full bg-[var(--theme-muted,oklch(0.93_0.01_90))] rounded-full overflow-hidden ${sizeClasses[size]} ${className}`}
      role="progressbar"
      aria-valuenow={indeterminate ? undefined : clampedValue}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuetext={indeterminate ? 'Loading...' : `${Math.round(clampedValue)}%`}
    >
      <div
        className={`h-full ${colorClasses[color]} rounded-full`}
        style={{
          ...fillStyles,
          ...stripedStyles,
        }}
      />
    </div>
  );
};

export default ProgressBar;
