import type {
  FieldTopValues,
  HistogramPoint,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { HistogramResolution } from '@sentinez/proto/sentinez/types/v1/known';
import { RESOLUTION_BUCKETS } from '@/lib/api/analytic/common';

// The sample buckets of a resolution: those ending with the one of `to`,
// else the current one
export function sampleBuckets(resolution: HistogramResolution, to?: Date): Date[] {
  const { millis, count } = RESOLUTION_BUCKETS[resolution];
  const end = Math.ceil((to ?? new Date(Date.now() + 1)).getTime() / millis);
  return Array.from({ length: count }, (_, i) => new Date((end - count + i) * millis));
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
