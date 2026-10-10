import { Resource } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import {
  CreateResourceRequest,
  CreateResourceResponse,
  DeleteResourceRequest,
  DeleteResourceResponse,
  GetResourceByDomainRequest,
  GetResourceByDomainResponse,
  GetResourceRequest,
  GetResourceResponse,
  ListResourceRequest,
  ListResourceResponse,
  UpdateResourceRequest,
  UpdateResourceResponse,
} from '@sentinez/proto/sentinez/apps/tenant/v1/tenant';
import { Plan, Status } from '@sentinez/proto/sentinez/types/v1/known';
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { paginate, toQuery } from '@/lib/api/pages';
// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

// ─── Sample data (used only when USE_SAMPLE is true) ─────────────────────────

const upstream = (server: string, protocol = 'PROXY_PROTOCOL_HTTPS') => ({ server, protocol });

// Written as proto JSON (the wire format), parsed with Resource.fromJSON
export const SAMPLE_RESOURCES: Resource[] = [
  {
    metadata: { createdAt: '2026-09-01T08:00:00Z', updatedAt: '2026-10-05T10:30:00Z' },
    id: '0199a6f0-4c1e-7a2b-9d3e-000000000001',
    resourceDomain: 'badcheese.s6z.io.vn',
    resourceName: 'badcheese',
    status: 'STATUS_ACTIVE',
    plan: 'PLAN_STANDARD',
    resourceSetting: {
      server: {
        name: 'badcheese',
        locations: [
          {
            location: '/',
            proxyRewrite: '/',
            proxyPass: [upstream('10.0.1.10'), upstream('10.0.1.11')],
            balanceStrategy: 'BALANCE_STRATEGY_ROUND_ROBIN',
            proxySetHeaders: { 'X-Forwarded-Proto': 'https' },
          },
          {
            location: '/api',
            proxyRewrite: '/',
            proxyPass: [upstream('api.badcheese.internal', 'PROXY_PROTOCOL_HTTP')],
            balanceStrategy: 'BALANCE_STRATEGY_ROUND_ROBIN',
            proxySetHeaders: {},
          },
        ],
      },
      security: {
        limiters: [{ timeWindow: '1m', maxRequests: '100', timeout: '5m' }],
      },
      trafficControl: {},
      delivery: {
        cdn: [
          {
            rule: {
              name: 'cache GET',
              status: 'STATUS_ACTIVE',
            },
          },
        ],
      },
      personal: {},
    },
  },
  {
    metadata: { createdAt: '2026-09-12T02:15:00Z', updatedAt: '2026-09-28T14:00:00Z' },
    id: '0199a6f0-4c1e-7a2b-9d3e-000000000002',
    resourceDomain: 'shop.example.com',
    resourceName: 'shop',
    status: 'STATUS_ACTIVE',
    plan: 'PLAN_PRO',
    resourceSetting: {
      server: {
        name: 'shop',
        locations: [
          {
            location: '/',
            proxyRewrite: '/',
            proxyPass: [upstream('203.0.113.20')],
            balanceStrategy: 'BALANCE_STRATEGY_ROUND_ROBIN',
            proxySetHeaders: {},
          },
        ],
      },
    },
  },
  {
    metadata: { createdAt: '2026-10-01T09:45:00Z', updatedAt: '2026-10-01T09:45:00Z' },
    id: '0199a6f0-4c1e-7a2b-9d3e-000000000003',
    resourceDomain: 'staging.example.com',
    resourceName: 'staging',
    status: 'STATUS_DISABLE',
    plan: 'PLAN_FREE',
  },
].map((r) => Resource.fromJSON(r));

const PLANS = [Plan.PLAN_FREE, Plan.PLAN_STANDARD, Plan.PLAN_PRO];

// Extra generated resources so the list has several pages to page through
const generatedResources: Resource[] = Array.from({ length: 20 }, (_, i) => {
  const n = i + 1;
  const day = String(1 + (n % 28)).padStart(2, '0');
  return {
    metadata: {
      createdAt: new Date(`2026-08-${day}T00:00:00Z`),
      updatedAt: new Date(`2026-09-${day}T00:00:00Z`),
    },
    id: `0199a6f0-4c1e-7a2b-9d3e-1000000000${String(n).padStart(2, '0')}`,
    resourceDomain: `site${n}.example.com`,
    resourceName: `site${n}`,
    status: n % 4 === 0 ? Status.STATUS_DISABLE : Status.STATUS_ACTIVE,
    plan: PLANS[n % PLANS.length] ?? Plan.PLAN_FREE,
  };
});

let sampleStore: Resource[] = [...SAMPLE_RESOURCES, ...generatedResources];

