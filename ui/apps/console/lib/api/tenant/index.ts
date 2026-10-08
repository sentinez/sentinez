import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { type Pages, pageQuery, paginate } from '@/lib/api/pages';
// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

/** Wire value of `typesv1Status` in tenant.swagger.json */
export type TenantStatus = 'STATUS_UNSPECIFIED' | 'STATUS_ACTIVE' | 'STATUS_DISABLE';

/** Wire value of `v1Plan` in tenant.swagger.json */
export type TenantPlan = 'PLAN_UNSPECIFIED' | 'PLAN_FREE' | 'PLAN_STANDARD' | 'PLAN_PRO';

/** Wire format of `v1Resource` in tenant.swagger.json */
export interface TenantResource {
  metadata?: { createdAt?: string; updatedAt?: string };
  id?: string;
  /** `v1Setting` (edge setting) as JSON, see Setting.fromJSON */
  resourceSetting?: Record<string, any>;
  resourceDomain?: string;
  resourceName?: string;
  status?: TenantStatus;
  plan?: TenantPlan;
}

/** Wire format of `v1CreateResourceRequest` in tenant.swagger.json */
export type CreateResourceRequest = Omit<TenantResource, 'id' | 'metadata'>;

/** Wire format of `v1UpdateResourceRequest` in tenant.swagger.json */
export type UpdateResourceRequest = Omit<TenantResource, 'metadata' | 'resourceSetting'>;

export interface ListResourcesParams {
  page?: Pages;
  resourceDomain?: string;
  resourceName?: string;
  status?: TenantStatus;
  plan?: TenantPlan;
}

/** Wire format of `v1ListResourceResponse` in tenant.swagger.json (total is int64) */
export interface ListResourcesResult {
  total?: string;
  resources?: TenantResource[];
}

// ─── Sample data (used only when USE_SAMPLE is true) ─────────────────────────

const upstream = (server: string, protocol = 'PROXY_PROTOCOL_HTTPS') => ({ server, protocol });

export const SAMPLE_RESOURCES: TenantResource[] = [
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
];

const PLANS: TenantPlan[] = ['PLAN_FREE', 'PLAN_STANDARD', 'PLAN_PRO'];

// Extra generated resources so the list has several pages to page through
const generatedResources: TenantResource[] = Array.from({ length: 20 }, (_, i) => {
  const n = i + 1;
  const day = String(1 + (n % 28)).padStart(2, '0');
  return {
    metadata: { createdAt: `2026-08-${day}T00:00:00Z`, updatedAt: `2026-09-${day}T00:00:00Z` },
    id: `0199a6f0-4c1e-7a2b-9d3e-1000000000${String(n).padStart(2, '0')}`,
    resourceDomain: `site${n}.example.com`,
    resourceName: `site${n}`,
    status: n % 4 === 0 ? 'STATUS_DISABLE' : 'STATUS_ACTIVE',
    plan: PLANS[n % PLANS.length],
  };
});

let sampleStore: TenantResource[] = [...SAMPLE_RESOURCES, ...generatedResources];

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

function filterSample(params?: ListResourcesParams): TenantResource[] {
  return sampleStore.filter(
    (r) =>
      (!params?.resourceDomain || r.resourceDomain === params.resourceDomain) &&
      (!params?.resourceName || r.resourceName === params.resourceName) &&
      (!params?.status || params.status === 'STATUS_UNSPECIFIED' || r.status === params.status) &&
      (!params?.plan || params.plan === 'PLAN_UNSPECIFIED' || r.plan === params.plan),
  );
}

// GET /tenant/resource?id=
export async function getResource(id: string, options?: ApiOptions) {
  return orNull('fetching resource', () =>
    withFallback<TenantResource | null>(
      'getResource',
      () => sampleStore.find((r) => r.id === id) ?? null,
    )(async () => {
      const resp = await axios.get(`${API_BASE_PATH}/tenant/resource`, {
        params: { id },
        signal: options?.signal,
      });
      return resp.data?.resource ?? null;
    }),
  );
}

// GET /tenant/resource/{resourceDomain}
export async function getResourceByDomain(domain: string, options?: ApiOptions) {
  return orNull('fetching resource by domain', () =>
    withFallback<TenantResource | null>(
      'getResourceByDomain',
      () => sampleStore.find((r) => r.resourceDomain === domain) ?? null,
    )(async () => {
      const endpoint = `${API_BASE_PATH}/tenant/resource/${encodeURIComponent(domain)}`;
      const resp = await axios.get(endpoint, { signal: options?.signal });
      return resp.data?.resource ?? null;
    }),
  );
}

// GET /tenant/resources
export async function listResources(params?: ListResourcesParams, options?: ApiOptions) {
  return orNull('listing resources', () =>
    withFallback<ListResourcesResult>('listResources', () => {
      const resources = filterSample(params);
      return { total: String(resources.length), resources: paginate(resources, params?.page) };
    })(async () => {
      const { page, ...filters } = params ?? {};
      const resp = await axios.get(`${API_BASE_PATH}/tenant/resources`, {
        params: { ...filters, ...pageQuery(page) },
        signal: options?.signal,
      });
      return resp.data ?? {};
    }),
  );
}

// POST /tenant/resource
export async function createResource(
  data: CreateResourceRequest,
  options?: ApiOptions,
): Promise<TenantResource> {
  return withFallback('createResource', () => {
    const now = new Date().toISOString();
    const created: TenantResource = {
      ...data,
      id: crypto.randomUUID(),
      metadata: { createdAt: now, updatedAt: now },
    };
    sampleStore = [...sampleStore, created];
    return created;
  })(async () => {
    const resp = await axios.post(`${API_BASE_PATH}/tenant/resource`, data, {
      signal: options?.signal,
    });
    return resp.data?.resource ?? {};
  });
}

// PUT /tenant/resource
export async function updateResource(
  data: UpdateResourceRequest,
  options?: ApiOptions,
): Promise<void> {
  return withFallback<void>('updateResource', () => {
    sampleStore = sampleStore.map((r) =>
      r.id === data.id
        ? { ...r, ...data, metadata: { ...r.metadata, updatedAt: new Date().toISOString() } }
        : r,
    );
  })(async () => {
    await axios.put(`${API_BASE_PATH}/tenant/resource`, data, { signal: options?.signal });
  });
}

// DELETE /tenant/resource?id=
export async function deleteResource(id: string, options?: ApiOptions): Promise<void> {
  return withFallback<void>('deleteResource', () => {
    sampleStore = sampleStore.filter((r) => r.id !== id);
  })(async () => {
    await axios.delete(`${API_BASE_PATH}/tenant/resource`, {
      params: { id },
      signal: options?.signal,
    });
  });
}
