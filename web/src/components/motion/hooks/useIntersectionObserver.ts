/**
 * useIntersectionObserver Hook
 * Observe element visibility in viewport
 */

import { useEffect, useRef, useState, type RefObject } from 'react';

export interface UseIntersectionObserverOptions {
  /** Element threshold for triggering (0-1) */
  threshold?: number;
  /** Root margin around viewport */
  rootMargin?: string;
  /** Whether to trigger only once */
  triggerOnce?: boolean;
  /** Root element for intersection */
  root?: Element | null;
}

/**
 * Hook to observe element intersection with viewport
 * @param options - Intersection observer options
 * @returns Tuple of [ref, isIntersecting, entry]
 */
export function useIntersectionObserver<T extends Element = HTMLDivElement>(
  options: UseIntersectionObserverOptions = {}
): [RefObject<T | null>, boolean, IntersectionObserverEntry | null] {
  const {
    threshold = 0.1,
    rootMargin = '0px',
    triggerOnce = false,
    root = null,
  } = options;

  const elementRef = useRef<T>(null);
  const [isIntersecting, setIsIntersecting] = useState(false);
  const [entry, setEntry] = useState<IntersectionObserverEntry | null>(null);
  const hasTriggered = useRef(false);

  useEffect(() => {
    const element = elementRef.current;
    if (!element) return;

    // Skip if already triggered once
    if (triggerOnce && hasTriggered.current) return;

    const observer = new IntersectionObserver(
      ([observedEntry]) => {
        setEntry(observedEntry);
        setIsIntersecting(observedEntry.isIntersecting);

        if (observedEntry.isIntersecting) {
          hasTriggered.current = true;
          
          if (triggerOnce) {
            observer.unobserve(element);
          }
        }
      },
      { threshold, rootMargin, root }
    );

    observer.observe(element);

    return () => {
      observer.disconnect();
    };
  }, [threshold, rootMargin, triggerOnce, root]);

  return [elementRef, isIntersecting, entry];
}

export default useIntersectionObserver;
