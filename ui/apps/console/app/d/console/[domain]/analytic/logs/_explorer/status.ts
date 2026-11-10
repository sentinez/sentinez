import { headerLinesOf } from '@/lib/type/edge';

// StatusClass the class of an HTTP status, a series of the histogram of
// the requests
export type StatusClass = 's2xx' | 's3xx' | 's4xx' | 's5xx' | 'other';

export const STATUS_CLASSES: StatusClass[] = ['s2xx', 's3xx', 's4xx', 's5xx', 'other'];

// The color of each class: success, redirect, client error, server error
export const STATUS_COLORS: Record<StatusClass, string> = {
  s2xx: 'oklch(0.696 0.17 162.48)',
  s3xx: 'oklch(0.685 0.169 237.323)',
  s4xx: 'oklch(0.769 0.188 70.08)',
  s5xx: 'oklch(0.637 0.237 25.331)',
  other: 'oklch(0.705 0.015 286.067)',
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
