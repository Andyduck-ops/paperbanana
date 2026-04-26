import { useEffect, useRef } from 'react';

const FOCUSABLE_SELECTORS = [
  'button:not([disabled])',
  'a[href]',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(', ');

export function useFocusTrap<T extends HTMLElement>(isActive: boolean) {
  const containerRef = useRef<T>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!isActive) return;

    // Store previous focus
    previousFocusRef.current = document.activeElement as HTMLElement;

    // Focus first focusable element inside container
    const container = containerRef.current;
    if (!container) return;

    const focusableElements = container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTORS);
    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    // Focus first element after a short delay to ensure render is complete
    const timer = setTimeout(() => {
      firstElement?.focus();
    }, 0);

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;

      // If no focusable elements, prevent tab
      if (focusableElements.length === 0) {
        e.preventDefault();
        return;
      }

      // Shift + Tab on first element -> wrap to last
      if (e.shiftKey && document.activeElement === firstElement) {
        e.preventDefault();
        lastElement?.focus();
        return;
      }

      // Tab on last element -> wrap to first
      if (!e.shiftKey && document.activeElement === lastElement) {
        e.preventDefault();
        firstElement?.focus();
        return;
      }
    };

    document.addEventListener('keydown', handleKeyDown);

    return () => {
      clearTimeout(timer);
      document.removeEventListener('keydown', handleKeyDown);
      // Restore previous focus on cleanup
      previousFocusRef.current?.focus();
    };
  }, [isActive]);

  return containerRef;
}
