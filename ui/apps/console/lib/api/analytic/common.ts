import type {
  FieldTopValues,
  HistogramPoint,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { HistogramInterval, HistogramResolution } from '@sentinez/proto/sentinez/types/v1/known';
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { toQuery } from '@/lib/api/pages';

// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

export function withFallback<T>(label: string, fallback: () => T) {
  return async (call: () => Promise<T>): Promise<T> => {
    try {
      return await call();
    } catch (err: any) {
      if (!USE_SAMPLE || axios.isCancel(err)) throw err;
      console.warn(`[analytic] ${label} failed, using sample response:`, err?.message);
      return fallback();
    }
  };
}

// The codecs of a request and of its response, generated from the proto
type Codecs<Req, Resp> = {
  req: { toJSON(message: Req): unknown };
  resp: { fromJSON(object: unknown): Resp };
};

// GET path with the query params of req, else the sample response; the
// repeated fields repeat their key (collectionFormat: multi)
export async function getJSON<Req, Resp>(
  path: string,
  req: Req,
  codecs: Codecs<Req, Resp>,
  sample: () => Resp,
  options?: ApiOptions,
): Promise<Resp> {
  return withFallback<Resp>(
    path,
    sample,
  )(async () => {
    const resp = await axios.get(`${API_BASE_PATH}${path}`, {
      params: toQuery(codecs.req.toJSON(req)),
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return codecs.resp.fromJSON(resp.data ?? {});
  });
}

export const SECOND_MILLIS = 1000;
const MINUTE_MILLIS = 60 * SECOND_MILLIS;
export const HOUR_MILLIS = 60 * MINUTE_MILLIS;
export const DAY_MILLIS = 24 * HOUR_MILLIS;

// BucketUnit the unit a duration is told in
type BucketUnit = 'second' | 'minute' | 'hour' | 'day';

// The width of the buckets of each interval (histogram _bucket)
const INTERVAL_MILLIS: Record<HistogramInterval, number> = {
  [HistogramInterval.HISTOGRAM_INTERVAL_UNSPECIFIED]: MINUTE_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_10_SECONDS]: 10 * SECOND_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_30_SECONDS]: 30 * SECOND_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_1_MINUTE]: MINUTE_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_5_MINUTES]: 5 * MINUTE_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_10_MINUTES]: 10 * MINUTE_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_15_MINUTES]: 15 * MINUTE_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_30_MINUTES]: 30 * MINUTE_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_1_HOUR]: HOUR_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_3_HOURS]: 3 * HOUR_MILLIS,
  [HistogramInterval.HISTOGRAM_INTERVAL_12_HOURS]: 12 * HOUR_MILLIS,
  [HistogramInterval.UNRECOGNIZED]: MINUTE_MILLIS,
};

const UNIT_MILLIS: [BucketUnit, number][] = [
  ['day', DAY_MILLIS],
  ['hour', HOUR_MILLIS],
  ['minute', MINUTE_MILLIS],
  ['second', SECOND_MILLIS],
];

// A duration (millis) in the largest unit it is a whole number of, e.g.
// 30 days, 12 hours
export function durationOf(millis: number): { count: number; unit: BucketUnit } {
  const [unit, size] = UNIT_MILLIS.find(([, size]) => millis % size === 0) ?? ['second', 1000];
  return { count: millis / size, unit };
}

// The range each resolution charts and the interval of its buckets, as
// the API has them (histogram _span, _resolutions); unspecified: the API
// picks one hour without a range
const RESOLUTION_SPECS: Record<HistogramResolution, { span: number; interval: HistogramInterval }> =
  {
    [HistogramResolution.HISTOGRAM_RESOLUTION_UNSPECIFIED]: {
      span: HOUR_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_1_MINUTE,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_5_MINUTES]: {
      span: 5 * MINUTE_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_10_SECONDS,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_15_MINUTES]: {
      span: 15 * MINUTE_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_30_SECONDS,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_30_MINUTES]: {
      span: 30 * MINUTE_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_1_MINUTE,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_1_HOUR]: {
      span: HOUR_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_1_MINUTE,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_3_HOURS]: {
      span: 3 * HOUR_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_5_MINUTES,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_6_HOURS]: {
      span: 6 * HOUR_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_10_MINUTES,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_12_HOURS]: {
      span: 12 * HOUR_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_15_MINUTES,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_24_HOURS]: {
      span: DAY_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_30_MINUTES,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_2_DAYS]: {
      span: 2 * DAY_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_1_HOUR,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_7_DAYS]: {
      span: 7 * DAY_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_3_HOURS,
    },
    [HistogramResolution.HISTOGRAM_RESOLUTION_30_DAYS]: {
      span: 30 * DAY_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_12_HOURS,
    },
    [HistogramResolution.UNRECOGNIZED]: {
      span: HOUR_MILLIS,
      interval: HistogramInterval.HISTOGRAM_INTERVAL_1_MINUTE,
    },
  };

