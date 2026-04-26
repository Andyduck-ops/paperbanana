/**
 * PaperBanana Motion Types
 * Type definitions for animation system
 */

export type AnimationType =
  | 'fade-in'
  | 'fade-out'
  | 'fade-up'
  | 'fade-down'
  | 'fade-left'
  | 'fade-right'
  | 'slide-in-left'
  | 'slide-in-right'
  | 'slide-in-up'
  | 'slide-in-down'
  | 'scale-in'
  | 'scale-out'
  | 'scale-spring'
  | 'modal-enter'
  | 'modal-exit'
  | 'toast-enter'
  | 'toast-exit';

export type AnimationDirection = 'up' | 'down' | 'left' | 'right';

export type AnimationEasing =
  | 'ease-out'
  | 'ease-in-out'
  | 'ease-spring'
  | 'ease-in'
  | 'ease-out-quart'
  | 'linear';

export type AnimationDuration =
  | 'fast'
  | 'base'
  | 'slow'
  | 'slower'
  | 'slowest'
  | number;

export interface AnimationProps {
  /** Animation type */
  animation?: AnimationType;
  /** Animation duration - token name or seconds */
  duration?: AnimationDuration;
  /** Animation easing function */
  easing?: AnimationEasing;
  /** Animation delay in seconds */
  delay?: number;
  /** Whether animation should only play once */
  once?: boolean;
  /** Whether to trigger animation when element enters viewport */
  triggerOnView?: boolean;
  /** Intersection threshold for viewport trigger (0-1) */
  threshold?: number;
  /** Root margin for intersection observer */
  rootMargin?: string;
  /** Custom class name */
  className?: string;
  /** Animation complete callback */
  onAnimationComplete?: () => void;
  /** Animation start callback */
  onAnimationStart?: () => void;
}

export interface RevealOptions {
  /** Direction to reveal from */
  direction?: AnimationDirection;
  /** Distance to travel (px) */
  distance?: number;
  /** Animation duration */
  duration?: AnimationDuration;
  /** Animation easing */
  easing?: AnimationEasing;
  /** Delay before animation */
  delay?: number;
  /** Threshold for triggering */
  threshold?: number;
  /** Root margin for intersection observer */
  rootMargin?: string;
  /** Whether to trigger only once */
  once?: boolean;
}

export interface StaggerOptions {
  /** Stagger increment in seconds */
  increment?: number;
  /** Maximum items to stagger */
  maxItems?: number;
  /** Base delay before starting stagger */
  baseDelay?: number;
  /** Child animation duration */
  duration?: AnimationDuration;
  /** Child animation easing */
  easing?: AnimationEasing;
}

export interface SkeletonProps {
  /** Type of skeleton */
  variant?: 'text' | 'circular' | 'rectangular' | 'rounded';
  /** Width of skeleton */
  width?: string | number;
  /** Height of skeleton */
  height?: string | number;
  /** Whether to animate shimmer */
  animated?: boolean;
  /** Number of lines for text variant */
  lines?: number;
  /** Custom class name */
  className?: string;
}

export interface ToastProps {
  /** Toast message */
  message: string;
  /** Toast type */
  type?: 'info' | 'success' | 'warning' | 'error';
  /** Duration in ms before auto-dismiss */
  duration?: number;
  /** Position on screen */
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left';
  /** Whether toast is visible */
  isOpen: boolean;
  /** Close callback */
  onClose: () => void;
  /** Action button text */
  actionLabel?: string;
  /** Action button callback */
  onAction?: () => void;
}

export interface ProgressBarProps {
  /** Progress value (0-100) */
  value?: number;
  /** Whether progress is indeterminate */
  indeterminate?: boolean;
  /** Whether to show striped animation */
  striped?: boolean;
  /** Size of progress bar */
  size?: 'sm' | 'md' | 'lg';
  /** Color variant */
  color?: 'primary' | 'success' | 'warning' | 'error';
  /** Custom class name */
  className?: string;
}

export interface ModalProps {
  /** Whether modal is open */
  isOpen: boolean;
  /** Close callback */
  onClose: () => void;
  /** Modal title */
  title?: string;
  /** Modal content */
  children: React.ReactNode;
  /** Whether to close on backdrop click */
  closeOnBackdrop?: boolean;
  /** Whether to close on escape key */
  closeOnEscape?: boolean;
  /** Modal size */
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  /** Custom class name */
  className?: string;
  /** Animation complete callback */
  onAnimationComplete?: () => void;
}
