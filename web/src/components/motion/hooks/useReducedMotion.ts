/**
 * useReducedMotion Hook
 * Detects user's preference for reduced motion
 */

import { useState, useEffect, useCallback } from 'react';

const REDUCED_MOTION_QUERY = '(prefers-reduced-motion: reduce)';

/**
 * Hook to detect and respond to reduced motion preference
 * @returns Object containing reduced motion state and helpers
 */
export function useReducedMotion(): {
  prefersReducedMotion: boolean;
  getAnimationDuration: (normalDuration: number) => number;
  shouldAnimate: boolean;
} {
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(() => {
    if (typeof window === 'undefined') return false;
    return window.matchMedia(REDUCED_MOTION_QUERY).matches;
  });

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const mediaQuery = window.matchMedia(REDUCED_MOTION_QUERY);
    
    const handleChange = (event: MediaQueryListEvent) => {
      setPrefersReducedMotion(event.matches);
    };

    // Modern API
    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleChange);
      return () => mediaQuery.removeEventListener('change', handleChange);
    } 
    // Legacy API fallback
    else {
      mediaQuery.addListener(handleChange);
      return () => mediaQuery.removeListener(handleChange);
    }
  }, []);

  /**
   * Get appropriate animation duration based on reduced motion preference
   */
  const getAnimationDuration = useCallback(
    (normalDuration: number): number => {
      return prefersReducedMotion ? 0.001 : normalDuration;
    },
    [prefersReducedMotion]
  );

  return {
    prefersReducedMotion,
    getAnimationDuration,
    shouldAnimate: !prefersReducedMotion,
  };
}

export default useReducedMotion;
