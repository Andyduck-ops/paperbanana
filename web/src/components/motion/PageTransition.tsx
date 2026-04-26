/**
 * PageTransition Component
 * Wrap page content with enter/exit animations
 */

import React, { useEffect, useState, useRef } from 'react';
import { useReducedMotion } from './hooks/useReducedMotion';

export interface PageTransitionProps {
  children: React.ReactNode;
  /** Unique key for the page */
  pageKey: string;
  /** Animation direction */
  direction?: 'up' | 'down' | 'fade';
  /** Transition duration in seconds */
  duration?: number;
  /** Custom class name */
  className?: string;
  /** Exit callback */
  onExitComplete?: () => void;
  /** Enter callback */
  onEnterComplete?: () => void;
}

export const PageTransition: React.FC<PageTransitionProps> = ({
  children,
  pageKey,
  direction = 'up',
  duration = 0.5,
  className = '',
  onExitComplete,
  onEnterComplete,
}) => {
  const { prefersReducedMotion } = useReducedMotion();
  const [displayChildren, setDisplayChildren] = useState(children);
  const [animationState, setAnimationState] = useState<'entering' | 'entered' | 'exiting' | 'exited'>('entered');
  const prevKeyRef = useRef(pageKey);
  const prevChildrenRef = useRef(children);

  useEffect(() => {
    if (pageKey !== prevKeyRef.current) {
      // Start exit animation
      setAnimationState('exiting');
      
      const exitTimer = setTimeout(() => {
        setDisplayChildren(prevChildrenRef.current);
        setAnimationState('exited');
        onExitComplete?.();
        
        // Start enter animation
        setTimeout(() => {
          setDisplayChildren(children);
          prevChildrenRef.current = children;
          prevKeyRef.current = pageKey;
          setAnimationState('entering');
          
          const enterTimer = setTimeout(() => {
            setAnimationState('entered');
            onEnterComplete?.();
          }, duration * 1000);
          
          return () => clearTimeout(enterTimer);
        }, 50);
      }, duration * 300); // Exit is faster

      return () => clearTimeout(exitTimer);
    }
  }, [pageKey, children, duration, onExitComplete, onEnterComplete]);

  const getAnimationStyles = (): React.CSSProperties => {
    if (prefersReducedMotion) {
      return { opacity: 1, transform: 'none' };
    }

    const baseTransition = `opacity ${duration * 0.6}s ease-out, transform ${duration}s cubic-bezier(0.16, 1, 0.3, 1)`;
    
    switch (animationState) {
      case 'entering':
        return {
          opacity: 0,
          transform: direction === 'up' ? 'translateY(10px)' : 
                     direction === 'down' ? 'translateY(-10px)' : 'none',
          animation: `page-enter ${duration}s cubic-bezier(0.16, 1, 0.3, 1) forwards`,
        };
      case 'entered':
        return {
          opacity: 1,
          transform: 'none',
          transition: baseTransition,
        };
      case 'exiting':
        return {
          opacity: 0,
          transform: direction === 'up' ? 'translateY(-10px)' : 
                     direction === 'down' ? 'translateY(10px)' : 'none',
          transition: `opacity ${duration * 0.3}s ease-in, transform ${duration * 0.3}s ease-in`,
        };
      case 'exited':
        return {
          opacity: 0,
          transform: 'none',
        };
      default:
        return {};
    }
  };

  return (
    <div
      className={`page-transition ${className}`}
      style={{
        ...getAnimationStyles(),
        willChange: animationState === 'entering' || animationState === 'exiting' 
          ? 'opacity, transform' 
          : undefined,
      }}
    >
      {displayChildren}
    </div>
  );
};

export default PageTransition;