// The width and number of the buckets of each resolution, as the API
// returns them: its span over the width of its interval
export const RESOLUTION_BUCKETS = Object.fromEntries(
  Object.entries(RESOLUTION_SPECS).map(([resolution, { span, interval }]) => {
    const millis = INTERVAL_MILLIS[interval];
    return [resolution, { millis, count: span / millis }];
  }),
) as Record<HistogramResolution, { millis: number; count: number }>;

// The time the buckets of a resolution cover
export const spanOf = (resolution: HistogramResolution) => RESOLUTION_SPECS[resolution].span;

// The interval of the buckets of a resolution
export const intervalOf = (resolution: HistogramResolution) =>
  RESOLUTION_SPECS[resolution].interval;

// The resolution of the visitors of a resource: the last 24 hours
export const VISITORS_RESOLUTION = HistogramResolution.HISTOGRAM_RESOLUTION_24_HOURS;

// The resolutions of a histogram, the shortest first (histogram
// _ascending): those of RESOLUTION_SPECS, but unspecified and unrecognized
export const RESOLUTIONS = (Object.keys(RESOLUTION_SPECS).map(Number) as HistogramResolution[])
  .filter((r) => r > HistogramResolution.HISTOGRAM_RESOLUTION_UNSPECIFIED)
  .sort((a, b) => spanOf(a) - spanOf(b));

// The longest resolution, that of a range longer than every one
const LONGEST_RESOLUTION = RESOLUTIONS[RESOLUTIONS.length - 1]!;

// How long the events, and their histograms, are kept: the API refuses a
// range starting before (schema.EventRetention)
export const EVENT_RETENTION_MILLIS = 30 * DAY_MILLIS;

// The buckets of seconds are kept a week (histogram _secondRetention)
const SECOND_RETENTION_MILLIS = 7 * DAY_MILLIS;

/**
 * Whether a resolution charts a range of span (millis) starting age ago,
 * as the API decides it (histogram resolutionSpec.fits): its buckets are
 * still kept then, and its span holds the range but one bucket.
 */
export function fits(resolution: HistogramResolution, span: number, age = 0): boolean {
  const { millis } = RESOLUTION_BUCKETS[resolution];
  if (millis < MINUTE_MILLIS && age > SECOND_RETENTION_MILLIS) return false;
  return spanOf(resolution) + millis >= span;
}

// The shortest resolution fitting a range of span (millis) starting age
// ago, the longest beyond them all
export function resolutionOfSpan(span: number, age = 0): HistogramResolution {
  return RESOLUTIONS.find((r) => fits(r, span, age)) ?? LONGEST_RESOLUTION;
}

// The width of the buckets of a histogram: the interval the API answered
// with, else the one of its resolution
export function bucketMillisOf(
  resolution: HistogramResolution,
  interval?: HistogramInterval,
): number {
  return interval ? INTERVAL_MILLIS[interval] : RESOLUTION_BUCKETS[resolution].millis;
}

// The sample buckets of a resolution: those ending with the one of `to`,
// else the current one
export function sampleBuckets(resolution: HistogramResolution, to?: Date): Date[] {
  const { millis, count } = RESOLUTION_BUCKETS[resolution];
  const end = Math.ceil((to ?? new Date(Date.now() + 1)).getTime() / millis);
  return Array.from({ length: count }, (_, i) => new Date((end - count + i) * millis));
}

