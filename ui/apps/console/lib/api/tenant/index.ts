import { Member, Resource, Tenant } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
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
import { Resource as EdgeResource } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { USE_SAMPLE, fallbackOf } from '@/lib/api/fallback';
import { applyMask } from '@/lib/api/mask';
import { paginate, toQuery } from '@/lib/api/pages';
import {
  SAMPLE_TENANT_ID,
  alreadyExists,
  caStore,
  filterSample,
  filterSampleCa,
  filterSampleMembers,
  filterSampleTenants,
  memberStore,
  resourceStore,
  sampleCaOf,
  tenantFieldsOf,
  tenantStore,
} from '@/lib/faker/tenant';
export interface ApiOptions {
  signal?: AbortSignal;
}

/** The edge resource of r: what the edge serves it as (EdgeResource of sznormalize) */
export function edgeResourceOf(r: Resource): EdgeResource {
  return EdgeResource.fromPartial({
    id: r.id,
    name: r.name,
    domain: r.domain,
    locations: r.locations,
    tlsMode: r.tlsMode,
  });
}

const withFallback = fallbackOf('tenant');

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

// GET /tenant/resource?id=
export async function getResource(req: Partial<GetResourceRequest>, options?: ApiOptions) {
  return orNull('fetching resource', () =>
    withFallback<GetResourceResponse>('getResource', () => ({
      resource: req.id ? resourceStore.find(req.id) : undefined,
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
      resource: resourceStore.all().find((r) => r.domain === req.resourceDomain),
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
      ...tenantFieldsOf(req.resource),
      id,
      tenantId: SAMPLE_TENANT_ID,
      status: req.status,
      metadata: { createdAt: now, updatedAt: now },
    };
    resourceStore.add(created);
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
    const r = resourceStore.find(req.id);
    if (r) {
      resourceStore.replace(req.id, {
        ...applyMask(r, req.resource, req.updateMask, Resource.fromPartial({})),
        id: r.id,
        tenantId: r.tenantId,
        metadata: { ...r.metadata, updatedAt: new Date() },
      });
    }
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
    resourceStore.remove(req.id);
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
    member: memberStore.find(req.id),
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
    const members = filterSampleMembers(req);
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
    if (memberStore.all().some((m) => m.tenantId === req.tenantId && m.userId === req.userId)) {
      throw alreadyExists('member already exists');
    }
    const now = new Date();
    const created = Member.fromPartial({
      ...req,
      id: `mem_${crypto.randomUUID()}`,
      metadata: { createdAt: now, updatedAt: now },
    });
    memberStore.add(created);
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
    const m = memberStore.find(req.id);
    if (m) {
      memberStore.replace(req.id, {
        ...m,
        roles: [...req.roles],
        metadata: { ...m.metadata, updatedAt: new Date() },
      });
    }
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
    memberStore.remove(req.id);
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
    tenant: tenantStore.find(req.id),
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
    const tenants = filterSampleTenants(req);
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
    tenantStore.add(created);
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
    const t = tenantStore.find(req.id);
    if (t) {
      tenantStore.replace(req.id, {
        ...applyMask(t, req.tenant, req.updateMask, Tenant.fromPartial({})),
        id: t.id,
        metadata: { ...t.metadata, updatedAt: new Date() },
      });
    }
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
    tenantStore.remove(req.id);
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
    caStore.add(...created);
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
    caCertificate: caStore.find(req.id),
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
    caStore.remove(req.id);
    return {};
  })(async () => {
    const resp = await axios.delete(`${API_BASE_PATH}/tenant/cacertificate`, {
      params: toQuery(DeleteCaCertificateRequest.toJSON(req)),
      signal: options?.signal,
    });
    return DeleteCaCertificateResponse.fromJSON(resp.data ?? {});
  });
}
