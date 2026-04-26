/**
 * ScaleIn Component
 * Scale animation wrapper with spring option
 */

import React from 'react';
import { AnimatedContainer } from './AnimatedContainer';
import type { AnimationProps } from './types';

export interface ScaleInProps extends Omit<AnimationProps, 'animation'> {
  children: React.ReactNode;
  /** Whether to use spring animation */
  spring?: boolean;
  /** Custom class name */
  className?: string;
  /** Custom styles */
  style?: React.CSSProperties;
}

export const ScaleIn: React.FC<ScaleInProps> = ({
  children,
  spring = false,
  duration = 'slow',
  easing = 'ease-out',
  delay = 0,
  once = true,
  triggerOnView = true,
  threshold = 0.2,
  rootMargin = '0px 0px -100px 0px',
  className = '',
  onAnimationComplete,
  onAnimationStart,
  style,
}) => {
  const animationType = spring ? 'scale-spring' : 'scale-in';
  const effectiveEasing = spring ? 'ease-spring' : easing;

  return (
    <AnimatedContainer
      animation={animationType}
      duration={duration}
      easing={effectiveEasing}
      delay={delay}
      once={once}
      triggerOnView={triggerOnView}
      threshold={threshold}
      rootMargin={rootMargin}
      className={className}
      onAnimationComplete={onAnimationComplete}
      onAnimationStart={onAnimationStart}
      style={style}
    >
      {children}
    </AnimatedContainer>
  );
};

export default ScaleIn;
