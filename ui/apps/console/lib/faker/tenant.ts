import {
  CaCertificate,
  CaCertificateUsage,
  Member,
  Resource,
  Tenant,
} from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import {
  CreateCaCertificateRequest,
  ListCaCertificateRequest,
  ListMemberRequest,
  ListResourceRequest,
  ListTenantRequest,
} from '@sentinez/proto/sentinez/apps/tenant/v1/tenant';
import { Resource as EdgeResource, TlsMode } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import { AxiosError, AxiosHeaders } from 'axios';

import { DAY_MILLIS } from '@/lib/api/analytic/common';
import { sampleRandom } from '@/lib/faker/random';
import { SampleStore, byResources } from '@/lib/faker/store';
import { ALL_SAMPLE_USERS } from '@/lib/faker/iam';

// Tenant that owns every sample resource
export const SAMPLE_TENANT_ID = 'tnt_console-sample';

/** The name, domain, locations and TLS mode of a resource from the edge resource of a request */
export function tenantFieldsOf(
  srv?: EdgeResource,
): Pick<Resource, 'name' | 'domain' | 'locations' | 'tlsMode'> {
  return {
    name: srv?.name ?? '',
    domain: srv?.domain ?? '',
    locations: srv?.locations ?? [],
    tlsMode: srv?.tlsMode ?? TlsMode.TLS_MODE_UNSPECIFIED,
  };
}

const upstream = (server: string, protocol = 'PROXY_PROTOCOL_HTTPS') => ({ server, protocol });

