import { Activity } from '@sentinez/proto/sentinez/apps/analytic/v1/model';
import {
  ListActivitiesRequest,
  ListActivitiesResponse,
} from '@sentinez/proto/sentinez/apps/analytic/v1/analytic';
import {
  GetUniqueVisitorHistogramRequest,
  GetUniqueVisitorHistogramResponse,
  type HistogramPoint,
  ListVisitorsRequest,
  ListVisitorsResponse,
  type VisitorPoint,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { HistogramResolution } from '@sentinez/proto/sentinez/types/v1/known';
import { type Pages, paginate } from '@/lib/api/pages';
import { SAMPLE_TENANT_ID } from '@/lib/api/tenant';
import {
  type ApiOptions,
  type SampleLevel,
  SAMPLE_IPS,
  VISITORS_RESOLUTION,
  getJSON,
  intervalOf,
  sampleBuckets,
  sampleLevel,
  sampleRandom,
} from './common';

export {
  EVENT_RETENTION_MILLIS,
  DAY_MILLIS,
  HOUR_MILLIS,
  RESOLUTIONS,
  RESOLUTION_BUCKETS,
  SECOND_MILLIS,
  VISITORS_RESOLUTION,
  bucketMillisOf,
  durationOf,
  fits,
  resolutionOfSpan,
  spanOf,
  type ApiOptions,
} from './common';
export * from './audit';
export * from './events';
export type { Pages };

// ---- Sample data, used as fallback when the API call fails ----

// Deterministic visitor series (no Math.random, safe for SSR)
const visitorSeries = (seed: number, length = 24): number[] =>
  Array.from({ length }, (_, i) => Math.round(50 + seed * 10 + 40 * Math.sin((i + seed) / 3)));

const sampleActivity = (n: number, resource: string, createdAt: string): Activity => ({
  owner: {
    createdAt: new Date(createdAt),
    updatedAt: new Date(createdAt),
    resourceId: `rsrc_${resource}`,
    tenantId: SAMPLE_TENANT_ID,
  },
  id: `actv_sample-${n}`,
  uniqueVisitor: visitorSeries(n),
});

export const SAMPLE_ACTIVITIES: Activity[] = [
  sampleActivity(1, 'badcheese', '2026-10-01T00:00:00Z'),
  sampleActivity(2, 'badcheese', '2026-10-02T00:00:00Z'),
  sampleActivity(3, 'shop', '2026-10-01T00:00:00Z'),
  sampleActivity(4, 'shop', '2026-10-02T00:00:00Z'),
  // A resource with no traffic yet
  { ...sampleActivity(5, 'staging', '2026-10-02T00:00:00Z'), uniqueVisitor: [] },
  // Extra rows so the list has several pages to page through
  ...Array.from({ length: 15 }, (_, i) =>
    sampleActivity(6 + i, `site${(i % 5) + 1}`, `2026-09-${String(10 + i)}T00:00:00Z`),
  ),
];

function filterSample(req: Partial<ListActivitiesRequest>): Activity[] {
  return SAMPLE_ACTIVITIES.filter(
    (a) =>
      (!req.ids?.length || req.ids.includes(a.id)) &&
      (!req.resourceIds?.length || req.resourceIds.includes(a.owner?.resourceId ?? '')),
  );
}

// GET /analytic/activities
export async function listActivities(
  req: Partial<ListActivitiesRequest> = {},
  options?: ApiOptions,
): Promise<ListActivitiesResponse> {
  return getJSON(
    '/analytic/activities',
    ListActivitiesRequest.fromPartial(req),
    { req: ListActivitiesRequest, resp: ListActivitiesResponse },
    () => {
      const list = filterSample(req);
      return { activities: paginate(list, req.page), total: list.length };
    },
    options,
  );
}

// The unique visitor histogram of the samples: per bucket and resource,
// all the visitors of sampleVisitorCount (not only the listed ones)
function sampleVisitorPoints(req: GetUniqueVisitorHistogramRequest): HistogramPoint[] {
  return sampleBuckets(sampleResolution(req), req.to).flatMap((time) =>
    req.resourceIds.map((resourceId) => ({
      time,
      resourceId,
      key: 0,
      total: sampleVisitorCount(resourceId, time),
    })),
  );
}

// GET /analytic/visitors/histogram
export async function getUniqueVisitorHistogram(
  req: Partial<GetUniqueVisitorHistogramRequest> = {},
  options?: ApiOptions,
): Promise<GetUniqueVisitorHistogramResponse> {
  const full = GetUniqueVisitorHistogramRequest.fromPartial(req);
  return getJSON(
    '/analytic/visitors/histogram',
    full,
    { req: GetUniqueVisitorHistogramRequest, resp: GetUniqueVisitorHistogramResponse },
    () => ({
      points: sampleVisitorPoints(full),
      interval: intervalOf(sampleResolution(full)),
      truncated: false,
    }),
    options,
  );
}

// The resolution of the samples of a request: the one asked, else that of
// the visitors
const sampleResolution = (req: { resolution: HistogramResolution }) =>
  req.resolution || VISITORS_RESOLUTION;

// The unique visitors of a busy sample bucket, large to test the layout
// (wide numbers in the charts, the axes and the tables)
const SAMPLE_VISITOR_SCALE = 6000;
// The most visitors a sample bucket lists (sampleVisitors); the histogram
// counts them all
const SAMPLE_LISTED_VISITORS = 150;

// The sample ip of a visitor: the known SAMPLE_IPS first, then made up ones
const sampleIp = (n: number) =>
  SAMPLE_IPS[n] ?? `198.51.${100 + Math.floor(n / 250)}.${(n % 250) + 1}`;

// The traffic of the sample visitors (sampleLevel)
const SAMPLE_VISITOR_LEVEL: SampleLevel = {
  quiet: 0.3,
  mean: 0.55,
  day: 0.35,
  noise: 0.3,
  burst: 2,
};

// The unique visitors of a resource in the bucket at time (sampleLevel)
function sampleVisitorCount(resourceId: string, time: Date): number {
  return Math.round(sampleLevel(resourceId, time, SAMPLE_VISITOR_LEVEL) * SAMPLE_VISITOR_SCALE);
}

// The listed sample visitors of a resource in the bucket at time (at most
// SAMPLE_LISTED_VISITORS of sampleVisitorCount), the most requests first:
// the SAMPLE_IPS every bucket, the others of a pool shifting per bucket, so
// the day has thousands of distinct ips
function sampleBucketVisitors(resourceId: string, time: Date): VisitorPoint[] {
  const count = Math.min(SAMPLE_LISTED_VISITORS, sampleVisitorCount(resourceId, time));
  const random = sampleRandom(`${resourceId}|${time.getTime()}|list`);
  const shift = (Math.floor(time.getTime() / 60_000) * 53) % 5000;
  return Array.from({ length: count }, (_, n) => ({
    time,
    resourceId,
    clientIp: sampleIp(n < SAMPLE_IPS.length ? n : n + shift),
    total: Math.round((20000 / (n + 1)) * (0.5 + random())) + Math.floor(random() * 40) + 1,
  }));
}

// Sample visitors of each resource per bucket of the resolution
// (sampleBuckets)
function sampleVisitors(req: ListVisitorsRequest): VisitorPoint[] {
  return sampleBuckets(sampleResolution(req), req.to).flatMap((time) =>
    req.resourceIds.flatMap((resourceId) => sampleBucketVisitors(resourceId, time)),
  );
}

// GET /analytic/visitors
export async function listVisitors(
  req: Partial<ListVisitorsRequest> = {},
  options?: ApiOptions,
): Promise<ListVisitorsResponse> {
  const full = ListVisitorsRequest.fromPartial(req);
  return getJSON(
    '/analytic/visitors',
    full,
    { req: ListVisitorsRequest, resp: ListVisitorsResponse },
    () => ({
      visitors: sampleVisitors(full),
      interval: intervalOf(sampleResolution(full)),
      truncated: false,
    }),
    options,
  );
}
