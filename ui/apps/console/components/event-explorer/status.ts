import { headerLinesOf } from '@/lib/type/edge';

// StatusClass the class of an HTTP status, a series of the histogram of
// the requests
export type StatusClass = 's2xx' | 's3xx' | 's4xx' | 's5xx' | 'other';

export const STATUS_CLASSES: StatusClass[] = ['s2xx', 's3xx', 's4xx', 's5xx', 'other'];

// The color of each class: success, redirect, client error, server error.
// The Tailwind palette of the theme globals.css imports, as its variables
// (Tailwind keeps those named in the sources), not copies of their values
export const STATUS_COLORS: Record<StatusClass, string> = {
  s2xx: 'var(--color-emerald-500)',
  s3xx: 'var(--color-sky-500)',
  s4xx: 'var(--color-amber-500)',
  s5xx: 'var(--color-red-500)',
  other: 'var(--color-zinc-400)',
};

// The class of a status, `other` for the informational and invalid ones
export function statusClassOf(status: number): StatusClass {
  if (status >= 200 && status < 300) return 's2xx';
  if (status >= 300 && status < 400) return 's3xx';
  if (status >= 400 && status < 500) return 's4xx';
  if (status >= 500 && status < 600) return 's5xx';
  return 'other';
}

// The headers of an event, sent as "Key: value" lines
export function headerPairsOf(headers: string): [string, string][] {
  return headerLinesOf(headers).map((h) => [h.name, h.value]);
}

// The query params of an event, sent as the raw query string
export function queryParamsOf(queries: string): [string, string][] {
  try {
    return [...new URLSearchParams(queries).entries()];
  } catch {
    return queries ? [[queries, '']] : [];
  }
}
