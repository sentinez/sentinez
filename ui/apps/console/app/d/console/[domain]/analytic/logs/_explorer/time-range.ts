import { HistogramResolution } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import type { DateTimeFormatOptions } from 'next-intl';

import { RESOLUTION_BUCKETS, resolutionOfSpan, spanOf } from '@/lib/api/analytic';

// QuickRange a relative range of the picker, the resolution of its histogram
export type QuickRange = 'last60s' | 'last60m' | 'last24h';

export const QUICK_RANGES: Record<QuickRange, HistogramResolution> = {
  last60s: HistogramResolution.HISTOGRAM_RESOLUTION_SECOND,
  last60m: HistogramResolution.HISTOGRAM_RESOLUTION_MINUTE,
  last24h: HistogramResolution.HISTOGRAM_RESOLUTION_HOUR,
};

/**
 * TimeRange the range of the page: a quick one, ending now, or an absolute
 * one picked on the calendar, from inclusive to exclusive (millis).
 */
export type TimeRange = { quick: QuickRange } | { from: number; to: number };

// Window the bounds of a range at a time, to undefined when it ends now
type Window = { from: number; to?: number };

// The second buckets are kept a week (histogram _secondRetention)
const SECOND_RETENTION_MILLIS = 7 * 24 * 60 * 60 * 1000;

/**
 * The resolution of the histogram of a range: that of a quick one, else the
 * finest whose buckets cover it (the seconds within their retention), the
 * hours beyond a day: the histogram then shows its last 24 hours only.
 */
export function resolutionOf(range: TimeRange, now: number): HistogramResolution {
  if ('quick' in range) return QUICK_RANGES[range.quick];
  const resolution = resolutionOfSpan(range.to - range.from);
  const expired = now - range.from > SECOND_RETENTION_MILLIS;
  return resolution === HistogramResolution.HISTOGRAM_RESOLUTION_SECOND && expired
    ? HistogramResolution.HISTOGRAM_RESOLUTION_MINUTE
    : resolution;
}

/** The bounds of a range: a quick one starts with its first bucket at now */
export function windowOf(range: TimeRange, now: number): Window {
  if ('from' in range) return { from: range.from, to: range.to };
  const { millis, count } = RESOLUTION_BUCKETS[QUICK_RANGES[range.quick]];
  return { from: (Math.floor(now / millis) - count + 1) * millis };
}

/** Whether the histogram of a range shows only its end (longer than a day) */
export function isClipped(range: TimeRange, now: number): boolean {
  return 'from' in range && range.to - range.from > spanOf(resolutionOf(range, now));
}

// The tick of a bucket start of each resolution
export const TICK_FORMATS: Record<HistogramResolution, DateTimeFormatOptions> = {
  [HistogramResolution.HISTOGRAM_RESOLUTION_UNSPECIFIED]: { hour: '2-digit', minute: '2-digit' },
  [HistogramResolution.HISTOGRAM_RESOLUTION_SECOND]: {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  },
  [HistogramResolution.HISTOGRAM_RESOLUTION_MINUTE]: { hour: '2-digit', minute: '2-digit' },
  [HistogramResolution.HISTOGRAM_RESOLUTION_HOUR]: {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  },
  [HistogramResolution.UNRECOGNIZED]: { hour: '2-digit', minute: '2-digit' },
};

// "HH:mm:ss" of a local time, the value of an <input type="time" step="1">
export function timeOf(date: Date): string {
  return [date.getHours(), date.getMinutes(), date.getSeconds()]
    .map((n) => String(n).padStart(2, '0'))
    .join(':');
}

// A day at a local "HH:mm[:ss]" time
export function atTime(day: Date, time: string): Date {
  const [h = 0, m = 0, s = 0] = time.split(':').map(Number);
  const date = new Date(day);
  date.setHours(h, m, s, 0);
  return date;
}

// A date and time with its zone, e.g. "08/11/2026, 10:05:03 GMT+7"
export const DATE_TIME_ZONE: DateTimeFormatOptions = { dateStyle: 'short', timeStyle: 'long' };

// The short name of the local zone at a date, e.g. "GMT+7"
export function timeZoneOf(locale: string, date: Date = new Date()): string {
  return (
    new Intl.DateTimeFormat(locale, { timeZoneName: 'short' })
      .formatToParts(date)
      .find((p) => p.type === 'timeZoneName')?.value ?? ''
  );
}
