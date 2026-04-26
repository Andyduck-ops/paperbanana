/**
 * AnimatedContainer Component
 * Generic container with configurable animations
 */

import React, { useRef, useEffect, useState } from 'react';
import { useReducedMotion } from './hooks/useReducedMotion';
import type { AnimationProps } from './types';

export interface AnimatedContainerProps extends AnimationProps {
  children: React.ReactNode;
  as?: keyof React.JSX.IntrinsicElements;
  style?: React.CSSProperties;
}

const animationKeyframes: Record<string, string> = {
  'fade-in': 'fade-in',
  'fade-out': 'fade-out',
  'fade-up': 'fade-up',
  'fade-down': 'fade-down',
  'fade-left': 'fade-left',
  'fade-right': 'fade-right',
  'slide-in-left': 'slide-in-left',
  'slide-in-right': 'slide-in-right',
  'slide-in-up': 'slide-in-up',
  'slide-in-down': 'slide-in-down',
  'scale-in': 'scale-in',
  'scale-out': 'scale-out',
  'scale-spring': 'scale-spring',
  'modal-enter': 'modal-enter',
  'modal-exit': 'modal-exit',
  'toast-enter': 'toast-enter',
  'toast-exit': 'toast-exit',
};

const durationValues: Record<string, string> = {
  fast: '0.15s',
  base: '0.3s',
  slow: '0.6s',
  slower: '0.8s',
  slowest: '1.2s',
};

const easingValues: Record<string, string> = {
  'ease-out': 'cubic-bezier(0.16, 1, 0.3, 1)',
  'ease-in-out': 'cubic-bezier(0.65, 0, 0.35, 1)',
  'ease-spring': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
  'ease-in': 'cubic-bezier(0.4, 0, 1, 1)',
  'ease-out-quart': 'cubic-bezier(0.25, 1, 0.5, 1)',
  linear: 'linear',
};

export const AnimatedContainer: React.FC<AnimatedContainerProps> = ({
  children,
  animation = 'fade-in',
  duration = 'base',
  easing = 'ease-out',
  delay = 0,
  once = true,
  triggerOnView = false,
  threshold = 0.1,
  rootMargin = '0px',
  className = '',
  onAnimationComplete,
  onAnimationStart,
  as: Component = 'div',
  style: userStyle = {},
}) => {
  const { prefersReducedMotion } = useReducedMotion();
  const elementRef = useRef<HTMLDivElement>(null);
  const [isVisible, setIsVisible] = useState(!triggerOnView);
  const hasAnimated = useRef(false);

  const durationValue = typeof duration === 'number' ? `${duration}s` : durationValues[duration] || '0.3s';
  const easingValue = easingValues[easing] || easingValues['ease-out'];

  useEffect(() => {
    if (!triggerOnView || prefersReducedMotion) {
      setIsVisible(true);
      return;
    }

    const element = elementRef.current;
    if (!element) return;

    // Skip if already animated and once=true
    if (once && hasAnimated.current) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
          hasAnimated.current = true;
          onAnimationStart?.();
          
          if (once) {
            observer.unobserve(element);
          }
        } else if (!once) {
          setIsVisible(false);
        }
      },
      { threshold, rootMargin }
    );

    observer.observe(element);

    return () => observer.disconnect();
  }, [triggerOnView, threshold, rootMargin, once, prefersReducedMotion, onAnimationStart]);

  useEffect(() => {
    if (isVisible && onAnimationComplete) {
      const timer = setTimeout(() => {
        onAnimationComplete();
      }, parseFloat(durationValue) * 1000 + delay * 1000);
      return () => clearTimeout(timer);
    }
  }, [isVisible, durationValue, delay, onAnimationComplete]);

  const animationStyle: React.CSSProperties = {
    animationName: isVisible && !prefersReducedMotion ? animationKeyframes[animation] || animation : 'none',
    animationDuration: durationValue,
    animationTimingFunction: easingValue,
    animationDelay: `${delay}s`,
    animationFillMode: 'both',
    opacity: isVisible || prefersReducedMotion ? undefined : 0,
    ...userStyle,
  };

  return React.createElement(
    Component,
    {
      ref: elementRef,
      className: `animated-container ${className}`,
      style: animationStyle,
    },
    children
  );
};

export default AnimatedContainer;
