/**
 * Modal Component
 * Animated modal dialog with backdrop
 */

import React, { useEffect, useState, useCallback } from 'react';
import { useReducedMotion } from './hooks/useReducedMotion';
import type { ModalProps } from './types';

const sizeClasses: Record<string, string> = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  full: 'max-w-full mx-4',
};

export const Modal: React.FC<ModalProps> = ({
  isOpen,
  onClose,
  title,
  children,
  closeOnBackdrop = true,
  closeOnEscape = true,
  size = 'md',
  className = '',
  onAnimationComplete,
}) => {
  const { prefersReducedMotion } = useReducedMotion();
  const [isClosing, setIsClosing] = useState(false);
  const [isVisible, setIsVisible] = useState(false);

  const handleClose = useCallback(() => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      setIsVisible(false);
      onClose();
      onAnimationComplete?.();
    }, prefersReducedMotion ? 0 : 300);
  }, [onClose, onAnimationComplete, prefersReducedMotion]);

  useEffect(() => {
    if (isOpen) {
      setIsVisible(true);
      setIsClosing(false);
    }
  }, [isOpen]);

  useEffect(() => {
    if (!closeOnEscape) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        handleClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, closeOnEscape, handleClose]);

  // Prevent body scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      const originalOverflow = document.body.style.overflow;
      document.body.style.overflow = 'hidden';
      return () => {
        document.body.style.overflow = originalOverflow;
      };
    }
  }, [isOpen]);

  if (!isOpen && !isVisible) return null;

  const backdropAnimation = prefersReducedMotion
    ? { opacity: isClosing ? 0 : 1 }
    : {
        opacity: isClosing ? 0 : 1,
        transition: 'opacity 0.3s ease-out',
      };

  const contentAnimation = prefersReducedMotion
    ? { opacity: isClosing ? 0 : 1, transform: 'none' }
    : {
        opacity: isClosing ? 0 : 1,
        transform: isClosing ? 'scale(0.95) translateY(-10px)' : 'scale(1) translateY(0)',
        transition: 'opacity 0.3s ease-out, transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1)',
      };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      role="dialog"
      aria-modal="true"
      aria-labelledby={title ? 'modal-title' : undefined}
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        style={backdropAnimation}
        onClick={closeOnBackdrop ? handleClose : undefined}
        aria-hidden="true"
      />

      {/* Modal Content */}
      <div
        className={`relative z-10 w-full ${sizeClasses[size]} bg-white dark:bg-gray-900 rounded-lg shadow-xl ${className}`}
        style={contentAnimation}
        onClick={(e) => e.stopPropagation()}
      >
        {title && (
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <h2 id="modal-title" className="text-lg font-semibold">
              {title}
            </h2>
            <button
              onClick={handleClose}
              className="p-1 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              aria-label="Close modal"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        )}
        <div className="p-6">{children}</div>
      </div>
    </div>
  );
};

export default Modal;
