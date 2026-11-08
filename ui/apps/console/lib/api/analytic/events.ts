import {
  EventKind,
  GetEventHistogramRequest,
  GetEventHistogramResponse,
  GetEventTopValuesRequest,
  GetEventTopValuesResponse,
  type FieldTopValues,
  type HistogramPoint,
  HistogramResolution,
  ListCoreRuleEventsRequest,
  ListCoreRuleEventsResponse,
  ListHttpEgressEventsRequest,
  ListHttpEgressEventsResponse,
  ListHttpIngressEventsRequest,
  ListHttpIngressEventsResponse,
  ListRateLimitEventsRequest,
  ListRateLimitEventsResponse,
  ListSecRuleEventsRequest,
  ListSecRuleEventsResponse,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import type { CoreRuleEvent } from '@sentinez/proto/sentinez/types/coreruleset/v1/event';
import type { EgressEvent, IngressEvent } from '@sentinez/proto/sentinez/types/net/http/v1/event';
import { ActionType, RuleBehavior, RuleService } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import type { SecRuleEvent } from '@sentinez/proto/sentinez/types/secrule/v1/event';
import { paginate } from '@/lib/api/pages';
import { SAMPLE_TENANT_ID } from '@/lib/api/tenant';
import {
  type ApiOptions,
  HOUR_MILLIS,
  getJSON,
  RESOLUTION_BUCKETS,
  SAMPLE_IPS,
  resolutionOfSpan,
  sampleBuckets,
} from './common';

// ---- Sample events, used as fallback when the API call fails ----
// Every kind derives from the sample requests received by the edge, so the
// pages agree: the blocked ones matched a rule, the others were proxied.

const SAMPLE_METHODS = ['GET', 'GET', 'GET', 'POST', 'GET', 'PUT', 'GET', 'DELETE'];
const SAMPLE_PATHS = ['/', '/api/products', '/login', '/static/app.js', '/cart', '/wp-admin.php'];
const SAMPLE_STATUSES = [200, 200, 429, 304, 200, 404, 200, 201, 403, 500, 429, 301];
const SAMPLE_COUNTRIES = ['VN', 'US', 'SG', 'JP', 'DE'];
const SAMPLE_AGENTS = [
  'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0 Safari/537.36',
  'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148',
  'curl/8.9.1',
];
const SAMPLE_SECRULE_IDS = ['srule_sample-block-admin', 'srule_sample-challenge-login'];
const SAMPLE_RATELIMIT_ID = 'srule_sample-api-limit';
const SAMPLE_CORE_RULES = [
  { id: 941100, message: 'XSS Attack Detected via libinjection' },
  { id: 942100, message: 'SQL Injection Attack Detected via libinjection' },
  { id: 930120, message: 'OS File Access Attempt' },
];

const pick = <T>(list: T[], k: number) => list[k % list.length] as T;

// The n-th sample request of a resource in the bucket starting at time
function sampleIngressEvent(resourceId: string, time: Date, millis: number, n: number) {
  const seed = Math.floor(time.getTime() / millis) + n * 7;
  return {
    id: `req_SAMPLE${String(seed).padStart(10, '0')}${String(n).padStart(4, '0')}`,
    method: pick(SAMPLE_METHODS, seed),
    scheme: 'https',
    host: 'shop.example.com',
    path: pick(SAMPLE_PATHS, seed + n),
    queries: seed % 3 === 0 ? `q=shoes&page=${(seed % 5) + 1}` : '',
    headers: `Host: shop.example.com\nAccept: */*\nUser-Agent: ${pick(SAMPLE_AGENTS, seed)}\n`,
    protocol: seed % 2 === 0 ? 'HTTP/2.0' : 'HTTP/1.1',
    remoteAddress: `${pick(SAMPLE_IPS, seed + n)}:${40000 + (seed % 20000)}`,
    status: pick(SAMPLE_STATUSES, seed * 5 + n),
    userAgent: pick(SAMPLE_AGENTS, seed),
    asn: 13335 + (seed % 4),
    country: pick(SAMPLE_COUNTRIES, seed + n),
    contentType: 'text/html; charset=utf-8',
    timestamp: time.getTime() + ((n * 997) % millis),
    fingerprint: `t13d1516h2_${(seed % 9999).toString(16)}`,
    clientIp: pick(SAMPLE_IPS, seed + n),
    resourceId,
    tenantId: SAMPLE_TENANT_ID,
  } satisfies IngressEvent;
}

// The sample requests of the buckets of a resolution, the latest first
function sampleIngressWindow(
  resolution: HistogramResolution,
  resourceIds: string[],
  to?: Date,
): IngressEvent[] {
  const { millis } = RESOLUTION_BUCKETS[resolution];
  const ids = resourceIds.length ? resourceIds : ['rsrc_sample'];
  return sampleBuckets(resolution, to)
    .flatMap((time, i) =>
      ids.flatMap((resourceId) =>
        Array.from({ length: 2 + Math.round(3 + 3 * Math.sin(i / 4)) }, (_, n) =>
          sampleIngressEvent(resourceId, time, millis, n),
        ),
      ),
    )
    .sort((a, b) => b.timestamp - a.timestamp);
}

// RangeRequest the bounds every list and top request has
type RangeRequest = { resourceIds: string[]; from?: Date; to?: Date };

// The sample requests within a range, of the finest resolution covering it
function sampleIngressRange(req: RangeRequest): IngressEvent[] {
  const span = (req.to ?? new Date()).getTime() - (req.from?.getTime() ?? 0);
  const resolution = resolutionOfSpan(span);
  const from = req.from?.getTime() ?? -Infinity;
  const to = req.to?.getTime() ?? Infinity;
  return sampleIngressWindow(resolution, req.resourceIds, req.to).filter(
    (e) => e.timestamp >= from && e.timestamp < to,
  );
}

// The leg of a sample request to the origin; none for a rejected one
function sampleEgressOf(e: IngressEvent): EgressEvent | null {
  if (e.status === 403 || e.status === 429) return null;
  return {
    id: e.id,
    upstream: e.path.startsWith('/api') ? 'http://10.0.0.12:8080' : 'http://10.0.0.11:3000',
    method: e.method,
    scheme: 'http',
    path: e.path,
    queries: e.queries,
    status: e.status === 500 ? 502 : e.status,
    contentType: e.contentType,
    contentLength: e.status === 304 ? 0 : 1024 + (e.timestamp % 50000),
    latencyMicros: 2000 + (e.timestamp % 400) * 750,
    timestamp: e.timestamp,
    resourceId: e.resourceId,
    tenantId: e.tenantId,
    protocol: e.protocol,
  };
}

// The request fields a rule event repeats
const requestOf = (e: IngressEvent) => ({
  requestDomain: e.host,
  requestScheme: e.scheme,
  requestPath: e.path,
  requestQuery: e.queries,
  requestIp: e.clientIp,
  requestId: e.id,
  resourceId: e.resourceId,
  tenantId: e.tenantId,
  matchedTime: e.timestamp,
});

// The secrule match of a sample request blocked by a secrule
function sampleSecRuleOf(e: IngressEvent): SecRuleEvent | null {
  if (e.status !== 403 || e.path === '/wp-admin.php') return null;
  const login = e.path === '/login';
  return {
    ...requestOf(e),
    id: e.id.replace('req_', 'srevt_'),
    matchedSecruleIds: [pick(SAMPLE_SECRULE_IDS, login ? 1 : 0)],
    messages: [login ? 'path prefix "/login"' : 'path prefix "/api"'],
    action: { type: login ? ActionType.ACTION_TYPE_CHALLENGE : ActionType.ACTION_TYPE_BLOCK },
    service: RuleService.RULE_SERVICE_SEC_RULE,
    tags: login ? { flow: 'login' } : {},
  };
}

// The rate limit rejection of a sample request answered 429
function sampleRateLimitOf(e: IngressEvent): SecRuleEvent | null {
  if (e.status !== 429) return null;
  return {
    ...requestOf(e),
    id: e.id.replace('req_', 'srevt_'),
    matchedSecruleIds: [SAMPLE_RATELIMIT_ID],
    messages: ['client ip over 100 requests / minute'],
    action: { type: ActionType.ACTION_TYPE_BLOCK },
    service: RuleService.RULE_SERVICE_RATE_LIMIT,
    tags: {},
  };
}

// The core rule match of a sample request to the admin page
function sampleCoreRuleOf(e: IngressEvent): CoreRuleEvent | null {
  if (e.status !== 403 || e.path !== '/wp-admin.php') return null;
  const rules = SAMPLE_CORE_RULES.slice(0, 1 + (e.timestamp % 3));
  return {
    ...requestOf(e),
    id: e.id.replace('req_', 'crevt_'),
    matchedRuleIds: rules.map((r) => r.id),
    severities: rules.map(() => 'critical'),
    messages: rules.map((r) => r.message),
    totalScore: rules.length * 5,
    service: RuleService.RULE_SERVICE_CORE_RULESETS,
    behavior: RuleBehavior.RULE_BEHAVIOR_DENY,
  };
}

const present = <T>(v: T | null): v is T => v !== null;

// Whether a sample event is of one of the requests, as the API filters
// request_ids: every one when empty
const ofRequests =
  <T>(requestIds: string[], requestOf: (e: T) => string) =>
  (e: T) =>
    requestIds.length === 0 || requestIds.includes(requestOf(e));

// Whether a sample record holds the values of the field filters, as the
// API filters them: an element of a repeated field
const matchesFilters = (r: SampleRecord, filters: string[]) =>
  filters.every((f) => {
    const at = f.indexOf(':');
    const want = f.slice(at + 1);
    const value = r.fields[f.slice(0, at)];
    return Array.isArray(value) ? value.includes(want) : value === want;
  });

// Whether the event of kind of a sample request holds the field filters
const ofFilters = (kind: EventKind, filters: string[]) => (e: IngressEvent) => {
  if (filters.length === 0) return true;
  const [record] = sampleRecordsOf(kind, [e]);
  return record !== undefined && matchesFilters(record, filters);
};

// SampleRecord a sample event as the histogram and the top values count it
type SampleRecord = {
  time: number;
  resourceId: string;
  key: number;
  fields: Record<string, string | string[]>;
};

// The records of the events of kind derived from the sample requests, the
// fields named as the API names them, in the order of a top without any
function sampleRecordsOf(kind: EventKind, requests: IngressEvent[]): SampleRecord[] {
  switch (kind) {
    case EventKind.EVENT_KIND_HTTP_INGRESS:
      return requests.map((e) => ({
        time: e.timestamp,
        resourceId: e.resourceId,
        key: e.status,
        fields: {
          method: e.method,
          status: String(e.status),
          host: e.host,
          client_ip: e.clientIp,
          country: e.country,
          protocol: e.protocol,
          path: e.path,
          user_agent: e.userAgent,
        },
      }));
    case EventKind.EVENT_KIND_HTTP_EGRESS:
      return requests
        .map(sampleEgressOf)
        .filter(present)
        .map((e) => ({
          time: e.timestamp,
          resourceId: e.resourceId,
          key: e.status,
          fields: {
            upstream: e.upstream,
            status: String(e.status),
            method: e.method,
            path: e.path,
            content_type: e.contentType,
            protocol: e.protocol,
          },
        }));
    case EventKind.EVENT_KIND_SEC_RULE:
    case EventKind.EVENT_KIND_RATE_LIMIT:
      return requests
        .map(kind === EventKind.EVENT_KIND_SEC_RULE ? sampleSecRuleOf : sampleRateLimitOf)
        .filter(present)
        .map((e) => ({
          time: e.matchedTime,
          resourceId: e.resourceId,
          key: 0,
          fields: {
            matched_secrule_ids: e.matchedSecruleIds,
            action: String(e.action?.type ?? ''),
            request_ip: e.requestIp,
            request_domain: e.requestDomain,
            request_path: e.requestPath,
          },
        }));
    case EventKind.EVENT_KIND_CORE_RULE:
      return requests
        .map(sampleCoreRuleOf)
        .filter(present)
        .map((e) => ({
          time: e.matchedTime,
          resourceId: e.resourceId,
          key: e.behavior,
          fields: {
            matched_rule_ids: e.matchedRuleIds.map(String),
            severities: e.severities,
            behavior: String(e.behavior),
            total_score: String(e.totalScore),
            request_ip: e.requestIp,
            request_domain: e.requestDomain,
            request_path: e.requestPath,
          },
        }));
    default:
      return [];
  }
}

// The sample histogram of kind, one point per bucket and resource at least
function sampleHistogram(req: GetEventHistogramRequest): HistogramPoint[] {
  const { millis } = RESOLUTION_BUCKETS[req.resolution];
  const ids = req.resourceIds.length ? req.resourceIds : ['rsrc_sample'];
  const points = new Map<string, HistogramPoint>();
  for (const time of sampleBuckets(req.resolution, req.to)) {
    for (const resourceId of ids) {
      points.set(`${time.getTime()}|${resourceId}|0`, { time, resourceId, key: 0, total: 0 });
    }
  }
  const requests = sampleIngressWindow(req.resolution, ids, req.to);
  for (const r of sampleRecordsOf(req.kind, requests)) {
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

// The sample top values, the way the API counts them: the hour before `to`
// without `from`, the empty values left out, an event once per element
function sampleTopValues(req: GetEventTopValuesRequest): GetEventTopValuesResponse {
  const to = req.to ?? new Date();
  const from = req.from ?? new Date(to.getTime() - HOUR_MILLIS);
  const records = sampleRecordsOf(
    req.kind,
    sampleIngressRange({ resourceIds: req.resourceIds, from, to }),
  ).filter((r) => matchesFilters(r, req.filters));
  const limit = Math.min(req.limit > 0 ? req.limit : 5, 20);
  const names = req.fields.length ? req.fields : Object.keys(records[0]?.fields ?? {});
  const fields: FieldTopValues[] = names.map((field) => {
    const counts = new Map<string, number>();
    for (const r of records) {
      const raw = r.fields[field] ?? '';
      for (const value of new Set(Array.isArray(raw) ? raw : [raw])) {
        if (value) counts.set(value, (counts.get(value) ?? 0) + 1);
      }
    }
    const values = [...counts.entries()]
      .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
      .slice(0, limit)
      .map(([value, total]) => ({ value, total, share: total / records.length }));
    return { field, values };
  });
  return { total: records.length, fields };
}

// GET /analytic/events/histogram
export async function getEventHistogram(
  req: Partial<GetEventHistogramRequest> = {},
  options?: ApiOptions,
): Promise<GetEventHistogramResponse> {
  const full = GetEventHistogramRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/histogram',
    full,
    { req: GetEventHistogramRequest, resp: GetEventHistogramResponse },
    () => ({ points: sampleHistogram(full) }),
    options,
  );
}

// GET /analytic/events/top-values
export async function getEventTopValues(
  req: Partial<GetEventTopValuesRequest> = {},
  options?: ApiOptions,
): Promise<GetEventTopValuesResponse> {
  const full = GetEventTopValuesRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/top-values',
    full,
    { req: GetEventTopValuesRequest, resp: GetEventTopValuesResponse },
    () => sampleTopValues(full),
    options,
  );
}

// GET /analytic/events/http-ingress
export async function listHttpIngressEvents(
  req: Partial<ListHttpIngressEventsRequest> = {},
  options?: ApiOptions,
): Promise<ListHttpIngressEventsResponse> {
  const full = ListHttpIngressEventsRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/http-ingress',
    full,
    { req: ListHttpIngressEventsRequest, resp: ListHttpIngressEventsResponse },
    () => ({
      events: paginate(
        sampleIngressRange(full)
          .filter(ofFilters(EventKind.EVENT_KIND_HTTP_INGRESS, full.filters))
          .filter(ofRequests(full.requestIds, (e) => e.id)),
        full.page,
      ),
    }),
    options,
  );
}

