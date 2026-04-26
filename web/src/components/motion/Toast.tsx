/**
 * Toast Component
 * Animated toast notification
 */

import React, { useEffect, useState, useCallback } from 'react';
import { useReducedMotion } from './hooks/useReducedMotion';
import type { ToastProps } from './types';

const typeStyles: Record<string, { bg: string; border: string; icon: string }> = {
  info: {
    bg: 'bg-blue-50 dark:bg-blue-900/20',
    border: 'border-blue-200 dark:border-blue-800',
    icon: 'text-blue-500',
  },
  success: {
    bg: 'bg-green-50 dark:bg-green-900/20',
    border: 'border-green-200 dark:border-green-800',
    icon: 'text-green-500',
  },
  warning: {
    bg: 'bg-yellow-50 dark:bg-yellow-900/20',
    border: 'border-yellow-200 dark:border-yellow-800',
    icon: 'text-yellow-500',
  },
  error: {
    bg: 'bg-red-50 dark:bg-red-900/20',
    border: 'border-red-200 dark:border-red-800',
    icon: 'text-red-500',
  },
};

const icons: Record<string, React.ReactNode> = {
  info: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  success: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
    </svg>
  ),
  warning: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
    </svg>
  ),
  error: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
    </svg>
  ),
};

export const Toast: React.FC<ToastProps> = ({
  message,
  type = 'info',
  duration = 5000,
  position = 'top-right',
  isOpen,
  onClose,
  actionLabel,
  onAction,
}) => {
  const { prefersReducedMotion } = useReducedMotion();
  const [isExiting, setIsExiting] = useState(false);
  const [isVisible, setIsVisible] = useState(false);

  const handleClose = useCallback(() => {
    setIsExiting(true);
    setTimeout(() => {
      setIsExiting(false);
      setIsVisible(false);
      onClose();
    }, prefersReducedMotion ? 0 : 300);
  }, [onClose, prefersReducedMotion]);

  useEffect(() => {
    if (isOpen) {
      setIsVisible(true);
      setIsExiting(false);

      if (duration > 0) {
        const timer = setTimeout(handleClose, duration);
        return () => clearTimeout(timer);
      }
    }
  }, [isOpen, duration, handleClose]);

  if (!isOpen && !isVisible) return null;

  const positionClasses: Record<string, string> = {
    'top-right': 'top-4 right-4',
    'top-left': 'top-4 left-4',
    'bottom-right': 'bottom-4 right-4',
    'bottom-left': 'bottom-4 left-4',
  };

  const animationStyles = prefersReducedMotion
    ? { opacity: isExiting ? 0 : 1 }
    : {
        opacity: isExiting ? 0 : 1,
        transform: isExiting 
          ? 'translateX(100%) scale(0.9)' 
          : 'translateX(0) scale(1)',
        transition: 'opacity 0.3s ease-out, transform 0.3s ease-out',
      };

  const styles = typeStyles[type];

  return (
    <div
      className={`fixed z-[9999] ${positionClasses[position]}`}
      style={animationStyles}
      role="alert"
      aria-live="polite"
    >
      <div
        className={`flex items-start gap-3 px-4 py-3 rounded-lg border shadow-lg min-w-[300px] max-w-md ${styles.bg} ${styles.border}`}
      >
        <div className={`flex-shrink-0 ${styles.icon}`}>
          {icons[type]}
        </div>
        
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
            {message}
          </p>
        </div>

        <div className="flex items-center gap-2">
          {actionLabel && onAction && (
            <button
              onClick={onAction}
              className="text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
            >
              {actionLabel}
            </button>
          )}
          <button
            onClick={handleClose}
            className="p-1 rounded-md hover:bg-black/5 dark:hover:bg-white/10 transition-colors"
            aria-label="Close notification"
          >
            <svg className="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
};

export default Toast;
