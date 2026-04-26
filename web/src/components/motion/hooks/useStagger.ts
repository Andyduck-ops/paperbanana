/**
 * useStagger Hook
 * Calculate stagger delays for child elements
 */

import { useMemo } from 'react';
import { useReducedMotion } from './useReducedMotion';

export interface UseStaggerOptions {
  /** Index of the item */
  index: number;
  /** Stagger increment in seconds */
  increment?: number;
  /** Maximum items to stagger */
  maxItems?: number;
  /** Base delay before starting */
  baseDelay?: number;
}

/**
 * Hook to calculate stagger delay for an element
 * @param options - Stagger options
 * @returns Calculated delay in seconds
 */
export function useStagger(options: UseStaggerOptions): number {
  const {
    index,
    increment = 0.05,
    maxItems = 8,
    baseDelay = 0,
  } = options;

  const { prefersReducedMotion } = useReducedMotion();

  return useMemo(() => {
    if (prefersReducedMotion) return 0;
    
    const effectiveIndex = Math.min(index, maxItems - 1);
    return baseDelay + effectiveIndex * increment;
  }, [index, increment, maxItems, baseDelay, prefersReducedMotion]);
}

export interface UseStaggerArrayOptions {
  /** Number of items */
  count: number;
  /** Stagger increment in seconds */
  increment?: number;
  /** Maximum items to stagger */
  maxItems?: number;
  /** Base delay before starting */
  baseDelay?: number;
}

/**
 * Hook to calculate stagger delays for an array of items
 * @param options - Stagger array options
 * @returns Array of delays in seconds
 */
export function useStaggerArray(options: UseStaggerArrayOptions): number[] {
  const {
    count,
    increment = 0.05,
    maxItems = 8,
    baseDelay = 0,
  } = options;

  const { prefersReducedMotion } = useReducedMotion();

  return useMemo(() => {
    if (prefersReducedMotion) {
      return Array(count).fill(0);
    }

    return Array.from({ length: count }, (_, index) => {
      const effectiveIndex = Math.min(index, maxItems - 1);
      return baseDelay + effectiveIndex * increment;
    });
  }, [count, increment, maxItems, baseDelay, prefersReducedMotion]);
}

export default useStagger;
