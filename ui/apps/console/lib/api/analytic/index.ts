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
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { type Pages, paginate, toQuery } from '@/lib/api/pages';
import { SAMPLE_TENANT_ID } from '@/lib/api/tenant';
// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

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

function withFallback<T>(label: string, fallback: () => T) {
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
  return withFallback<ListActivitiesResponse>('listActivities', () => {
    const list = filterSample(req);
    return { activities: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/analytic/activities`, {
      params: toQuery(ListActivitiesRequest.toJSON(ListActivitiesRequest.fromPartial(req))),
      // repeat `ids` / `resourceIds` keys (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListActivitiesResponse.fromJSON(resp.data ?? {});
  });
}

const HOUR_MILLIS = 60 * 60 * 1000;

// The 24 sample hourly buckets, as the API returns them: those ending with
// the hour of `to`, else the current one
function sampleHours(to?: Date): Date[] {
  const end = Math.floor((to ?? new Date()).getTime() / HOUR_MILLIS) + 1;
  return Array.from({ length: 24 }, (_, i) => new Date((end - 24 + i) * HOUR_MILLIS));
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
  return withFallback<GetUniqueVisitorHistogramResponse>('getUniqueVisitorHistogram', () => ({
    points: sampleVisitorPoints(full),
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/analytic/visitors/histogram`, {
      params: toQuery(GetUniqueVisitorHistogramRequest.toJSON(full)),
      // repeat `resourceIds` keys (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return GetUniqueVisitorHistogramResponse.fromJSON(resp.data ?? {});
  });
}

// Sample client ips: the first ones visit every hour, the others now and then
const SAMPLE_IPS = ['203.0.113.7', '198.51.100.23', '192.0.2.41', '203.0.113.88', '198.51.100.5'];

// Hourly sample visitors of each resource (sampleHours): the most requests
// first in each bucket
function sampleVisitors(req: ListVisitorsRequest): VisitorPoint[] {
  return sampleHours(req.to).flatMap((time, i) =>
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
  return withFallback<ListVisitorsResponse>('listVisitors', () => ({
    visitors: sampleVisitors(full),
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/analytic/visitors`, {
      params: toQuery(ListVisitorsRequest.toJSON(full)),
      // repeat `resourceIds` keys (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListVisitorsResponse.fromJSON(resp.data ?? {});
  });
}
