import {
  CaCertificate,
  CaCertificateUsage,
  Member,
  Resource,
  Tenant,
} from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import {
  CreateCaCertificateRequest,
  CreateCaCertificateResponse,
  CreateMemberRequest,
  CreateMemberResponse,
  CreateResourceRequest,
  CreateResourceResponse,
  CreateTenantRequest,
  CreateTenantResponse,
  DeleteCaCertificateRequest,
  DeleteCaCertificateResponse,
  DeleteMemberRequest,
  DeleteMemberResponse,
  DeleteResourceRequest,
  DeleteResourceResponse,
  DeleteTenantRequest,
  DeleteTenantResponse,
  GetCaCertificateRequest,
  GetCaCertificateResponse,
  GetMemberRequest,
  GetMemberResponse,
  GetResourceByDomainRequest,
  GetResourceByDomainResponse,
  GetResourceRequest,
  GetResourceResponse,
  GetTenantRequest,
  GetTenantResponse,
  ListCaCertificateRequest,
  ListCaCertificateResponse,
  ListMemberRequest,
  ListMemberResponse,
  ListResourceRequest,
  ListResourceResponse,
  ListTenantRequest,
  ListTenantResponse,
  UpdateMemberRequest,
  UpdateMemberResponse,
  UpdateResourceRequest,
  UpdateResourceResponse,
  UpdateTenantRequest,
  UpdateTenantResponse,
} from '@sentinez/proto/sentinez/apps/tenant/v1/tenant';
import { Server } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import axios, { AxiosError, AxiosHeaders } from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { ALL_SAMPLE_USERS } from '@/lib/api/iam/user';
import { paginate, toQuery } from '@/lib/api/pages';
// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

// ─── Sample data (used only when USE_SAMPLE is true) ─────────────────────────

// Tenant that owns every sample resource
export const SAMPLE_TENANT_ID = 'tnt_console-sample';

const upstream = (server: string, protocol = 'PROXY_PROTOCOL_HTTPS') => ({ server, protocol });

