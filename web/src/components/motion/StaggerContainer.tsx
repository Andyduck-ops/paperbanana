/**
 * StaggerContainer Component
 * Container that staggers animations for child elements
 */

import React, { useRef, useEffect, useState, Children, cloneElement, isValidElement } from 'react';
import { useReducedMotion } from './hooks/useReducedMotion';
import type { StaggerOptions, AnimationProps } from './types';

export interface StaggerContainerProps extends StaggerOptions, Omit<AnimationProps, 'delay'> {
  children: React.ReactNode;
  /** Custom class name */
  className?: string;
  /** Custom styles */
  style?: React.CSSProperties;
  /** Wrapper element tag */
  as?: keyof React.JSX.IntrinsicElements;
}

export const StaggerContainer: React.FC<StaggerContainerProps> = ({
  children,
  increment = 0.05,
  maxItems = 8,
  baseDelay = 0,
  duration = 'slow',
  easing = 'ease-out',
  once = true,
  triggerOnView = true,
  threshold = 0.2,
  rootMargin = '0px 0px -100px 0px',
  className = '',
  style = {},
  as: Component = 'div',
}) => {
  const { prefersReducedMotion } = useReducedMotion();
  const containerRef = useRef<HTMLDivElement>(null);
  const [isVisible, setIsVisible] = useState(!triggerOnView);
  const hasTriggered = useRef(false);

  const durationMap: Record<string, number> = {
    fast: 0.15,
    base: 0.3,
    slow: 0.6,
    slower: 0.8,
    slowest: 1.2,
  };
  const durationValue = typeof duration === 'number' ? duration : durationMap[duration] || 0.6;

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
    if (!triggerOnView || prefersReducedMotion) {
      setIsVisible(true);
      return;
    }

    const container = containerRef.current;
    if (!container) return;

    if (once && hasTriggered.current) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
          hasTriggered.current = true;
          
          if (once) {
            observer.unobserve(container);
          }
        } else if (!once) {
          setIsVisible(false);
        }
      },
      { threshold, rootMargin }
    );

    observer.observe(container);

    return () => observer.disconnect();
  }, [triggerOnView, threshold, rootMargin, once, prefersReducedMotion]);

  // Apply stagger styles to children
  const staggeredChildren = Children.map(children, (child, index) => {
    if (!isValidElement(child)) return child;

    const effectiveIndex = Math.min(index, maxItems - 1);
    const delay = baseDelay + effectiveIndex * increment;

    const childStyle: React.CSSProperties = {
      opacity: isVisible || prefersReducedMotion ? 1 : 0,
      transform: isVisible || prefersReducedMotion ? 'translateY(0)' : 'translateY(20px)',
      transition: prefersReducedMotion
        ? 'none'
        : `opacity ${durationValue}s ${easingValue} ${delay}s, transform ${durationValue}s ${easingValue} ${delay}s`,
      willChange: isVisible ? undefined : 'opacity, transform',
    };

    return cloneElement(child as React.ReactElement<{ style?: React.CSSProperties }>, {
      style: {
        ...(child.props as { style?: React.CSSProperties }).style,
        ...childStyle,
      },
    });
  });

  return React.createElement(
    Component,
    {
      ref: containerRef,
      className: `stagger-container ${className}`,
      style,
    },
    staggeredChildren
  );
};

export default StaggerContainer;
