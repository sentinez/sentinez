'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { handleStatusError } from '@/lib/error-handler';
import axios from 'axios';

/**
 * A standard hook for managing data fetching with automatic error toasts and request cancellation.
 */
export function useApi<T>(apiFn: (...args: any[]) => Promise<T>, ...args: any[]) {
  const [data, setData] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<any>(null);

  // Use a ref to store the previous args to avoid unnecessary re-fetches
  const argsRef = useRef(args);
  const argsChanged = JSON.stringify(args) !== JSON.stringify(argsRef.current);

  if (argsChanged) {
    argsRef.current = args;
  }

  const fetchData = useCallback(
    async (signal?: AbortSignal) => {
      setIsLoading(true);
      setError(null);
      try {
        // Pass the signal to apiFn by appending it to the arguments
        const result = await apiFn(...argsRef.current, { signal });

        if (!result) {
          if (!signal?.aborted) {
            handleStatusError({ response: { status: 404 } });
            setError(new Error('Not Found'));
          }
        } else {
          setData(result);
        }
      } catch (err: any) {
        if (axios.isCancel(err) || signal?.aborted) {
          // Silent abort
          return;
        }
        handleStatusError(err);
        setError(err);
      } finally {
        if (!signal?.aborted) {
          setIsLoading(false);
        }
      }
    },
    [apiFn],
  );

  useEffect(() => {
    const controller = new AbortController();

    fetchData(controller.signal);

    return () => {
      controller.abort();
    };
  }, [fetchData, argsChanged]);

  return { data, isLoading, error, refresh: fetchData };
}
