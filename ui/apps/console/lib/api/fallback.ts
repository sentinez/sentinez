import axios from 'axios';

// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses (lib/faker)
// when an API call fails
export const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

/**
 * The withFallback of a service: a call that answers with fallback() when it
 * fails, if USE_SAMPLE is set; a cancellation is always thrown.
 */
export function fallbackOf(service: string) {
  return function withFallback<T>(label: string, fallback: () => T) {
    return async (call: () => Promise<T>): Promise<T> => {
      try {
        return await call();
      } catch (err: unknown) {
        if (!USE_SAMPLE || axios.isCancel(err)) throw err;
        const message = err instanceof Error ? err.message : err;
        console.warn(`[${service}] ${label} failed, using sample response:`, message);
        return fallback();
      }
    };
  };
}