// A random number generator of [0, 1) seeded by seed (mulberry32 over an
// FNV-1a hash): the same seed gives the same numbers, so samples asked twice
// (the histogram, then the visitors) agree and do not change on re-render
export function sampleRandom(seed: string): () => number {
  let h = 0x811c9dc5;
  for (let i = 0; i < seed.length; i++) {
    h = Math.imul(h ^ seed.charCodeAt(i), 0x01000193);
  }
  return () => {
    h = (h + 0x6d2b79f5) | 0;
    let t = Math.imul(h ^ (h >>> 15), 1 | h);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

// The daily rhythm of the samples at time (millis): -1 at 00:00 UTC, 1 at
// 12:00
export const sampleDayCurve = (time: number) =>
  Math.sin(((new Date(time).getUTCHours() - 6) / 24) * 2 * Math.PI);

// A random element of list, drawn with random
export const sampleOneOf = <T>(random: () => number, list: readonly T[]): T =>
  list[Math.floor(random() * list.length)]!;

// SampleLevel the shape of the traffic of a sample resource: its share of a
// quiet resource (the rest scaled by the resource), the mean, the swing of
// the day and the noise around it, and the factor of a burst (1 bucket in 12)
export type SampleLevel = {
  quiet: number;
  mean: number;
  day: number;
  noise: number;
  burst: number;
};

// The traffic level of a resource in the bucket at time, random but the same
// for the same resource and bucket: a busy or quiet resource (its own
// scale), busier by day, with noise and a burst now and then
export function sampleLevel(resourceId: string, time: Date, l: SampleLevel): number {
  const scale = sampleRandom(resourceId)();
  const random = sampleRandom(`${resourceId}|${time.getTime()}`);
  const burst = random() < 0.08 ? l.burst : 1;
  const day = sampleDayCurve(time.getTime());
  return (l.quiet + (1 - l.quiet) * scale) * (l.mean + l.day * day + l.noise * random()) * burst;
}

// Sample client ips: the first ones visit every hour, the others now and then
export const SAMPLE_IPS = [
  '203.0.113.7',
  '198.51.100.23',
  '192.0.2.41',
  '203.0.113.88',
  '198.51.100.5',
];

// ---- Sample records, counted the way the API counts them ----

// SampleRecord a sample event as the histogram and the top values count it
export type SampleRecord = {
  time: number;
  resourceId: string;
  key: number;
  fields: Record<string, string | string[]>;
};

// Whether a sample event is of one of the requests, as the API filters
// request_ids: every one when empty
export const ofRequests =
  <T>(requestIds: string[], requestOf: (e: T) => string) =>
  (e: T) =>
    requestIds.length === 0 || requestIds.includes(requestOf(e));

// Whether a sample record holds the values of the field filters, as the
// API filters them: an element of a repeated field
export const matchesFilters = (r: SampleRecord, filters: string[]) =>
  filters.every((f) => {
    const at = f.indexOf(':');
    const want = f.slice(at + 1);
    const value = r.fields[f.slice(0, at)];
    return Array.isArray(value) ? value.includes(want) : value === want;
  });

// Adds the records to the points of the histogram of buckets of millis,
// one per bucket, resource and key, and returns the points
export function countPoints(
  points: Map<string, HistogramPoint>,
  records: SampleRecord[],
  millis: number,
): HistogramPoint[] {
  for (const r of records) {
    const start = Math.floor(r.time / millis) * millis;
    const id = `${start}|${r.resourceId}|${r.key}`;
    const point = points.get(id) ?? {
      time: new Date(start),
      resourceId: r.resourceId,
      key: r.key,
      total: 0,
    };
    point.total++;
    points.set(id, point);
  }
  return [...points.values()];
}

// The top values of the fields of the records (all of them when none),
// the empty values left out, a record once per element, at most limit
// (5 by default, up to 20) per field
export function countTopValues(
  records: SampleRecord[],
  fields: string[],
  limit: number,
): { total: number; fields: FieldTopValues[] } {
  const max = Math.min(limit > 0 ? limit : 5, 20);
  const names = fields.length ? fields : Object.keys(records[0]?.fields ?? {});
  return {
    total: records.length,
    fields: names.map((field) => {
      const counts = new Map<string, number>();
      for (const r of records) {
        const raw = r.fields[field] ?? '';
        for (const value of new Set(Array.isArray(raw) ? raw : [raw])) {
          if (value) counts.set(value, (counts.get(value) ?? 0) + 1);
        }
      }
      const values = [...counts.entries()]
        .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
        .slice(0, max)
        .map(([value, total]) => ({ value, total, share: total / records.length }));
      return { field, values };
    }),
  };
}
