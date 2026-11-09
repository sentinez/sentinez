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
  SAMPLE_IPS,
  VISITORS_RESOLUTION,
  getJSON,
  intervalOf,
  sampleBuckets,
} from './common';

export {
  EVENT_RETENTION_MILLIS,
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

// The unique visitor histogram of the sample visitors, so both agree: the
// distinct ips per bucket and resource
function sampleVisitorPoints(req: GetUniqueVisitorHistogramRequest): HistogramPoint[] {
  const counts = new Map<string, HistogramPoint>();
  for (const v of sampleVisitors(req)) {
    const key = `${v.time?.getTime()}|${v.resourceId}`;
    const point = counts.get(key) ?? { time: v.time, resourceId: v.resourceId, key: 0, total: 0 };
    point.total++;
    counts.set(key, point);
  }
  return [...counts.values()];
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
    }),
    options,
  );
}

// The resolution of the samples of a request: the one asked, else that of
// the visitors
const sampleResolution = (req: { resolution: HistogramResolution }) =>
  req.resolution || VISITORS_RESOLUTION;

// Sample visitors of each resource per bucket of the resolution
// (sampleBuckets): the most requests first in each bucket
function sampleVisitors(req: ListVisitorsRequest): VisitorPoint[] {
  return sampleBuckets(sampleResolution(req), req.to).flatMap((time, i) =>
    req.resourceIds.flatMap((resourceId) =>
      SAMPLE_IPS.slice(0, 2 + (i % 4)).map((clientIp, n) => ({
        time,
        resourceId,
        clientIp,
        total: Math.round(60 / (n + 1) + 20 * Math.sin((i + n) / 3)) + 25,
      })),
    ),
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
    }),
    options,
  );
}