// Written as proto JSON (the wire format), parsed with Resource.fromJSON
export const SAMPLE_RESOURCES: Resource[] = [
  {
    metadata: { createdAt: '2026-09-01T08:00:00Z', updatedAt: '2026-10-05T10:30:00Z' },
    id: 'rsrc_0199a6f0-4c1e-7a2b-9d3e-000000000001',
    tenantId: SAMPLE_TENANT_ID,
    status: 'STATUS_ACTIVE',
    name: 'badcheese',
    domain: 'badcheese.s6z.io.vn',
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
  {
    metadata: { createdAt: '2026-09-12T02:15:00Z', updatedAt: '2026-09-28T14:00:00Z' },
    id: 'rsrc_0199a6f0-4c1e-7a2b-9d3e-000000000002',
    tenantId: SAMPLE_TENANT_ID,
    status: 'STATUS_ACTIVE',
    name: 'shop',
    domain: 'shop.example.com',
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
  {
    metadata: { createdAt: '2026-10-01T09:45:00Z', updatedAt: '2026-10-01T09:45:00Z' },
    id: 'rsrc_0199a6f0-4c1e-7a2b-9d3e-000000000003',
    tenantId: SAMPLE_TENANT_ID,
    status: 'STATUS_DISABLE',
    name: 'staging',
    domain: 'staging.example.com',
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
    name: `site${n}`,
    domain: `site${n}.example.com`,
    locations: [],
    status: n % 4 === 0 ? Status.STATUS_DISABLE : Status.STATUS_ACTIVE,
    tlsMode: TlsMode.TLS_MODE_UNSPECIFIED,
  };
});

export const resourceStore = new SampleStore([...SAMPLE_RESOURCES, ...generatedResources]);

// Written as proto JSON (the wire format), parsed with Tenant.fromJSON
const SAMPLE_TENANTS: Tenant[] = [
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

export const tenantStore = new SampleStore(SAMPLE_TENANTS);

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

const SAMPLE_MEMBERS: Member[] = [
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

export const memberStore = new SampleStore(SAMPLE_MEMBERS);

/** What the backend answers when the user is already in the tenant */
export function alreadyExists(message: string): AxiosError {
  const response = { status: 409, statusText: 'Conflict', data: { message } };
  return new AxiosError(message, AxiosError.ERR_BAD_REQUEST, undefined, undefined, {
    ...response,
    headers: {},
    config: { headers: new AxiosHeaders() },
  });
}

export function filterSample(req: Partial<ListResourceRequest>): Resource[] {
  return resourceStore
    .all()
    .filter(
      (r) =>
        (!req.resourceDomain || r.domain === req.resourceDomain) &&
        (!req.resourceName || r.name === req.resourceName) &&
        (!req.status || r.status === req.status),
    );
}

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

const SAMPLE_CA_CERTIFICATES: CaCertificate[] = [
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

// A date days from now (before it when negative): the samples keep their
// expiry badges (expired, expiring soon) whenever the console runs
const caDaysFromNow = (days: number) => new Date(Date.now() + days * DAY_MILLIS);

// The more samples of the first resource: one of each usage the base ones
// lack, one expiring soon
const SAMPLE_CA_EXTRAS: CaCertificate[] = [
  sampleCa(4, {
    name: 'shop.example.com',
    description: 'The certificate the edge serves, renewed soon',
    provider: 'Acme Trust',
    subject: 'CN=shop.example.com',
    issuer: 'CN=Acme Issuing CA',
    usage: CaCertificateUsage.CA_CERTIFICATE_USAGE_SERVER,
    hasPrivateKey: true,
    notBefore: caDaysFromNow(-353),
    notAfter: caDaysFromNow(12),
  }),
  sampleCa(5, {
    name: 'Origin mTLS client',
    description: 'Presented by the edge to the origin servers',
    provider: 'Acme Trust',
    subject: 'CN=edge.sentinez,O=Acme Trust',
    issuer: 'CN=Acme Issuing CA',
    usage: CaCertificateUsage.CA_CERTIFICATE_USAGE_UPSTREAM_CLIENT,
    hasPrivateKey: true,
    notBefore: caDaysFromNow(-165),
    notAfter: caDaysFromNow(200),
  }),
  sampleCa(6, {
    name: 'Partner Client CA',
    description: 'Signs the client certificates of the partner apps',
    provider: 'Partner PKI',
    subject: 'CN=Partner Client CA,O=Partner PKI',
    issuer: 'CN=Partner Client CA,O=Partner PKI',
    usage: CaCertificateUsage.CA_CERTIFICATE_USAGE_CLIENT,
    selfSigned: true,
  }),
];

const SAMPLE_CA_USAGES = [
  CaCertificateUsage.CA_CERTIFICATE_USAGE_UPSTREAM,
  CaCertificateUsage.CA_CERTIFICATE_USAGE_CLIENT,
  CaCertificateUsage.CA_CERTIFICATE_USAGE_SERVER,
  CaCertificateUsage.CA_CERTIFICATE_USAGE_UPSTREAM_CLIENT,
];
const SAMPLE_CA_PROVIDERS = ['Acme Trust', 'Contoso PKI', 'Internal PKI'];

// A sample hex string of length, the same for the same seed
function caHexOf(seed: string, length: number): string {
  const random = sampleRandom(seed);
  return Array.from({ length }, () => Math.floor(random() * 16).toString(16)).join('');
}

// The days the k-th generated sample has left: some expired, some expiring
// soon, the others valid for months or years
function caDaysLeftOf(k: number): number {
  if (k % 7 === 0) return -(10 + (k % 50));
  if (k % 5 === 0) return 3 + (k % 25);
  return 60 + ((k * 37) % 900);
}

// The k-th generated sample (from 1), of the resource of domain: numbered
// after the base and extra ones
function generatedCa(k: number, resource: Resource): CaCertificate {
  const n = SAMPLE_CA_CERTIFICATES.length + SAMPLE_CA_EXTRAS.length + k;
  const usage = SAMPLE_CA_USAGES[k % SAMPLE_CA_USAGES.length]!;
  const provider = SAMPLE_CA_PROVIDERS[k % SAMPLE_CA_PROVIDERS.length]!;
  // the server and upstream client ones are leaf certificates with their key
  const leaf =
    usage === CaCertificateUsage.CA_CERTIFICATE_USAGE_SERVER ||
    usage === CaCertificateUsage.CA_CERTIFICATE_USAGE_UPSTREAM_CLIENT;
  const name =
    usage === CaCertificateUsage.CA_CERTIFICATE_USAGE_SERVER
      ? `*.${resource.domain}`
      : leaf
        ? `edge-client-${k}.${resource.domain}`
        : `${provider} Issuing CA ${k}`;
  const selfSigned = !leaf && k % 3 === 0;
  const daysLeft = caDaysLeftOf(k);
  return sampleCa(n, {
    name,
    description: leaf ? `Generated for ${resource.domain}` : `Generated CA of ${provider}`,
    owner: { resourceId: resource.id, tenantId: SAMPLE_TENANT_ID },
    provider,
    usage,
    subject: `CN=${name},O=${provider}`,
    issuer: selfSigned ? `CN=${name},O=${provider}` : `CN=${provider} Root CA,O=${provider}`,
    selfSigned,
    hasPrivateKey: leaf,
    serialNumber: caHexOf(`${n}|serial`, 16),
    fingerprintSha256: caHexOf(`${n}|sha256`, 64),
    notBefore: caDaysFromNow(daysLeft - (leaf ? 365 : 3650)),
    notAfter: caDaysFromNow(daysLeft),
  });
}

// The generated samples: every sample resource has some, the first one
// more than a page (10) with the base and extra ones
const SAMPLE_CA_GENERATED: CaCertificate[] = [...SAMPLE_RESOURCES, ...generatedResources]
  .flatMap((resource, i) => Array.from({ length: i === 0 ? 9 : 1 + (i % 4) }, () => resource))
  .map((resource, k) => generatedCa(k + 1, resource));

export const caStore = new SampleStore([
  ...SAMPLE_CA_CERTIFICATES,
  ...SAMPLE_CA_EXTRAS,
  ...SAMPLE_CA_GENERATED,
]);

const CA_PEM_BLOCK = /-----BEGIN CERTIFICATE-----[\s\S]*?-----END CERTIFICATE-----/g;

/** One sample record per certificate of the bundle; the sample does not parse the PEM */
export function sampleCaOf(req: CreateCaCertificateRequest): CaCertificate[] {
  const blocks = req.pem.match(CA_PEM_BLOCK) ?? [req.pem];
  const now = new Date();
  const nextYear = new Date(now.getTime() + 365 * DAY_MILLIS);

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

export function filterSampleCa(req: Partial<ListCaCertificateRequest>): CaCertificate[] {
  return byResources(caStore.all(), req.resourceIds).filter(
    (c) =>
      (!req.usage || c.usage === req.usage) &&
      (!req.fingerprintSha256 || c.fingerprintSha256 === req.fingerprintSha256) &&
      (!req.provider || c.provider === req.provider),
  );
}

export function filterSampleMembers(req: Partial<ListMemberRequest>): Member[] {
  return memberStore
    .all()
    .filter(
      (m) =>
        (!req.tenantIds?.length || req.tenantIds.includes(m.tenantId)) &&
        (!req.userId || m.userId === req.userId),
    );
}

export function filterSampleTenants(req: Partial<ListTenantRequest>): Tenant[] {
  const name = req.name?.toLowerCase();
  return tenantStore
    .all()
    .filter(
      (t) => (!req.plan || t.plan === req.plan) && (!name || t.name.toLowerCase().includes(name)),
    );
}
