import { Activity } from '@sentinez/proto/sentinez/apps/analytic/v1/model';
import {
  ListActivitiesRequest,
  ListActivitiesResponse,
} from '@sentinez/proto/sentinez/apps/analytic/v1/analytic';
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { type Pages, paginate, toQuery } from '@/lib/api/pages';
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
  metadata: { createdAt: new Date(createdAt), updatedAt: new Date(createdAt) },
  id: `senz.analytic.activity.sample-${n}`,
  resourceId: `senz.analytic.resources.${resource}`,
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
      (!req.resourceIds?.length || req.resourceIds.includes(a.resourceId)),
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
