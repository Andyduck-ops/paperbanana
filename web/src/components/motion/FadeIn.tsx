/**
 * FadeIn Component
 * Fade animation wrapper with direction support
 */

import React from 'react';
import { AnimatedContainer } from './AnimatedContainer';
import type { AnimationProps, AnimationDirection, AnimationType } from './types';

export interface FadeInProps extends Omit<AnimationProps, 'animation'> {
  children: React.ReactNode;
  /** Fade direction */
  direction?: AnimationDirection | 'none';
  /** Custom class name */
  className?: string;
  /** Custom styles */
  style?: React.CSSProperties;
}

export const FadeIn: React.FC<FadeInProps> = ({
  children,
  direction = 'none',
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
  const getAnimationType = (): AnimationType => {
    switch (direction) {
      case 'up':
        return 'fade-up';
      case 'down':
        return 'fade-down';
      case 'left':
        return 'fade-left';
      case 'right':
        return 'fade-right';
      default:
        return 'fade-in';
    }
  };

  return (
    <AnimatedContainer
      animation={getAnimationType()}
      duration={duration}
      easing={easing}
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

export default FadeIn;
