/**
 * Skeleton Component
 * Loading placeholder with shimmer animation
 */

import React from 'react';
import { useReducedMotion } from './hooks/useReducedMotion';
import type { SkeletonProps } from './types';

export interface SkeletonComponentProps extends SkeletonProps {
  /** Custom class name */
  className?: string;
}

export const Skeleton: React.FC<SkeletonComponentProps> = ({
  variant = 'text',
  width,
  height,
  animated = true,
  lines = 3,
  className = '',
}) => {
  const { prefersReducedMotion } = useReducedMotion();

  const baseStyles: React.CSSProperties = {
    backgroundColor: 'var(--theme-muted, oklch(0.93 0.01 90))',
    borderRadius: variant === 'circular' ? '50%' : variant === 'rounded' ? '8px' : '4px',
  };

  const shimmerStyles: React.CSSProperties = animated && !prefersReducedMotion
    ? {
        background: `linear-gradient(
          90deg,
          var(--theme-muted, oklch(0.93 0.01 90)) 25%,
          var(--theme-background, oklch(0.97 0.01 90)) 50%,
          var(--theme-muted, oklch(0.93 0.01 90)) 75%
        )`,
        backgroundSize: '200% 100%',
        animation: 'shimmer 1.5s linear infinite',
      }
    : {};

  const getDimensions = (): React.CSSProperties => {
    const styles: React.CSSProperties = {};
    
    if (width !== undefined) {
      styles.width = typeof width === 'number' ? `${width}px` : width;
    }
    if (height !== undefined) {
      styles.height = typeof height === 'number' ? `${height}px` : height;
    }

    // Default dimensions for different variants
    if (variant === 'text' && height === undefined) {
      styles.height = '1em';
    }
    if (variant === 'circular' && width === undefined && height === undefined) {
      styles.width = '40px';
      styles.height = '40px';
    }

    return styles;
  };

  const renderSkeleton = (key?: number) => (
    <div
      key={key}
      className={`skeleton skeleton-${variant} ${className}`}
      style={{
        ...baseStyles,
        ...shimmerStyles,
        ...getDimensions(),
        marginBottom: variant === 'text' ? '0.5em' : undefined,
        width: variant === 'text' && key === lines - 1 ? '80%' : getDimensions().width,
      }}
    />
  );

  if (variant === 'text' && lines > 1) {
    return (
      <div className="skeleton-text-group">
        {Array.from({ length: lines }, (_, i) => renderSkeleton(i))}
      </div>
    );
  }

  return renderSkeleton();
};

export default Skeleton;
