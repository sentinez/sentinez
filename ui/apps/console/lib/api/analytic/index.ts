import {
  ListActivitiesRequest,
  ListActivitiesResponse,
} from '@sentinez/proto/sentinez/apps/analytic/v1/analytic';
import {
  GetUniqueVisitorHistogramRequest,
  GetUniqueVisitorHistogramResponse,
  ListVisitorsRequest,
  ListVisitorsResponse,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { type Pages, paginate } from '@/lib/api/pages';
import {
  sampleActivities,
  sampleResolution,
  sampleVisitorPoints,
  sampleVisitors,
} from '@/lib/faker/analytic/activity';
import { type ApiOptions, getJSON, intervalOf } from './common';

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
      const list = sampleActivities(req);
      return { activities: paginate(list, req.page), total: list.length };
    },
    options,
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