// GET /analytic/events/http-egress
export async function listHttpEgressEvents(
  req: Partial<ListHttpEgressEventsRequest> = {},
  options?: ApiOptions,
): Promise<ListHttpEgressEventsResponse> {
  const full = ListHttpEgressEventsRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/http-egress',
    full,
    { req: ListHttpEgressEventsRequest, resp: ListHttpEgressEventsResponse },
    () => ({
      events: paginate(
        sampleIngressRange(full)
          .filter(ofFilters(EventKind.EVENT_KIND_HTTP_EGRESS, full.filters))
          .map(sampleEgressOf)
          .filter(present)
          .filter(ofRequests(full.requestIds, (e) => e.id)),
        full.page,
      ),
    }),
    options,
  );
}

// GET /analytic/events/secrule
export async function listSecRuleEvents(
  req: Partial<ListSecRuleEventsRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRuleEventsResponse> {
  const full = ListSecRuleEventsRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/secrule',
    full,
    { req: ListSecRuleEventsRequest, resp: ListSecRuleEventsResponse },
    () => ({
      events: paginate(
        sampleIngressRange(full)
          .filter(ofFilters(EventKind.EVENT_KIND_SEC_RULE, full.filters))
          .map(sampleSecRuleOf)
          .filter(present)
          .filter(ofRequests(full.requestIds, (e) => e.requestId)),
        full.page,
      ),
    }),
    options,
  );
}

