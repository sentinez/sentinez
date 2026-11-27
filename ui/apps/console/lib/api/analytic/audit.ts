import {
  GetAuditLogHistogramRequest,
  GetAuditLogHistogramResponse,
  GetAuditLogTopValuesRequest,
  GetAuditLogTopValuesResponse,
  ListAuditLogsRequest,
  ListAuditLogsResponse,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { paginate } from '@/lib/api/pages';
import { filterSample, sampleHistogram, sampleTopValues } from '@/lib/faker/analytic/audit';
import { type ApiOptions, getJSON, intervalOf } from './common';

// GET /analytic/audit-logs
export async function listAuditLogs(
  req: Partial<ListAuditLogsRequest> = {},
  options?: ApiOptions,
): Promise<ListAuditLogsResponse> {
  const full = ListAuditLogsRequest.fromPartial(req);
  return getJSON(
    '/analytic/audit-logs',
    full,
    { req: ListAuditLogsRequest, resp: ListAuditLogsResponse },
    () => ({ logs: paginate(filterSample(full), full.page) }),
    options,
  );
}

// GET /analytic/audit-logs/histogram
export async function getAuditLogHistogram(
  req: Partial<GetAuditLogHistogramRequest> = {},
  options?: ApiOptions,
): Promise<GetAuditLogHistogramResponse> {
  const full = GetAuditLogHistogramRequest.fromPartial(req);
  return getJSON(
    '/analytic/audit-logs/histogram',
    full,
    { req: GetAuditLogHistogramRequest, resp: GetAuditLogHistogramResponse },
    () => ({
      points: sampleHistogram(full),
      interval: intervalOf(full.resolution),
      truncated: false,
    }),
    options,
  );
}

// GET /analytic/audit-logs/top-values
export async function getAuditLogTopValues(
  req: Partial<GetAuditLogTopValuesRequest> = {},
  options?: ApiOptions,
): Promise<GetAuditLogTopValuesResponse> {
  const full = GetAuditLogTopValuesRequest.fromPartial(req);
  return getJSON(
    '/analytic/audit-logs/top-values',
    full,
    { req: GetAuditLogTopValuesRequest, resp: GetAuditLogTopValuesResponse },
    () => sampleTopValues(full),
    options,
  );
}
