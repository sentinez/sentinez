import {
  GetAuditLogHistogramRequest,
  GetAuditLogHistogramResponse,
  GetAuditLogTopValuesRequest,
  GetAuditLogTopValuesResponse,
  type HistogramPoint,
  ListAuditLogsRequest,
  ListAuditLogsResponse,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import type { AuditLog, FieldChange } from '@sentinez/proto/sentinez/types/audit/v1/audit';
import { ControlPlane, HistogramResolution } from '@sentinez/proto/sentinez/types/v1/known';
import { paginate } from '@/lib/api/pages';
import { SAMPLE_TENANT_ID } from '@/lib/api/tenant';
import {
  type ApiOptions,
  type SampleRecord,
  HOUR_MILLIS,
  RESOLUTION_BUCKETS,
  SAMPLE_IPS,
  countPoints,
  countTopValues,
  getJSON,
  matchesFilters,
  intervalOf,
  ofRequests,
  sampleBuckets,
} from './common';

// ---- Sample audit logs, used as fallback when the API call fails ----

const SAMPLE_AGENT =
  'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0 Safari/537.36';

// A call of the samples: the method, its request, the resource it is
// about, the fields it changed, and the grpc code and error it ended with
type SampleCall = {
  method: string;
  request: Record<string, unknown>;
  resourceId?: string;
  change?: FieldChange[];
  // A call about no tenant, such as a login
  tenantless?: boolean;
  code?: number;
  error?: string;
};

// The OWASP CRS configuration of the samples, before and after an update
const SAMPLE_CRS_BEFORE = [
  'SecRuleEngine On',
  'SecRequestBodyAccess On',
  'SecAction "id:900000,phase:1,pass,nolog,setvar:tx.blocking_paranoia_level=1"',
  'SecAction "id:900110,phase:1,pass,nolog,setvar:tx.inbound_anomaly_score_threshold=5"',
  'SecAction "id:900200,phase:1,pass,nolog,setvar:\'tx.allowed_methods=GET HEAD POST OPTIONS\'"',
  'SecDefaultAction "phase:1,log,auditlog,pass"',
  'SecDefaultAction "phase:2,log,auditlog,pass"',
].join('\n');

const SAMPLE_CRS_AFTER = SAMPLE_CRS_BEFORE.replace(
  'blocking_paranoia_level=1',
  'blocking_paranoia_level=2',
).replace('GET HEAD POST OPTIONS', 'GET HEAD POST PUT OPTIONS');

// The calls of the samples in tenantId
const sampleCalls = (tenantId: string): SampleCall[] => [
  {
    method: '/sentinez.apps.security.v1.SecurityService/CreateSecRule',
    request: { secRule: { name: 'block-admin', owner: { resourceId: 'rsrc_sample' } } },
    resourceId: 'rsrc_sample',
  },
  {
    method: '/sentinez.apps.security.v1.SecurityService/UpdateRateLimit',
    request: { id: 'srule_sample-api-limit', updateMask: 'timeWindow' },
    resourceId: 'rsrc_sample',
    change: [{ field: 'time_window', before: '60s', after: '30s' }],
  },
  {
    method: '/sentinez.apps.tenant.v1.TenantService/CreateMember',
    request: { member: { tenantId, userId: 'user_bob', roles: ['ROLE_LEADER'] } },
  },
  {
    method: '/sentinez.apps.security.v1.SecurityService/DeleteSecRule',
    request: { id: 'srule_missing' },
    resourceId: 'rsrc_sample',
    code: 5,
    error: 'secrule not found: id=srule_missing',
  },
  {
    method: '/sentinez.apps.tenant.v1.TenantService/UpdateResource',
    request: { id: 'rsrc_sample', resource: { name: 'shop' }, updateMask: 'name' },
    resourceId: 'rsrc_sample',
    change: [{ field: 'name', before: 'store', after: 'shop' }],
  },
  {
    method: '/sentinez.apps.security.v1.SecurityService/UpdateCoreRuleset',
    request: {
      id: 'crs_sample',
      coreRuleset: { content: SAMPLE_CRS_AFTER },
      updateMask: 'content',
    },
    resourceId: 'rsrc_sample',
    change: [{ field: 'content', before: SAMPLE_CRS_BEFORE, after: SAMPLE_CRS_AFTER }],
  },
  {
    method: '/sentinez.apps.iam.v1.IdentityAccessManagementService/Login',
    request: { emailOrUsername: 'mallory', password: 'REDACTED' },
    tenantless: true,
    code: 16,
    error: 'invalid email or password',
  },
  {
    method: '/sentinez.apps.tenant.v1.TenantService/UpdateTenant',
    request: { id: tenantId, tenant: { plan: 'PLAN_PRO' } },
    code: 7,
    error: 'role ROLE_SUPER_ADMIN is required in tenant ' + tenantId,
  },
];

const SAMPLE_USERS = [
  { userId: 'user_alice', name: 'Alice' },
  { userId: 'user_bob', name: 'Bob' },
];

// The samples of tenantId, the latest first, one every 7 minutes over the
// 24 hours before end
function sampleAuditLogs(end: number, tenantId: string): AuditLog[] {
  const calls = sampleCalls(tenantId);
  const step = 7 * 60 * 1000;
  const last = Math.floor(end / step) * step;
  return Array.from({ length: Math.floor((24 * HOUR_MILLIS) / step) }, (_, i) => {
    const seed = Math.floor(last / step) - i;
    const call = calls[seed % calls.length]!;
    const user = SAMPLE_USERS[seed % SAMPLE_USERS.length]!;
    return {
      id: `aud_SAMPLE${String(seed).padStart(12, '0')}`,
      method: call.method,
      userId: user.userId,
      name: user.name,
      controlPlane: ControlPlane.CONTROL_PLANE_PORTAL,
      remoteAddr: `${SAMPLE_IPS[seed % SAMPLE_IPS.length]}:${40000 + (seed % 20000)}`,
      userAgent: SAMPLE_AGENT,
      request: call.request,
      code: call.code ?? 0,
      error: call.error ?? '',
      latency: { seconds: 0, nanos: ((seed % 40) + 3) * 1_000_000 },
      tenantId: call.tenantless ? '' : tenantId,
      timestamp: last - i * step,
      resourceId: call.resourceId ?? '',
      change: call.change && { fields: call.change },
    };
  });
}

// The record of a log as the API filters and counts it, by grpc code
const recordOf = (l: AuditLog): SampleRecord => ({
  time: l.timestamp,
  resourceId: l.resourceId,
  key: l.code,
  fields: {
    method: l.method,
    user_id: l.userId,
    code: String(l.code),
    control_plane: String(l.controlPlane),
    tenant_id: l.tenantId,
    remote_addr: l.remoteAddr,
    user_agent: l.userAgent,
  },
});

// AuditSelection what every request of the audit logs selects
type AuditSelection = Pick<
  ListAuditLogsRequest,
  'tenantIds' | 'withoutTenant' | 'resourceIds' | 'requestIds' | 'filters' | 'from' | 'to'
>;

// The samples a request selects, as the API does; they are of its first
// tenant, the sample tenant without, so that any tenant has some
function filterSample(req: AuditSelection): AuditLog[] {
  const from = req.from?.getTime() ?? -Infinity;
  const to = req.to?.getTime() ?? Infinity;
  const tenantId = req.tenantIds[0] ?? SAMPLE_TENANT_ID;
  return sampleAuditLogs(req.to?.getTime() ?? Date.now(), tenantId)
    .filter(ofRequests(req.requestIds, (l) => l.id))
    .filter(
      (l) =>
        l.timestamp >= from &&
        l.timestamp < to &&
        (req.withoutTenant
          ? l.tenantId === ''
          : !req.tenantIds.length || req.tenantIds.includes(l.tenantId)) &&
        (!req.resourceIds.length || req.resourceIds.includes(l.resourceId)) &&
        matchesFilters(recordOf(l), req.filters),
    );
}

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

// The histogram of the samples: per bucket, resource and grpc code
function sampleHistogram(req: GetAuditLogHistogramRequest): HistogramPoint[] {
  const resolution = req.resolution || HistogramResolution.HISTOGRAM_RESOLUTION_1_HOUR;
  const { millis } = RESOLUTION_BUCKETS[resolution];
  const buckets = sampleBuckets(resolution, req.to);
  const from = buckets[0]?.getTime() ?? 0;
  const to = from + buckets.length * millis;
  const logs = filterSample({ ...req, from: new Date(from), to: new Date(to) });
  return countPoints(new Map(), logs.map(recordOf), millis);
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

// The top values of the samples, as the API counts them
function sampleTopValues(req: GetAuditLogTopValuesRequest): GetAuditLogTopValuesResponse {
  const to = req.to ?? new Date();
  const from = req.from ?? new Date(to.getTime() - HOUR_MILLIS);
  return countTopValues(filterSample({ ...req, from, to }).map(recordOf), req.fields, req.limit);
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