// GET /analytic/events/ratelimit
export async function listRateLimitEvents(
  req: Partial<ListRateLimitEventsRequest> = {},
  options?: ApiOptions,
): Promise<ListRateLimitEventsResponse> {
  const full = ListRateLimitEventsRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/ratelimit',
    full,
    { req: ListRateLimitEventsRequest, resp: ListRateLimitEventsResponse },
    () => ({
      events: paginate(
        sampleIngressRange(full)
          .filter(ofFilters(EventKind.EVENT_KIND_RATE_LIMIT, full.filters))
          .map(sampleRateLimitOf)
          .filter(present)
          .filter(ofRequests(full.requestIds, (e) => e.requestId)),
        full.page,
      ),
    }),
    options,
  );
}

// GET /analytic/events/coreruleset
export async function listCoreRuleEvents(
  req: Partial<ListCoreRuleEventsRequest> = {},
  options?: ApiOptions,
): Promise<ListCoreRuleEventsResponse> {
  const full = ListCoreRuleEventsRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/coreruleset',
    full,
    { req: ListCoreRuleEventsRequest, resp: ListCoreRuleEventsResponse },
    () => ({
      events: paginate(
        sampleIngressRange(full)
          .filter(ofFilters(EventKind.EVENT_KIND_CORE_RULE, full.filters))
          .map(sampleCoreRuleOf)
          .filter(present)
          .filter(ofRequests(full.requestIds, (e) => e.requestId)),
        full.page,
      ),
    }),
    options,
  );
}