// Written as proto JSON (the wire format), parsed with Resource.fromJSON
export const SAMPLE_RESOURCES: Resource[] = [
  {
    metadata: { createdAt: '2026-09-01T08:00:00Z', updatedAt: '2026-10-05T10:30:00Z' },
    id: 'rsrc_0199a6f0-4c1e-7a2b-9d3e-000000000001',
    tenantId: SAMPLE_TENANT_ID,
    status: 'STATUS_ACTIVE',
    server: {
      name: 'badcheese',
      resourceId: 'rsrc_0199a6f0-4c1e-7a2b-9d3e-000000000001',
      resourceDomain: 'badcheese.s6z.io.vn',
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
  },
  {
    metadata: { createdAt: '2026-09-12T02:15:00Z', updatedAt: '2026-09-28T14:00:00Z' },
    id: 'rsrc_0199a6f0-4c1e-7a2b-9d3e-000000000002',
    tenantId: SAMPLE_TENANT_ID,
    status: 'STATUS_ACTIVE',
    server: {
      name: 'shop',
      resourceId: 'rsrc_0199a6f0-4c1e-7a2b-9d3e-000000000002',
      resourceDomain: 'shop.example.com',
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
  {
    metadata: { createdAt: '2026-10-01T09:45:00Z', updatedAt: '2026-10-01T09:45:00Z' },
    id: 'rsrc_0199a6f0-4c1e-7a2b-9d3e-000000000003',
    tenantId: SAMPLE_TENANT_ID,
    status: 'STATUS_DISABLE',
    server: {
      name: 'staging',
      resourceId: 'rsrc_0199a6f0-4c1e-7a2b-9d3e-000000000003',
      resourceDomain: 'staging.example.com',
    },
  },
].map((r) => Resource.fromJSON(r));

// Extra generated resources so the list has several pages to page through
const generatedResources: Resource[] = Array.from({ length: 20 }, (_, i): Resource => {
  const n = i + 1;
  const id = `rsrc_0199a6f0-4c1e-7a2b-9d3e-1000000000${String(n).padStart(2, '0')}`;
  const day = String(1 + (n % 28)).padStart(2, '0');
  return {
    metadata: {
      createdAt: new Date(`2026-08-${day}T00:00:00Z`),
      updatedAt: new Date(`2026-09-${day}T00:00:00Z`),
    },
    id,
    tenantId: SAMPLE_TENANT_ID,
    server: Server.fromPartial({
      name: `site${n}`,
      resourceId: id,
      resourceDomain: `site${n}.example.com`,
    }),
    status: n % 4 === 0 ? Status.STATUS_DISABLE : Status.STATUS_ACTIVE,
  };
});

let sampleStore: Resource[] = [...SAMPLE_RESOURCES, ...generatedResources];

// Written as proto JSON (the wire format), parsed with Tenant.fromJSON
export const SAMPLE_TENANTS: Tenant[] = [
  {
    metadata: { createdAt: '2026-08-20T03:00:00Z', updatedAt: '2026-09-30T07:20:00Z' },
    id: SAMPLE_TENANT_ID,
    name: 'Sentinez',
    description: 'Console sample tenant',
    plan: 'PLAN_PRO',
  },
  {
    metadata: { createdAt: '2026-09-15T06:10:00Z', updatedAt: '2026-09-15T06:10:00Z' },
    id: 'tnt_0199a6f0-4c1e-7a2b-9d3e-200000000002',
    name: 'Playground',
    description: '',
    plan: 'PLAN_FREE',
  },
].map((t) => Tenant.fromJSON(t));

let sampleTenantStore: Tenant[] = [...SAMPLE_TENANTS];

// Roles by position in the sample tenant: one admin, a few leaders (one of
// them also an admin), the rest members
const SAMPLE_MEMBER_ROLES = [
  ['ROLE_ADMIN'],
  ['ROLE_LEADER'],
  ['ROLE_MEMBER'],
  ['ROLE_ADMIN', 'ROLE_LEADER'],
  ['ROLE_LEADER'],
  ['ROLE_MEMBER'],
];

const sampleMember = (n: number, userId: string, tenantId: string, roles: string[]) => {
  const createdAt = new Date(Date.UTC(2026, 8, 2 + n, 8)).toISOString();
  return Member.fromJSON({
    metadata: { createdAt, updatedAt: createdAt },
    id: `mem_0199a6f0-4c1e-7a2b-9d3e-3000000000${String(n).padStart(2, '0')}`,
    userId,
    tenantId,
    roles,
  });
};

// The sample tenant has 2 pages of members; the last 3 users belong to no
// tenant so they can be invited, and Playground shares a few with it
const tenantUsers = ALL_SAMPLE_USERS.slice(0, -3);
const PLAYGROUND_TENANT_ID = 'tnt_0199a6f0-4c1e-7a2b-9d3e-200000000002';

export const SAMPLE_MEMBERS: Member[] = [
  ...tenantUsers.map((user, i) =>
    sampleMember(
      i + 1,
      user.id,
      SAMPLE_TENANT_ID,
      SAMPLE_MEMBER_ROLES[i < 2 ? i : 2 + (i % 4)] ?? ['ROLE_MEMBER'],
    ),
  ),
  ...tenantUsers
    .slice(1, 4)
    .map((user, i) =>
      sampleMember(tenantUsers.length + i + 1, user.id, PLAYGROUND_TENANT_ID, [
        i === 0 ? 'ROLE_ADMIN' : 'ROLE_MEMBER',
      ]),
    ),
];

let sampleMemberStore: Member[] = [...SAMPLE_MEMBERS];

/** What the backend answers when the user is already in the tenant */
function alreadyExists(message: string): AxiosError {
  const response = { status: 409, statusText: 'Conflict', data: { message } };
  return new AxiosError(message, AxiosError.ERR_BAD_REQUEST, undefined, undefined, {
    ...response,
    headers: {},
    config: { headers: new AxiosHeaders() },
  });
}

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
      (!req.resourceDomain || r.server?.resourceDomain === req.resourceDomain) &&
      (!req.resourceName || r.server?.name === req.resourceName) &&
      (!req.status || r.status === req.status),
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
      resource: sampleStore.find((r) => r.server?.resourceDomain === req.resourceDomain),
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
        // repeated fields as ?tenantIds=a&tenantIds=b
        paramsSerializer: { indexes: null },
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
    const id = `rsrc_${crypto.randomUUID()}`;
    const created: Resource = {
      ...req,
      id,
      tenantId: SAMPLE_TENANT_ID,
      // the API stores the server under the resource id
      server: req.server && { ...req.server, resourceId: id },
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
      r.id === req.id
        ? {
            ...r,
            ...req,
            // An update without a server keeps the stored one
            server: req.server ?? r.server,
            metadata: { ...r.metadata, updatedAt: new Date() },
          }
        : r,
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

// ─── Members ─────────────────────────────────────────────────────────────────

// GET /tenant/member?id=
export async function getMember(
  req: GetMemberRequest,
  options?: ApiOptions,
): Promise<GetMemberResponse> {
  return withFallback<GetMemberResponse>('getMember', () => ({
    member: sampleMemberStore.find((m) => m.id === req.id),
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/tenant/member`, {
      params: toQuery(GetMemberRequest.toJSON(req)),
      signal: options?.signal,
    });
    return GetMemberResponse.fromJSON(resp.data ?? {});
  });
}

// GET /tenant/members
export async function listMembers(
  req: Partial<ListMemberRequest> = {},
  options?: ApiOptions,
): Promise<ListMemberResponse> {
  return withFallback<ListMemberResponse>('listMembers', () => {
    const members = sampleMemberStore.filter(
      (m) =>
        (!req.tenantIds?.length || req.tenantIds.includes(m.tenantId)) &&
        (!req.userId || m.userId === req.userId),
    );
    return { total: members.length, members: paginate(members, req.page) };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/tenant/members`, {
      params: toQuery(ListMemberRequest.toJSON(ListMemberRequest.fromPartial(req))),
      // repeated fields as ?tenantId=a&tenantId=b
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListMemberResponse.fromJSON(resp.data ?? {});
  });
}

// POST /tenant/member
export async function createMember(
  req: CreateMemberRequest,
  options?: ApiOptions,
): Promise<CreateMemberResponse> {
  return withFallback<CreateMemberResponse>('createMember', () => {
    // CreateMember: a user can be a member of a tenant only once
    if (sampleMemberStore.some((m) => m.tenantId === req.tenantId && m.userId === req.userId)) {
      throw alreadyExists('member already exists');
    }
    const now = new Date();
    const created = Member.fromPartial({
      ...req,
      id: `mem_${crypto.randomUUID()}`,
      metadata: { createdAt: now, updatedAt: now },
    });
    sampleMemberStore = [...sampleMemberStore, created];
    return { member: created };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/tenant/member`,
      CreateMemberRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateMemberResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /tenant/member
export async function updateMember(
  req: UpdateMemberRequest,
  options?: ApiOptions,
): Promise<UpdateMemberResponse> {
  return withFallback<UpdateMemberResponse>('updateMember', () => {
    sampleMemberStore = sampleMemberStore.map((m) =>
      m.id === req.id
        ? { ...m, roles: [...req.roles], metadata: { ...m.metadata, updatedAt: new Date() } }
        : m,
    );
    return {};
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/tenant/member`,
      UpdateMemberRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateMemberResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /tenant/member?id=
export async function deleteMember(
  req: DeleteMemberRequest,
  options?: ApiOptions,
): Promise<DeleteMemberResponse> {
  return withFallback<DeleteMemberResponse>('deleteMember', () => {
    sampleMemberStore = sampleMemberStore.filter((m) => m.id !== req.id);
    return {};
  })(async () => {
    const resp = await axios.delete(`${API_BASE_PATH}/tenant/member`, {
      params: toQuery(DeleteMemberRequest.toJSON(req)),
      signal: options?.signal,
    });
    return DeleteMemberResponse.fromJSON(resp.data ?? {});
  });
}

// ─── Tenants ─────────────────────────────────────────────────────────────────

// GET /tenant/tenant?id=
export async function getTenant(
  req: GetTenantRequest,
  options?: ApiOptions,
): Promise<GetTenantResponse> {
  return withFallback<GetTenantResponse>('getTenant', () => ({
    tenant: sampleTenantStore.find((t) => t.id === req.id),
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/tenant/tenant`, {
      params: toQuery(GetTenantRequest.toJSON(req)),
      signal: options?.signal,
    });
    return GetTenantResponse.fromJSON(resp.data ?? {});
  });
}

// GET /tenant/tenants
export async function listTenants(
  req: Partial<ListTenantRequest> = {},
  options?: ApiOptions,
): Promise<ListTenantResponse> {
  return withFallback<ListTenantResponse>('listTenants', () => {
    const name = req.name?.toLowerCase();
    const tenants = sampleTenantStore.filter(
      (t) => (!req.plan || t.plan === req.plan) && (!name || t.name.toLowerCase().includes(name)),
    );
    return { total: tenants.length, tenants: paginate(tenants, req.page) };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/tenant/tenants`, {
      params: toQuery(ListTenantRequest.toJSON(ListTenantRequest.fromPartial(req))),
      signal: options?.signal,
    });
    return ListTenantResponse.fromJSON(resp.data ?? {});
  });
}

// POST /tenant/tenant
export async function createTenant(
  req: CreateTenantRequest,
  options?: ApiOptions,
): Promise<CreateTenantResponse> {
  return withFallback<CreateTenantResponse>('createTenant', () => {
    const now = new Date();
    const created = Tenant.fromPartial({
      ...req,
      id: `tnt_${crypto.randomUUID()}`,
      metadata: { createdAt: now, updatedAt: now },
    });
    sampleTenantStore = [...sampleTenantStore, created];
    return { tenant: created };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/tenant/tenant`,
      CreateTenantRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateTenantResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /tenant/tenant
export async function updateTenant(
  req: UpdateTenantRequest,
  options?: ApiOptions,
): Promise<UpdateTenantResponse> {
  return withFallback<UpdateTenantResponse>('updateTenant', () => {
    sampleTenantStore = sampleTenantStore.map((t) =>
      t.id === req.id
        ? {
            ...t,
            plan: req.plan,
            // empty name / description keep the stored ones
            name: req.name || t.name,
            description: req.description || t.description,
            metadata: { ...t.metadata, updatedAt: new Date() },
          }
        : t,
    );
    return {};
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/tenant/tenant`,
      UpdateTenantRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateTenantResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /tenant/tenant?id=
export async function deleteTenant(
  req: DeleteTenantRequest,
  options?: ApiOptions,
): Promise<DeleteTenantResponse> {
  return withFallback<DeleteTenantResponse>('deleteTenant', () => {
    sampleTenantStore = sampleTenantStore.filter((t) => t.id !== req.id);
    return {};
  })(async () => {
    const resp = await axios.delete(`${API_BASE_PATH}/tenant/tenant`, {
      params: toQuery(DeleteTenantRequest.toJSON(req)),
      signal: options?.signal,
    });
    return DeleteTenantResponse.fromJSON(resp.data ?? {});
  });
}

// ─── CA certificates ─────────────────────────────────────────────────────────

// A certificate is immutable: there is no update, delete it and create it again.

const SAMPLE_CA_RESOURCE_ID = SAMPLE_RESOURCES[0]?.id ?? '';

const sampleCa = (n: number, partial: Partial<CaCertificate>): CaCertificate =>
  CaCertificate.fromPartial({
    id: `cacrt_console-sample-${n}`,
    owner: { resourceId: SAMPLE_CA_RESOURCE_ID, tenantId: SAMPLE_TENANT_ID },
    status: Status.STATUS_ACTIVE,
    usage: CaCertificateUsage.CA_CERTIFICATE_USAGE_UPSTREAM,
    pem: '-----BEGIN CERTIFICATE-----\nMIIBsample\n-----END CERTIFICATE-----\n',
    serialNumber: `0${n}a1`,
    fingerprintSha256: `${n}`.repeat(64),
    notBefore: new Date('2026-01-01T00:00:00Z'),
    notAfter: new Date('2036-01-01T00:00:00Z'),
    ...partial,
  });

export const SAMPLE_CA_CERTIFICATES: CaCertificate[] = [
  sampleCa(1, {
    name: 'Acme Root CA',
    description: 'Root of the Acme private PKI',
    provider: 'Acme Trust',
    subject: 'CN=Acme Root CA,O=Acme Trust',
    issuer: 'CN=Acme Root CA,O=Acme Trust',
    selfSigned: true,
    hasPrivateKey: true,
  }),
  sampleCa(2, {
    name: 'Acme Issuing CA',
    description: 'Signs the origin servers',
    provider: 'Acme Trust',
    subject: 'CN=Acme Issuing CA',
    issuer: 'CN=Acme Root CA,O=Acme Trust',
    usage: CaCertificateUsage.CA_CERTIFICATE_USAGE_CLIENT,
  }),
  sampleCa(3, {
    name: 'Legacy CA',
    description: 'Expired, kept to show the badge',
    provider: 'Legacy PKI',
    subject: 'CN=Legacy CA,O=Legacy PKI',
    issuer: 'CN=Legacy CA,O=Legacy PKI',
    selfSigned: true,
    notBefore: new Date('2020-01-01T00:00:00Z'),
    notAfter: new Date('2025-01-01T00:00:00Z'),
  }),
];

let sampleCaStore: CaCertificate[] = [...SAMPLE_CA_CERTIFICATES];

const CA_PEM_BLOCK = /-----BEGIN CERTIFICATE-----[\s\S]*?-----END CERTIFICATE-----/g;

/** One sample record per certificate of the bundle; the sample does not parse the PEM */
function sampleCaOf(req: CreateCaCertificateRequest): CaCertificate[] {
  const blocks = req.pem.match(CA_PEM_BLOCK) ?? [req.pem];
  const now = new Date();
  const nextYear = new Date(now.getTime() + 365 * 24 * 3600 * 1000);

  return blocks.map((pem, i) => {
    const n = i + 1;
    const id = crypto.randomUUID();
    return CaCertificate.fromPartial({
      id: `cacrt_${id}`,
      owner: { resourceId: req.resourceId, tenantId: SAMPLE_TENANT_ID },
      name: !req.name ? `Sample CA ${n}` : blocks.length > 1 ? `${req.name} #${n}` : req.name,
      description: req.description,
      usage: req.usage,
      status: Status.STATUS_ACTIVE,
      pem: `${pem.trim()}\n`,
      subject: `CN=Sample CA ${n}`,
      issuer: `CN=Sample CA ${n}`,
      provider: 'Sample CA',
      serialNumber: id.slice(0, 8),
      fingerprintSha256: id.replaceAll('-', '').repeat(2),
      notBefore: now,
      notAfter: nextYear,
      selfSigned: true,
      hasPrivateKey: req.privateKey !== '',
    });
  });
}

function filterSampleCa(req: Partial<ListCaCertificateRequest>): CaCertificate[] {
  return sampleCaStore.filter(
    (c) =>
      (!req.resourceIds?.length || req.resourceIds.includes(c.owner?.resourceId ?? '')) &&
      (!req.usage || c.usage === req.usage) &&
      (!req.fingerprintSha256 || c.fingerprintSha256 === req.fingerprintSha256) &&
      (!req.provider || c.provider === req.provider),
  );
}

/** The statuses with which the server turns a CA call down for a reason of its own */
const CA_REFUSALS = new Set([400, 403, 404, 409, 422]);

/**
 * Like withFallback, but the sample answers only when the backend is out of
 * reach: a refusal of the server (an expired certificate, one that exists
 * already...) is thrown, or the page would report a success that never was.
 */
function caFallback<T>(label: string, fallback: () => T) {
  return async (call: () => Promise<T>): Promise<T> => {
    try {
      return await call();
    } catch (err: unknown) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined;
      const refused = status !== undefined && CA_REFUSALS.has(status);
      if (!USE_SAMPLE || axios.isCancel(err) || refused) throw err;
      console.warn(`[tenant] ${label} failed, using sample response:`, (err as Error)?.message);
      return fallback();
    }
  };
}

// GET /tenant/cacertificates
export async function listCaCertificates(
  req: Partial<ListCaCertificateRequest> = {},
  options?: ApiOptions,
): Promise<ListCaCertificateResponse> {
  return caFallback<ListCaCertificateResponse>('listCaCertificates', () => {
    const certs = filterSampleCa(req);
    return { total: certs.length, caCertificates: paginate(certs, req.page) };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/tenant/cacertificates`, {
      params: toQuery(ListCaCertificateRequest.toJSON(ListCaCertificateRequest.fromPartial(req))),
      // repeated fields as ?resourceIds=a&resourceIds=b
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListCaCertificateResponse.fromJSON(resp.data ?? {});
  });
}

// POST /tenant/cacertificate
export async function createCaCertificate(
  req: CreateCaCertificateRequest,
  options?: ApiOptions,
): Promise<CreateCaCertificateResponse> {
  return caFallback<CreateCaCertificateResponse>('createCaCertificate', () => {
    const created = sampleCaOf(req);
    sampleCaStore = [...sampleCaStore, ...created];
    return { caCertificates: created };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/tenant/cacertificate`,
      CreateCaCertificateRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateCaCertificateResponse.fromJSON(resp.data ?? {});
  });
}

// GET /tenant/cacertificate?id=
export async function getCaCertificate(
  req: GetCaCertificateRequest,
  options?: ApiOptions,
): Promise<GetCaCertificateResponse> {
  return caFallback<GetCaCertificateResponse>('getCaCertificate', () => ({
    caCertificate: sampleCaStore.find((c) => c.id === req.id),
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/tenant/cacertificate`, {
      params: toQuery(GetCaCertificateRequest.toJSON(req)),
      signal: options?.signal,
    });
    return GetCaCertificateResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /tenant/cacertificate?id=
export async function deleteCaCertificate(
  req: DeleteCaCertificateRequest,
  options?: ApiOptions,
): Promise<DeleteCaCertificateResponse> {
  return caFallback<DeleteCaCertificateResponse>('deleteCaCertificate', () => {
    sampleCaStore = sampleCaStore.filter((c) => c.id !== req.id);
    return {};
  })(async () => {
    const resp = await axios.delete(`${API_BASE_PATH}/tenant/cacertificate`, {
      params: toQuery(DeleteCaCertificateRequest.toJSON(req)),
      signal: options?.signal,
    });
    return DeleteCaCertificateResponse.fromJSON(resp.data ?? {});
  });
}
