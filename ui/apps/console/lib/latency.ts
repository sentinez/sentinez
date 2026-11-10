import type { useFormatter } from 'next-intl';

/** A latency in microseconds, as milliseconds with a decimal, e.g. "2.5 ms" */
export function latencyMs(format: ReturnType<typeof useFormatter>, micros: number): string {
  return `${format.number(micros / 1000, { maximumFractionDigits: 1 })} ms`;
}