function withFallback<T>(label: string, fallback: () => T) {
  return async (call: () => Promise<T>): Promise<T> => {
    try {
      return await call();
    } catch (err: any) {
      if (!USE_SAMPLE || axios.isCancel(err)) throw err;
      console.warn(`[tenant] ${label} failed, using sample response:`, err?.message);
      return fallback();
    }
  };
}

/** Older callers expect null instead of an error; keep that contract */
async function orNull<T>(label: string, run: () => Promise<T>): Promise<T | null> {
  try {
    return await run();
  } catch (err: any) {
    if (axios.isCancel(err)) throw err;
    console.error(`Error ${label}:`, err?.message);
    return null;
  }
}

function filterSample(req: Partial<ListResourceRequest>): Resource[] {
  return sampleStore.filter(
    (r) =>
      (!req.resourceDomain || r.resourceDomain === req.resourceDomain) &&
      (!req.resourceName || r.resourceName === req.resourceName) &&
      (!req.status || r.status === req.status) &&
      (!req.plan || r.plan === req.plan),
  );
}

// GET /tenant/resource?id=
export async function getResource(req: Partial<GetResourceRequest>, options?: ApiOptions) {
  return orNull('fetching resource', () =>
    withFallback<GetResourceResponse>('getResource', () => ({
      resource: sampleStore.find((r) => r.id === req.id),
    }))(async () => {
      const resp = await axios.get(`${API_BASE_PATH}/tenant/resource`, {
        params: toQuery(GetResourceRequest.toJSON(GetResourceRequest.fromPartial(req))),
        signal: options?.signal,
      });
      return GetResourceResponse.fromJSON(resp.data ?? {});
    }),
  );
}

// GET /tenant/resource/{resourceDomain}
export async function getResourceByDomain(req: GetResourceByDomainRequest, options?: ApiOptions) {
  return orNull('fetching resource by domain', () =>
    withFallback<GetResourceByDomainResponse>('getResourceByDomain', () => ({
      resource: sampleStore.find((r) => r.resourceDomain === req.resourceDomain),
    }))(async () => {
      const domain = encodeURIComponent(req.resourceDomain);
      const resp = await axios.get(`${API_BASE_PATH}/tenant/resource/${domain}`, {
        signal: options?.signal,
      });
      return GetResourceByDomainResponse.fromJSON(resp.data ?? {});
    }),
  );
}

// GET /tenant/resources
export async function listResources(req: Partial<ListResourceRequest> = {}, options?: ApiOptions) {
  return orNull('listing resources', () =>
    withFallback<ListResourceResponse>('listResources', () => {
      const resources = filterSample(req);
      return { total: resources.length, resources: paginate(resources, req.page) };
    })(async () => {
      const resp = await axios.get(`${API_BASE_PATH}/tenant/resources`, {
        params: toQuery(ListResourceRequest.toJSON(ListResourceRequest.fromPartial(req))),
        signal: options?.signal,
      });
      return ListResourceResponse.fromJSON(resp.data ?? {});
    }),
  );
}

// POST /tenant/resource
export async function createResource(
  req: CreateResourceRequest,
  options?: ApiOptions,
): Promise<CreateResourceResponse> {
  return withFallback<CreateResourceResponse>('createResource', () => {
    const now = new Date();
    const created: Resource = {
      ...req,
      id: crypto.randomUUID(),
      metadata: { createdAt: now, updatedAt: now },
    };
    sampleStore = [...sampleStore, created];
    return { resource: created };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/tenant/resource`,
      CreateResourceRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateResourceResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /tenant/resource
export async function updateResource(
  req: UpdateResourceRequest,
  options?: ApiOptions,
): Promise<UpdateResourceResponse> {
  return withFallback<UpdateResourceResponse>('updateResource', () => {
    sampleStore = sampleStore.map((r) =>
      r.id === req.id ? { ...r, ...req, metadata: { ...r.metadata, updatedAt: new Date() } } : r,
    );
    return {};
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/tenant/resource`,
      UpdateResourceRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateResourceResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /tenant/resource?id=
export async function deleteResource(
  req: DeleteResourceRequest,
  options?: ApiOptions,
): Promise<DeleteResourceResponse> {
  return withFallback<DeleteResourceResponse>('deleteResource', () => {
    sampleStore = sampleStore.filter((r) => r.id !== req.id);
    return {};
  })(async () => {
    const resp = await axios.delete(`${API_BASE_PATH}/tenant/resource`, {
      params: toQuery(DeleteResourceRequest.toJSON(req)),
      signal: options?.signal,
    });
    return DeleteResourceResponse.fromJSON(resp.data ?? {});
  });
}
