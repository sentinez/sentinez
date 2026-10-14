'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { handleStatusError } from '@/lib/error-handler';
import axios from 'axios';
import { useTranslations } from 'next-intl';

/**
 * A standard hook for managing data fetching with automatic error toasts and request cancellation.
 */
export function useApi<T>(apiFn: (...args: any[]) => Promise<T>, ...args: any[]) {
  const t = useTranslations('Errors');
  const [data, setData] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(args[0] !== null);
  const [error, setError] = useState<any>(null);

  // Refetch only when the serialized args change, not on every new object.
  // The ref is synced in an effect (declared before the fetch effect so it
  // runs first): mutating it during render breaks under Strict Mode, where
  // the second render saw "unchanged" args and the fetch never re-ran.
  const argsKey = JSON.stringify(args);
  const argsRef = useRef(args);
  useEffect(() => {
    argsRef.current = args;
  });

  const fetchData = useCallback(
    async (signal?: AbortSignal) => {
      if (argsRef.current[0] === null) {
        setIsLoading(false);
        return;
      }
      setIsLoading(true);
      setError(null);
      try {
        // Pass the signal to apiFn by appending it to the arguments
        const result = await apiFn(...argsRef.current, { signal });

        if (!result) {
          if (!signal?.aborted) {
            handleStatusError({ response: { status: 404 } }, t);
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
        handleStatusError(err, t);
        setError(err);
      } finally {
        if (!signal?.aborted) {
          setIsLoading(false);
        }
      }
    },
    [apiFn, t],
  );

  useEffect(() => {
    const controller = new AbortController();

    fetchData(controller.signal);

    return () => {
      controller.abort();
    };
  }, [fetchData, argsKey]);

  return { data, isLoading, error, refresh: fetchData };
}
