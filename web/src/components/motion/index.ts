/**
 * PaperBanana Motion Components
 * React components and hooks for animations
 */

// Components
export { AnimatedContainer } from './AnimatedContainer';
export { FadeIn } from './FadeIn';
export { SlideIn } from './SlideIn';
export { ScaleIn } from './ScaleIn';
export { StaggerContainer } from './StaggerContainer';
export { Skeleton } from './Skeleton';
export { PageTransition } from './PageTransition';
export { Modal } from './Modal';
export { Toast } from './Toast';
export { ProgressBar } from './ProgressBar';
export { Spinner } from './Spinner';

// Hooks
export {
  useIntersectionObserver,
  useScrollReveal,
  useReducedMotion,
  useAnimatedValue,
  useStagger,
  useStaggerArray,
} from './hooks';

// Types
export type {
  AnimationType,
  AnimationDirection,
  AnimationEasing,
  AnimationDuration,
  AnimationProps,
  RevealOptions,
  StaggerOptions,
  SkeletonProps,
  ToastProps,
  ProgressBarProps,
  ModalProps,
} from './types';

export type {
  UseIntersectionObserverOptions,
  UseAnimatedValueOptions,
  UseStaggerOptions,
  UseStaggerArrayOptions,
} from './hooks';
