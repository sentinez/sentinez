import { HistogramResolution } from '@sentinez/proto/sentinez/types/v1/known';
import type { DateTimeFormatOptions } from 'next-intl';

import {
  HOUR_MILLIS,
  RESOLUTIONS,
  RESOLUTION_BUCKETS,
  durationOf,
  fits,
  resolutionOfSpan,
  spanOf,
} from '@/lib/api/analytic';

// The quick ranges of the picker, ending now, and the resolution of their
// histogram
export const QUICK_RANGES = {
  last5m: HistogramResolution.HISTOGRAM_RESOLUTION_5_MINUTES,
  last15m: HistogramResolution.HISTOGRAM_RESOLUTION_15_MINUTES,
  last30m: HistogramResolution.HISTOGRAM_RESOLUTION_30_MINUTES,
  last1h: HistogramResolution.HISTOGRAM_RESOLUTION_1_HOUR,
  last3h: HistogramResolution.HISTOGRAM_RESOLUTION_3_HOURS,
  last6h: HistogramResolution.HISTOGRAM_RESOLUTION_6_HOURS,
  last12h: HistogramResolution.HISTOGRAM_RESOLUTION_12_HOURS,
  last24h: HistogramResolution.HISTOGRAM_RESOLUTION_24_HOURS,
  last2d: HistogramResolution.HISTOGRAM_RESOLUTION_2_DAYS,
  last7d: HistogramResolution.HISTOGRAM_RESOLUTION_7_DAYS,
  last30d: HistogramResolution.HISTOGRAM_RESOLUTION_30_DAYS,
} as const satisfies Record<string, HistogramResolution>;

// QuickRange a relative range of the picker
export type QuickRange = keyof typeof QUICK_RANGES;

// The quick range the page opens on
export const DEFAULT_QUICK_RANGE: QuickRange = 'last1h';

// Window the bounds of a range at a time, to undefined when it ends now
type Window = { from: number; to?: number };

/**
 * TimeRange the range of the page: a quick one, ending now, or an absolute
 * one, from inclusive to exclusive (millis): picked on the calendar, or
 * zoomed into on the histogram, then with the resolution of its histogram
 * and, when the zoom kept it, the window its buckets were loaded for.
 */
export type TimeRange =
  | { quick: QuickRange }
  | { from: number; to: number; resolution?: HistogramResolution; loaded?: Window };

/**
 * The resolution of the histogram of a range: that of a quick one, else the
 * one it was zoomed with while it still fits it, else the shortest fitting
 * it, as the API picks it (resolutionOfSpan); buckets of seconds only
 * within their retention.
 */
export function resolutionOf(range: TimeRange, now: number): HistogramResolution {
  if ('quick' in range) return QUICK_RANGES[range.quick];
  const span = range.to - range.from;
  const age = now - range.from;
  if (range.resolution && fits(range.resolution, span, age)) return range.resolution;
  return resolutionOfSpan(span, age);
}

/**
 * The range a drag over the histogram zooms into, from inclusive to
 * exclusive: at least one resolution shorter than `current`, shorter still
 * when the range fits a shorter one. At the shortest one the resolution is
 * kept, so are the buckets of `loaded`, the window of the histogram: they
 * hold the narrower range already.
 */
export function zoomOf(
  current: HistogramResolution,
  from: number,
  to: number,
  loaded: Window,
): TimeRange {
  const shorter = Math.max(RESOLUTIONS.indexOf(current) - 1, 0);
  const fitting = RESOLUTIONS.indexOf(resolutionOfSpan(to - from));
  const resolution = RESOLUTIONS[Math.min(shorter, fitting)];
  const kept = resolution === current && loaded.to !== undefined;
  return { from, to, resolution, ...(kept && { loaded }) };
}

/** The bounds of a range: a quick one starts with its first bucket at now */
export function windowOf(range: TimeRange, now: number): Window {
  if ('from' in range) return { from: range.from, to: range.to };
  const { millis, count } = RESOLUTION_BUCKETS[QUICK_RANGES[range.quick]];
  return { from: (Math.floor(now / millis) - count + 1) * millis };
}

export const SECONDS_TICK: DateTimeFormatOptions = {
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
};
const MINUTES_TICK: DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit' };
const DAYS_TICK: DateTimeFormatOptions = {
  day: '2-digit',
  month: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
};

// The tick of a bucket start: its seconds for buckets of seconds, its day
// for a histogram over a day or more, else its time
export function tickFormatOf(bucketMillis: number, resolution: HistogramResolution) {
  if (durationOf(bucketMillis).unit === 'second') return SECONDS_TICK;
  return spanOf(resolution) > 12 * HOUR_MILLIS ? DAYS_TICK : MINUTES_TICK;
}

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
