/**
 * useAnimatedValue Hook
 * Animate numeric values with easing
 */

import { useState, useEffect, useRef, useCallback } from 'react';
import { useReducedMotion } from './useReducedMotion';

type EasingFunction = (t: number) => number;

const easings: Record<string, EasingFunction> = {
  linear: (t) => t,
  'ease-out': (t) => 1 - Math.pow(1 - t, 3),
  'ease-in': (t) => t * t * t,
  'ease-in-out': (t) => t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2,
  'ease-spring': (t) => {
    // Spring-like easing with overshoot
    const c4 = (2 * Math.PI) / 3;
    return t === 0 ? 0 : t === 1 ? 1 : Math.pow(2, -10 * t) * Math.sin((t * 10 - 0.75) * c4) + 1;
  },
};

export interface UseAnimatedValueOptions {
  /** Target value */
  target: number;
  /** Animation duration in seconds */
  duration?: number;
  /** Easing function name or custom function */
  easing?: string | EasingFunction;
  /** Delay before starting animation */
  delay?: number;
  /** Whether to start animation immediately */
  immediate?: boolean;
  /** Value precision for rounding */
  precision?: number;
  /** Callback when animation completes */
  onComplete?: () => void;
}

/**
 * Hook to animate a numeric value
 * @param options - Animation options
 * @returns Tuple of [currentValue, setTarget, isAnimating]
 */
export function useAnimatedValue(
  options: UseAnimatedValueOptions
): [number, (target: number) => void, boolean] {
  const {
    target: initialTarget,
    duration = 0.6,
    easing = 'ease-out',
    delay = 0,
    immediate = true,
    precision = 2,
    onComplete,
  } = options;

  const { prefersReducedMotion, getAnimationDuration } = useReducedMotion();
  const [currentValue, setCurrentValue] = useState(0);
  const [targetValue, setTargetValue] = useState(initialTarget);
  const [isAnimating, setIsAnimating] = useState(false);
  
  const startTimeRef = useRef<number | null>(null);
  const startValueRef = useRef(0);
  const animationFrameRef = useRef<number | null>(null);

  const effectiveDuration = getAnimationDuration(duration);
  const easingFn = typeof easing === 'string' ? easings[easing] || easings['ease-out'] : easing;

  const animate = useCallback((timestamp: number) => {
    if (startTimeRef.current === null) {
      startTimeRef.current = timestamp;
    }

    const elapsed = timestamp - startTimeRef.current;
    const progress = Math.min(elapsed / (effectiveDuration * 1000), 1);
    const easedProgress = easingFn(progress);
    
    const newValue = startValueRef.current + (targetValue - startValueRef.current) * easedProgress;
    setCurrentValue(Number(newValue.toFixed(precision)));

    if (progress < 1) {
      animationFrameRef.current = requestAnimationFrame(animate);
    } else {
      setIsAnimating(false);
      setCurrentValue(Number(targetValue.toFixed(precision)));
      onComplete?.();
    }
  }, [targetValue, effectiveDuration, easingFn, precision, onComplete]);

  const startAnimation = useCallback(() => {
    if (prefersReducedMotion) {
      setCurrentValue(targetValue);
      return;
    }

    // Cancel any existing animation
    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
    }

    startValueRef.current = currentValue;
    startTimeRef.current = null;
    setIsAnimating(true);

    if (delay > 0) {
      setTimeout(() => {
        animationFrameRef.current = requestAnimationFrame(animate);
      }, delay * 1000);
    } else {
      animationFrameRef.current = requestAnimationFrame(animate);
    }
  }, [targetValue, currentValue, delay, prefersReducedMotion, animate]);

  useEffect(() => {
    if (immediate) {
      startAnimation();
    }

    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
    };
  }, [targetValue, immediate]);

  const setTarget = useCallback((newTarget: number) => {
    setTargetValue(newTarget);
  }, []);

  return [currentValue, setTarget, isAnimating];
}

export default useAnimatedValue;
