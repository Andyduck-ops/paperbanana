import { useState, useEffect, useCallback, useRef } from 'react';

export interface NetworkStatus {
  online: boolean;
  wasOffline: boolean;
  onReconnect: (callback: () => void) => () => void;
}

export function useNetworkStatus(): NetworkStatus {
  const [online, setOnline] = useState(() => {
    if (typeof navigator !== 'undefined') {
      return navigator.onLine;
    }
    return true;
  });
  const [wasOffline, setWasOffline] = useState(false);
  const reconnectCallbacksRef = useRef<Set<() => void>>(new Set());

  const onReconnect = useCallback((callback: () => void) => {
    reconnectCallbacksRef.current.add(callback);
    return () => {
      reconnectCallbacksRef.current.delete(callback);
    };
  }, []);

  useEffect(() => {
    const handleOnline = () => {
      setOnline(true);
      if (wasOffline) {
        reconnectCallbacksRef.current.forEach((cb) => {
          try {
            cb();
          } catch {
            // ignore callback errors
          }
        });
      }
    };
    const handleOffline = () => {
      setOnline(false);
      setWasOffline(true);
    };

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, [wasOffline]);

  return { online, wasOffline, onReconnect };
}
