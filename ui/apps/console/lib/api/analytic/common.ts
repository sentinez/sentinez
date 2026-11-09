import { HistogramResolution } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { toQuery } from '@/lib/api/pages';

// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

export function withFallback<T>(label: string, fallback: () => T) {
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

// The codecs of a request and of its response, generated from the proto
type Codecs<Req, Resp> = {
  req: { toJSON(message: Req): unknown };
  resp: { fromJSON(object: unknown): Resp };
};

// GET path with the query params of req, else the sample response; the
// repeated fields repeat their key (collectionFormat: multi)
export async function getJSON<Req, Resp>(
  path: string,
  req: Req,
  codecs: Codecs<Req, Resp>,
  sample: () => Resp,
  options?: ApiOptions,
): Promise<Resp> {
  return withFallback<Resp>(
    path,
    sample,
  )(async () => {
    const resp = await axios.get(`${API_BASE_PATH}${path}`, {
      params: toQuery(codecs.req.toJSON(req)),
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return codecs.resp.fromJSON(resp.data ?? {});
  });
}

const SECOND_MILLIS = 1000;
const MINUTE_MILLIS = 60 * SECOND_MILLIS;
export const HOUR_MILLIS = 60 * MINUTE_MILLIS;

// The width and number of the buckets of each resolution, as the API
// returns them (unspecified: minutes)
export const RESOLUTION_BUCKETS: Record<
  HistogramResolution,
  { millis: number; count: number; unit: 'second' | 'minute' | 'hour' }
> = {
  [HistogramResolution.HISTOGRAM_RESOLUTION_UNSPECIFIED]: {
    millis: MINUTE_MILLIS,
    count: 60,
    unit: 'minute',
  },
  [HistogramResolution.HISTOGRAM_RESOLUTION_SECOND]: {
    millis: SECOND_MILLIS,
    count: 60,
    unit: 'second',
  },
  [HistogramResolution.HISTOGRAM_RESOLUTION_MINUTE]: {
    millis: MINUTE_MILLIS,
    count: 60,
    unit: 'minute',
  },
  [HistogramResolution.HISTOGRAM_RESOLUTION_HOUR]: {
    millis: HOUR_MILLIS,
    count: 24,
    unit: 'hour',
  },
  [HistogramResolution.UNRECOGNIZED]: { millis: MINUTE_MILLIS, count: 60, unit: 'minute' },
};

// The resolutions of a histogram, the finest first
export const RESOLUTIONS = [
  HistogramResolution.HISTOGRAM_RESOLUTION_SECOND,
  HistogramResolution.HISTOGRAM_RESOLUTION_MINUTE,
  HistogramResolution.HISTOGRAM_RESOLUTION_HOUR,
];

// The most buckets a histogram shows, fetched in pages of the buckets the
// API returns for its resolution
const MAX_BUCKETS = 360;

// The pages of buckets a histogram of a resolution fetches at most
export const maxPagesOf = (resolution: HistogramResolution) =>
  Math.max(1, Math.floor(MAX_BUCKETS / RESOLUTION_BUCKETS[resolution].count));

// The time the buckets of a resolution cover
export const spanOf = (resolution: HistogramResolution) =>
  RESOLUTION_BUCKETS[resolution].millis * RESOLUTION_BUCKETS[resolution].count;

// The finest resolution whose buckets cover span (millis), the hours beyond
export function resolutionOfSpan(span: number): HistogramResolution {
  return (
    RESOLUTIONS.find((r) => span <= spanOf(r)) ?? HistogramResolution.HISTOGRAM_RESOLUTION_HOUR
  );
}

// The sample buckets of a resolution: those ending with the one of `to`,
// else the current one
export function sampleBuckets(resolution: HistogramResolution, to?: Date): Date[] {
  const { millis, count } = RESOLUTION_BUCKETS[resolution];
  const end = Math.floor((to ?? new Date()).getTime() / millis) + 1;
  return Array.from({ length: count }, (_, i) => new Date((end - count + i) * millis));
}

// Sample client ips: the first ones visit every hour, the others now and then
export const SAMPLE_IPS = [
  '203.0.113.7',
  '198.51.100.23',
  '192.0.2.41',
  '203.0.113.88',
  '198.51.100.5',
];
