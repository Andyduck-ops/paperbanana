/**
 * SlideIn Component
 * Slide animation wrapper with direction support
 */

import React from 'react';
import { AnimatedContainer } from './AnimatedContainer';
import type { AnimationProps, AnimationDirection, AnimationType } from './types';

export interface SlideInProps extends Omit<AnimationProps, 'animation'> {
  children: React.ReactNode;
  /** Slide direction */
  direction?: AnimationDirection;
  /** Custom class name */
  className?: string;
  /** Custom styles */
  style?: React.CSSProperties;
}

export const SlideIn: React.FC<SlideInProps> = ({
  children,
  direction = 'up',
  duration = 'slower',
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
        return 'slide-in-up';
      case 'down':
        return 'slide-in-down';
      case 'left':
        return 'slide-in-left';
      case 'right':
        return 'slide-in-right';
      default:
        return 'slide-in-up';
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

export default SlideIn;
