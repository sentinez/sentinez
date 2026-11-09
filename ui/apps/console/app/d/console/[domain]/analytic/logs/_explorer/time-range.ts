import { HistogramResolution } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import type { DateTimeFormatOptions } from 'next-intl';

import { RESOLUTIONS, RESOLUTION_BUCKETS, maxPagesOf, resolutionOfSpan } from '@/lib/api/analytic';

// QuickRange a relative range of the picker, the resolution of its histogram
export type QuickRange = 'last60s' | 'last60m' | 'last24h';

export const QUICK_RANGES: Record<QuickRange, HistogramResolution> = {
  last60s: HistogramResolution.HISTOGRAM_RESOLUTION_SECOND,
  last60m: HistogramResolution.HISTOGRAM_RESOLUTION_MINUTE,
  last24h: HistogramResolution.HISTOGRAM_RESOLUTION_HOUR,
};

/**
 * TimeRange the range of the page: a quick one, ending now, or an absolute
 * one, from inclusive to exclusive (millis): picked on the calendar, or
 * zoomed into on the histogram, then with the resolution of its histogram.
 */
export type TimeRange =
  { quick: QuickRange } | { from: number; to: number; resolution?: HistogramResolution };

// Window the bounds of a range at a time, to undefined when it ends now
type Window = { from: number; to?: number };

// The second buckets are kept a week (histogram _secondRetention)
const SECOND_RETENTION_MILLIS = 7 * 24 * 60 * 60 * 1000;

/**
 * The resolution of the histogram of a range: that of a quick one, else the
 * one it was zoomed with, else the finest whose buckets of one page cover
 * it; the seconds only within their retention.
 */
export function resolutionOf(range: TimeRange, now: number): HistogramResolution {
  if ('quick' in range) return QUICK_RANGES[range.quick];
  const resolution = range.resolution ?? resolutionOfSpan(range.to - range.from);
  const expired = now - range.from > SECOND_RETENTION_MILLIS;
  return resolution === HistogramResolution.HISTOGRAM_RESOLUTION_SECOND && expired
    ? HistogramResolution.HISTOGRAM_RESOLUTION_MINUTE
    : resolution;
}

/**
 * The range a drag over the histogram zooms into, from inclusive to
 * exclusive: one resolution finer than `current` (hours to minutes to
 * seconds), finer still when the range fits in a page of a finer one.
 */
export function zoomOf(current: HistogramResolution, from: number, to: number): TimeRange {
  const finer = Math.max(RESOLUTIONS.indexOf(current) - 1, 0);
  const fitting = RESOLUTIONS.indexOf(resolutionOfSpan(to - from));
  return { from, to, resolution: RESOLUTIONS[Math.min(finer, fitting)] };
}

/** The bounds of a range: a quick one starts with its first bucket at now */
export function windowOf(range: TimeRange, now: number): Window {
  if ('from' in range) return { from: range.from, to: range.to };
  const { millis, count } = RESOLUTION_BUCKETS[QUICK_RANGES[range.quick]];
  return { from: (Math.floor(now / millis) - count + 1) * millis };
}

/**
 * The buckets the histogram of a range shows when they are only its end,
 * the range longer than its pages, else undefined.
 */
export function clippedOf(range: TimeRange, now: number): number | undefined {
  if ('quick' in range) return undefined;
  const resolution = resolutionOf(range, now);
  const { count } = RESOLUTION_BUCKETS[resolution];
  const buckets = count * maxPagesOf(resolution);
  return range.to - range.from > RESOLUTION_BUCKETS[resolution].millis * buckets
    ? buckets
    : undefined;
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
