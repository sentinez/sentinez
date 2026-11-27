import {
  GetAuditLogHistogramRequest,
  GetAuditLogTopValuesRequest,
  GetAuditLogTopValuesResponse,
  type HistogramPoint,
  ListAuditLogsRequest,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import type { AuditLog, FieldChange } from '@sentinez/proto/sentinez/types/audit/v1/audit';
import { ControlPlane, HistogramResolution } from '@sentinez/proto/sentinez/types/v1/known';
import { SAMPLE_TENANT_ID } from '@/lib/faker/tenant';
import { HOUR_MILLIS, RESOLUTION_BUCKETS } from '@/lib/api/analytic/common';
import {
  countPoints,
  countTopValues,
  matchesFilters,
  ofRequests,
  SAMPLE_IPS,
  sampleBuckets,
  type SampleRecord,
} from '@/lib/faker/analytic/records';
import { sampleDayCurve, sampleOneOf, sampleRandom } from '@/lib/faker/random';

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
    request: { secrule: { name: 'block-admin', owner: { resourceId: 'rsrc_sample' } } },
    resourceId: 'rsrc_sample',
  },
  {
    method: '/sentinez.apps.security.v1.SecurityService/UpdateSecRateLimit',
    request: { id: 'srlim_sample-api-limit', updateMask: 'timeWindow' },
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
    method: '/sentinez.apps.security.v1.SecurityService/UpdateSecRuleset',
    request: {
      id: 'srs_sample',
      secruleset: { setting: SAMPLE_CRS_AFTER },
      updateMask: 'setting',
    },
    resourceId: 'rsrc_sample',
    change: [{ field: 'setting', before: SAMPLE_CRS_BEFORE, after: SAMPLE_CRS_AFTER }],
  },
  {
    method: '/sentinez.apps.iam.v1.IdentityAccessManagementService/Login',
    request: { emailOrUsername: 'mallory', password: 'REDACTED' },
    tenantless: true,
    code: 16,
    error: 'invalid email or password',
  },
  {
    method: '/sentinez.apps.iam.v1.IdentityAccessManagementService/Login',
    request: { emailOrUsername: 'alice', password: 'REDACTED' },
    tenantless: true,
  },
  {
    method: '/sentinez.apps.iam.v1.IdentityAccessManagementService/PasskeyLoginVerify',
    request: { credential: 'REDACTED' },
    tenantless: true,
    code: 16,
    error: 'passkey verification failed',
  },
  {
    method: '/sentinez.apps.iam.v1.IdentityAccessManagementService/GetUser',
    request: { id: 'user_alice' },
    tenantless: true,
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

// The samples are made per slot of SAMPLE_SLOT_MILLIS
const SAMPLE_SLOT_MILLIS = 7 * 60 * 1000;

// The number of samples of tenantId in the slot of index slot, random but
// the same for the same tenant and slot: busier by day, a burst of calls
// now and then, none in some slots
function sampleSlotCount(tenantId: string, slot: number): number {
  const random = sampleRandom(`${tenantId}|${slot}`);
  const day = sampleDayCurve(slot * SAMPLE_SLOT_MILLIS);
  const burst = random() < 0.05 ? 3 : 1;
  return Math.floor((1.4 + 0.6 * day + random()) * burst);
}

// The k-th sample of tenantId in the slot of index slot: its call (of
// calls, the sampleCalls of tenantId), user, address, latency and time in
// the slot random, but the same for the same tenant, slot and k
function sampleAuditLog(tenantId: string, calls: SampleCall[], slot: number, k: number): AuditLog {
  const random = sampleRandom(`${tenantId}|${slot}|${k}`);
  const oneOf = <T>(list: readonly T[]) => sampleOneOf(random, list);
  const call = oneOf(calls);
  const user = oneOf(SAMPLE_USERS);
  return {
    id: `aud_SAMPLE${String(slot).padStart(10, '0')}${String(k).padStart(2, '0')}`,
    method: call.method,
    userId: user.userId,
    name: user.name,
    controlPlane: ControlPlane.CONTROL_PLANE_PORTAL,
    remoteAddr: `${oneOf(SAMPLE_IPS)}:${40000 + Math.floor(random() * 20000)}`,
    userAgent: SAMPLE_AGENT,
    request: call.request,
    code: call.code ?? 0,
    error: call.error ?? '',
    latency: { seconds: 0, nanos: Math.round(3 + random() * random() * 120) * 1_000_000 },
    tenantId: call.tenantless ? '' : tenantId,
    timestamp: slot * SAMPLE_SLOT_MILLIS + Math.floor(random() * SAMPLE_SLOT_MILLIS),
    resourceId: call.resourceId ?? '',
    change: call.change && { fields: call.change },
  };
}

// The samples of tenantId over the 24 hours before end, the latest first:
// a random number of them per slot (sampleSlotCount)
function sampleAuditLogs(end: number, tenantId: string, withoutTenant = false): AuditLog[] {
  const last = Math.floor(end / SAMPLE_SLOT_MILLIS);
  const slots = Math.floor((24 * HOUR_MILLIS) / SAMPLE_SLOT_MILLIS);
  // the logs of no tenant are drawn among the calls about no tenant, or the
  // few of them would leave most slots empty
  const calls = sampleCalls(tenantId).filter((c) => !withoutTenant || c.tenantless);
  return Array.from({ length: slots }, (_, i) => last - i)
    .flatMap((slot) =>
      Array.from({ length: sampleSlotCount(tenantId, slot) }, (_, k) =>
        sampleAuditLog(tenantId, calls, slot, k),
      ),
    )
    .filter((l) => l.timestamp < end)
    .sort((a, b) => b.timestamp - a.timestamp);
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
export function filterSample(req: AuditSelection): AuditLog[] {
  const from = req.from?.getTime() ?? -Infinity;
  const to = req.to?.getTime() ?? Infinity;
  const tenantId = req.tenantIds[0] ?? SAMPLE_TENANT_ID;
  return sampleAuditLogs(req.to?.getTime() ?? Date.now(), tenantId, req.withoutTenant)
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

// The histogram of the samples: per bucket, resource and grpc code
export function sampleHistogram(req: GetAuditLogHistogramRequest): HistogramPoint[] {
  const resolution = req.resolution || HistogramResolution.HISTOGRAM_RESOLUTION_1_HOUR;
  const { millis } = RESOLUTION_BUCKETS[resolution];
  const buckets = sampleBuckets(resolution, req.to);
  const from = buckets[0]?.getTime() ?? 0;
  const to = from + buckets.length * millis;
  const logs = filterSample({ ...req, from: new Date(from), to: new Date(to) });
  return countPoints(new Map(), logs.map(recordOf), millis);
}

// The top values of the samples, as the API counts them
export function sampleTopValues(req: GetAuditLogTopValuesRequest): GetAuditLogTopValuesResponse {
  const to = req.to ?? new Date();
  const from = req.from ?? new Date(to.getTime() - HOUR_MILLIS);
  return countTopValues(filterSample({ ...req, from, to }).map(recordOf), req.fields, req.limit);
}
