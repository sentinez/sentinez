import {
  EventKind,
  EventSelection,
  GetEgressLatencyRequest,
  GetEgressLatencyResponse,
  GetEventHistogramRequest,
  GetEventTopValuesRequest,
  GetEventTopValuesResponse,
  type HistogramPoint,
  type UpstreamLatency,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { HistogramResolution } from '@sentinez/proto/sentinez/types/v1/known';
import type { SecRulesetEvent } from '@sentinez/proto/sentinez/types/ruleset/v1/event';
import type { EgressEvent, IngressEvent } from '@sentinez/proto/sentinez/types/net/http/v1/event';
import { ActionType, RuleBehavior, RuleService } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import type { SecRuleEvent } from '@sentinez/proto/sentinez/types/secrule/v1/event';
import { SAMPLE_TENANT_ID } from '@/lib/faker/tenant';
import {
  DAY_MILLIS,
  HOUR_MILLIS,
  RESOLUTION_BUCKETS,
  resolutionOfSpan,
} from '@/lib/api/analytic/common';
import {
  countPoints,
  countTopValues,
  matchesFilters,
  ofRequests,
  SAMPLE_IPS,
  sampleBuckets,
  type SampleRecord,
} from '@/lib/faker/analytic/records';
import { type SampleLevel, sampleLevel, sampleOneOf, sampleRandom } from '@/lib/faker/random';

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
const SAMPLE_ADMIN_PATH = '/wp-admin.php';
// The anomaly score CRS adds per severity of a matched rule
const SEVERITY_SCORE = { critical: 5, warning: 3, notice: 2 };
// The CRS rules the sample requests match
type SampleRule = { id: number; severity: keyof typeof SEVERITY_SCORE; message: string };
const crit = (id: number, message: string): SampleRule => ({
  id,
  severity: 'critical',
  message,
});
const SAMPLE_ATTACKS: { query: string; rule: SampleRule }[] = [
  {
    query: "id=1' OR '1'='1",
    rule: crit(942100, 'SQL Injection Attack Detected via libinjection'),
  },
  {
    query: 'q=<script>alert(1)</script>',
    rule: crit(941100, 'XSS Attack Detected via libinjection'),
  },
  { query: 'file=../../etc/passwd', rule: crit(930120, 'OS File Access Attempt') },
  {
    query: 'cmd=;cat /etc/passwd',
    rule: crit(932160, 'Remote Command Execution: Unix Shell Code Found'),
  },
];
const SAMPLE_SCANNER_RULE = crit(913100, 'Found User-Agent associated with security scanner');
// Rules too weak alone to reach the inbound threshold (5): only logged
const SAMPLE_WEAK_RULES: SampleRule[] = [
  { id: 920350, severity: 'warning', message: 'Host header is a numeric IP address' },
  { id: 920300, severity: 'notice', message: 'Request Missing an Accept Header' },
  { id: 920320, severity: 'notice', message: 'Missing User Agent Header' },
];
// The anomaly score from which CRS denies a request
const SAMPLE_INBOUND_THRESHOLD = 5;

const pick = <T>(list: T[], k: number) => list[k % list.length] as T;

// A small hash of a string, to pick a sample value spread over the requests
const hashOf = (value: string) => [...value].reduce((h, c) => (h * 31 + c.charCodeAt(0)) >>> 0, 7);

// The n-th sample request of a resource in the bucket starting at time:
// its values random, but the same for the same resource, bucket and n
function sampleIngressEvent(resourceId: string, time: Date, millis: number, n: number) {
  const seed = Math.floor(time.getTime() / millis) + n * 7;
  const random = sampleRandom(`${resourceId}|${time.getTime()}|${n}`);
  const oneOf = <T>(list: readonly T[]) => sampleOneOf(random, list);
  const path = oneOf(SAMPLE_PATHS);
  // The probes of the admin page carry an attack, that CRS denies
  const admin = path === SAMPLE_ADMIN_PATH;
  const agent = oneOf(SAMPLE_AGENTS);
  const ip = oneOf(SAMPLE_IPS);
  return {
    id: `req_SAMPLE${String(seed).padStart(10, '0')}${String(n).padStart(4, '0')}`,
    method: oneOf(SAMPLE_METHODS),
    scheme: 'https',
    host: 'shop.example.com',
    path,
    queries: admin
      ? oneOf(SAMPLE_ATTACKS).query
      : random() < 0.3
        ? `q=shoes&page=${Math.floor(random() * 5) + 1}`
        : '',
    headers: `Host: shop.example.com\nAccept: */*\nUser-Agent: ${agent}\n`,
    protocol: random() < 0.5 ? 'HTTP/2.0' : 'HTTP/1.1',
    remoteAddress: `${ip}:${40000 + Math.floor(random() * 20000)}`,
    status: admin ? 403 : oneOf(SAMPLE_STATUSES),
    userAgent: agent,
    asn: 13335 + Math.floor(random() * 4),
    country: oneOf(SAMPLE_COUNTRIES),
    contentType: 'text/html; charset=utf-8',
    timestamp: time.getTime() + Math.floor(random() * millis),
    fingerprint: `t13d1516h2_${Math.floor(random() * 9999).toString(16)}`,
    clientIp: ip,
    resourceId,
    tenantId: SAMPLE_TENANT_ID,
  } satisfies IngressEvent;
}

// The traffic of the sample requests (sampleLevel)
const SAMPLE_REQUEST_LEVEL: SampleLevel = {
  quiet: 0.4,
  mean: 0.6,
  day: 0.3,
  noise: 0.4,
  burst: 2.5,
};

// The number of sample requests of a resource in the bucket starting at
// time (sampleLevel)
function sampleRequestCount(resourceId: string, time: Date): number {
  const level = sampleLevel(resourceId, time, SAMPLE_REQUEST_LEVEL);
  // large enough to test the layout: wide counts on the axes and in the tables
  return Math.max(1, Math.round(level * 40));
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
    .flatMap((time) =>
      ids.flatMap((resourceId) =>
        Array.from({ length: sampleRequestCount(resourceId, time) }, (_, n) =>
          sampleIngressEvent(resourceId, time, millis, n),
        ),
      ),
    )
    .sort((a, b) => b.timestamp - a.timestamp);
}

// RangeRequest the bounds every list and top request has
type RangeRequest = { resourceIds: string[]; from?: Date; to?: Date };

// The sample requests within a range, of the finest resolution covering it
export function sampleIngressRange(req: RangeRequest): IngressEvent[] {
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
    headers: `X-Forwarded-Prefix: /\nX-Real-IP: ${e.clientIp}\n`,
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
export function sampleSecRuleOf(e: IngressEvent): SecRuleEvent | null {
  if (e.status !== 403 || e.path === SAMPLE_ADMIN_PATH) return null;
  const login = e.path === '/login';
  return {
    ...requestOf(e),
    id: e.id.replace('req_', 'srevt_'),
    matchedSecruleIds: [pick(SAMPLE_SECRULE_IDS, login ? 1 : 0)],
    messages: [login ? 'path prefix "/login"' : 'path prefix "/api"'],
    action: login ? ActionType.ACTION_TYPE_CHALLENGE : ActionType.ACTION_TYPE_BLOCK,
    service: RuleService.RULE_SERVICE_SECRULE,
    tags: login ? { flow: 'login' } : {},
  };
}

// The rate limit rejection of a sample request answered 429
export function sampleRateLimitOf(e: IngressEvent): SecRuleEvent | null {
  if (e.status !== 429) return null;
  return {
    ...requestOf(e),
    id: e.id.replace('req_', 'srevt_'),
    matchedSecruleIds: [SAMPLE_RATELIMIT_ID],
    messages: ['client ip over 100 requests / minute'],
    action: ActionType.ACTION_TYPE_RATE_LIMIT,
    service: RuleService.RULE_SERVICE_SECRATELIMIT,
    tags: {},
  };
}

// The CRS rules a sample request matches: an attack on the admin page,
// with the scanner rule for curl, and now and then a weak rule alone on
// another request it lets through
function sampleRulesOf(e: IngressEvent): SampleRule[] {
  if (e.path === SAMPLE_ADMIN_PATH) {
    const attack = SAMPLE_ATTACKS.find((a) => a.query === e.queries);
    const rules = attack ? [attack.rule] : [];
    if (e.userAgent.startsWith('curl/')) rules.push(SAMPLE_SCANNER_RULE);
    return rules;
  }
  if (e.status === 403 || e.status === 429 || e.timestamp % 5 !== 0) return [];
  return [pick(SAMPLE_WEAK_RULES, hashOf(e.id))];
}

// The ruleset match of a sample request: denied over the threshold, else
// only logged
export function sampleRulesetOf(e: IngressEvent): SecRulesetEvent | null {
  const rules = sampleRulesOf(e);
  if (rules.length === 0) return null;
  const totalScore = rules.reduce((total, r) => total + SEVERITY_SCORE[r.severity], 0);
  return {
    ...requestOf(e),
    id: e.id.replace('req_', 'rsevt_'),
    matchedRuleIds: rules.map((r) => r.id),
    severities: rules.map((r) => r.severity),
    messages: rules.map((r) => r.message),
    totalScore,
    service: RuleService.RULE_SERVICE_SECRULESET,
    behavior:
      totalScore >= SAMPLE_INBOUND_THRESHOLD
        ? RuleBehavior.RULE_BEHAVIOR_DENY
        : RuleBehavior.RULE_BEHAVIOR_LOG,
  };
}

export const present = <T>(v: T | null): v is T => v !== null;

// Whether the event of kind of a sample request holds the field filters
export const ofFilters = (kind: EventKind, filters: string[]) => (e: IngressEvent) => {
  if (filters.length === 0) return true;
  const [record] = sampleRecordsOf(kind, [e]);
  return record !== undefined && matchesFilters(record, filters);
};

// The sample egress events of a selection: its range, field filters and
// requests, as the egress list and latency select them
export const sampleEgressOfSelection = (sel: EventSelection): EgressEvent[] =>
  sampleIngressRange(sel)
    .filter(ofFilters(EventKind.EVENT_KIND_HTTP_EGRESS, sel.filters))
    .map(sampleEgressOf)
    .filter(present)
    .filter(ofRequests(sel.requestIds, (e) => e.id));

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
    case EventKind.EVENT_KIND_SECRULE:
    case EventKind.EVENT_KIND_SECRATELIMIT:
      return requests
        .map(kind === EventKind.EVENT_KIND_SECRULE ? sampleSecRuleOf : sampleRateLimitOf)
        .filter(present)
        .map((e) => ({
          time: e.matchedTime,
          resourceId: e.resourceId,
          key: e.action,
          fields: {
            matched_secrule_ids: e.matchedSecruleIds,
            action: String(e.action),
            request_ip: e.requestIp,
            request_domain: e.requestDomain,
            request_path: e.requestPath,
          },
        }));
    case EventKind.EVENT_KIND_SECRULESET:
      return requests
        .map(sampleRulesetOf)
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
export function sampleHistogram(req: GetEventHistogramRequest): HistogramPoint[] {
  const { millis } = RESOLUTION_BUCKETS[req.resolution];
  const ids = req.resourceIds.length ? req.resourceIds : ['rsrc_sample'];
  const points = new Map<string, HistogramPoint>();
  for (const time of sampleBuckets(req.resolution, req.to)) {
    for (const resourceId of ids) {
      points.set(`${time.getTime()}|${resourceId}|0`, { time, resourceId, key: 0, total: 0 });
    }
  }
  // The events of the requests and filters only, as the API counts them
  const requests = sampleIngressWindow(req.resolution, ids, req.to).filter(
    ofRequests(req.requestIds, (e) => e.id),
  );
  const records = sampleRecordsOf(req.kind, requests).filter((r) => matchesFilters(r, req.filters));
  return countPoints(points, records, millis);
}

// The sample top values, the way the API counts them: the hour before `to`
// without `from`, the empty values left out, an event once per element
export function sampleTopValues(req: GetEventTopValuesRequest): GetEventTopValuesResponse {
  const to = req.to ?? new Date();
  const from = req.from ?? new Date(to.getTime() - HOUR_MILLIS);
  const requests = sampleIngressRange({ resourceIds: req.resourceIds, from, to }).filter(
    ofRequests(req.requestIds, (e) => e.id),
  );
  const records = sampleRecordsOf(req.kind, requests).filter((r) => matchesFilters(r, req.filters));
  return countTopValues(records, req.fields, req.limit);
}

// The p-th quantile (0 to 1) of micros, sorted ascending
const quantileOf = (sorted: number[], p: number) =>
  sorted[Math.min(sorted.length - 1, Math.floor(p * sorted.length))] ?? 0;

// The latency of the sample requests proxied to an upstream
function sampleLatencyOf(upstream: string, micros: number[]): UpstreamLatency {
  const sorted = micros.sort((a, b) => a - b);
  const sum = sorted.reduce((total, m) => total + m, 0);
  return {
    upstream,
    total: sorted.length,
    avgMicros: Math.round(sum / sorted.length),
    p50Micros: quantileOf(sorted, 0.5),
    p95Micros: quantileOf(sorted, 0.95),
    p99Micros: quantileOf(sorted, 0.99),
    maxMicros: sorted[sorted.length - 1] ?? 0,
  };
}

// The sample latency per upstream of the requests of the selection, the
// last 24 hours up to its end without a start
export function sampleEgressLatency(req: GetEgressLatencyRequest): GetEgressLatencyResponse {
  const selection = EventSelection.fromPartial(req.selection ?? {});
  const to = selection.to ?? new Date();
  const from = selection.from ?? new Date(to.getTime() - DAY_MILLIS);
  const byUpstream = new Map<string, number[]>();
  sampleEgressOfSelection({ ...selection, from, to }).forEach((e) => {
    const micros = byUpstream.get(e.upstream) ?? [];
    micros.push(e.latencyMicros);
    byUpstream.set(e.upstream, micros);
  });
  const upstreams = [...byUpstream]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([upstream, micros]) => sampleLatencyOf(upstream, micros));
  return { upstreams, from, to };
}
