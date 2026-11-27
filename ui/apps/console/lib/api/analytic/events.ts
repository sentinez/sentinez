import {
  EventKind,
  GetEgressLatencyRequest,
  GetEgressLatencyResponse,
  GetEventHistogramRequest,
  GetEventHistogramResponse,
  GetEventTopValuesRequest,
  GetEventTopValuesResponse,
  ListSecRulesetEventsRequest,
  ListSecRulesetEventsResponse,
  ListHttpEgressEventsRequest,
  ListHttpEgressEventsResponse,
  ListHttpIngressEventsRequest,
  ListHttpIngressEventsResponse,
  ListSecRateLimitEventsRequest,
  ListSecRateLimitEventsResponse,
  ListSecRuleEventsRequest,
  ListSecRuleEventsResponse,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { paginate } from '@/lib/api/pages';
import {
  ofFilters,
  present,
  sampleEgressLatency,
  sampleEgressOfSelection,
  sampleHistogram,
  sampleIngressRange,
  sampleRateLimitOf,
  sampleRulesetOf,
  sampleSecRuleOf,
  sampleTopValues,
} from '@/lib/faker/analytic/events';
import { ofRequests } from '@/lib/faker/analytic/records';
import { type ApiOptions, getJSON, intervalOf } from './common';

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
    () => ({
      points: sampleHistogram(full),
      interval: intervalOf(full.resolution),
      truncated: false,
    }),
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
      events: paginate(sampleEgressOfSelection(full), full.page),
    }),
    options,
  );
}

// GET /analytic/events/http-egress/latency
export async function getEgressLatency(
  req: Partial<GetEgressLatencyRequest> = {},
  options?: ApiOptions,
): Promise<GetEgressLatencyResponse> {
  const full = GetEgressLatencyRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/http-egress/latency',
    full,
    { req: GetEgressLatencyRequest, resp: GetEgressLatencyResponse },
    () => sampleEgressLatency(full),
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
          .filter(ofFilters(EventKind.EVENT_KIND_SECRULE, full.filters))
          .map(sampleSecRuleOf)
          .filter(present)
          .filter(ofRequests(full.requestIds, (e) => e.requestId)),
        full.page,
      ),
    }),
    options,
  );
}

// GET /analytic/events/secratelimit
export async function listSecRateLimitEvents(
  req: Partial<ListSecRateLimitEventsRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRateLimitEventsResponse> {
  const full = ListSecRateLimitEventsRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/secratelimit',
    full,
    { req: ListSecRateLimitEventsRequest, resp: ListSecRateLimitEventsResponse },
    () => ({
      events: paginate(
        sampleIngressRange(full)
          .filter(ofFilters(EventKind.EVENT_KIND_SECRATELIMIT, full.filters))
          .map(sampleRateLimitOf)
          .filter(present)
          .filter(ofRequests(full.requestIds, (e) => e.requestId)),
        full.page,
      ),
    }),
    options,
  );
}

// GET /analytic/events/secruleset
export async function listSecRulesetEvents(
  req: Partial<ListSecRulesetEventsRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRulesetEventsResponse> {
  const full = ListSecRulesetEventsRequest.fromPartial(req);
  return getJSON(
    '/analytic/events/secruleset',
    full,
    { req: ListSecRulesetEventsRequest, resp: ListSecRulesetEventsResponse },
    () => ({
      events: paginate(
        sampleIngressRange(full)
          .filter(ofFilters(EventKind.EVENT_KIND_SECRULESET, full.filters))
          .map(sampleRulesetOf)
          .filter(present)
          .filter(ofRequests(full.requestIds, (e) => e.requestId)),
        full.page,
      ),
    }),
    options,
  );
}
