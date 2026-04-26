/**
 * useScrollReveal Hook
 * Trigger animations when elements scroll into view
 */

import { useEffect, useState, useCallback, type RefObject } from 'react';
import { useIntersectionObserver } from './useIntersectionObserver';
import { useReducedMotion } from './useReducedMotion';
import type { RevealOptions } from '../types';

/**
 * Hook to reveal elements on scroll
 * @param options - Reveal animation options
 * @returns Tuple of [ref, isVisible, style]
 */
export function useScrollReveal<T extends HTMLElement = HTMLDivElement>(
  options: RevealOptions = {}
): [
  RefObject<T | null>,
  boolean,
  React.CSSProperties
] {
  const {
    direction = 'up',
    distance = 20,
    duration = 'slow',
    easing = 'ease-out',
    delay = 0,
    threshold = 0.2,
    rootMargin = '0px 0px -100px 0px',
    once = true,
  } = options;

  const { prefersReducedMotion } = useReducedMotion();
  const [ref, isIntersecting] = useIntersectionObserver<T>({
    threshold,
    rootMargin,
    triggerOnce: once,
  });
  const [isVisible, setIsVisible] = useState(false);

  // Get duration in seconds
  const durationMap: Record<string, number> = {
    fast: 0.15,
    base: 0.3,
    slow: 0.6,
    slower: 0.8,
    slowest: 1.2,
  };
  const durationValue = typeof duration === 'number' ? duration : durationMap[duration] || 0.6;

  // Get easing value
  const easingMap: Record<string, string> = {
    'ease-out': 'cubic-bezier(0.16, 1, 0.3, 1)',
    'ease-in-out': 'cubic-bezier(0.65, 0, 0.35, 1)',
    'ease-spring': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
    'ease-in': 'cubic-bezier(0.4, 0, 1, 1)',
    'ease-out-quart': 'cubic-bezier(0.25, 1, 0.5, 1)',
    linear: 'linear',
  };
  const easingValue = easingMap[easing] || easingMap['ease-out'];

  useEffect(() => {
    if (isIntersecting) {
      // Small delay to ensure DOM is ready
      const timer = setTimeout(() => {
        setIsVisible(true);
      }, 50);
      return () => clearTimeout(timer);
    } else if (!once) {
      setIsVisible(false);
    }
  }, [isIntersecting, once]);

  // Calculate initial transform based on direction
  const getInitialTransform = useCallback(() => {
    switch (direction) {
      case 'up':
        return `translateY(${distance}px)`;
      case 'down':
        return `translateY(-${distance}px)`;
      case 'left':
        return `translateX(${distance}px)`;
      case 'right':
        return `translateX(-${distance}px)`;
      default:
        return `translateY(${distance}px)`;
    }
  }, [direction, distance]);

  // Generate style
  const style: React.CSSProperties = {
    opacity: isVisible || prefersReducedMotion ? 1 : 0,
    transform: isVisible || prefersReducedMotion ? 'none' : getInitialTransform(),
    transition: prefersReducedMotion
      ? 'none'
      : `opacity ${durationValue}s ${easingValue} ${delay}s, transform ${durationValue}s ${easingValue} ${delay}s`,
    willChange: isVisible ? undefined : 'opacity, transform',
  };

  return [ref, isVisible, style];
}

export default useScrollReveal;
