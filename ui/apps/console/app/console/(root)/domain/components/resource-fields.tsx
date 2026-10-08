'use client';

import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import { PlusIcon, Trash2Icon } from 'lucide-react';
import type {
  CreateResourceRequest,
  TenantPlan,
  TenantResource,
  TenantStatus,
} from '@/lib/api/tenant';

type Protocol = 'PROXY_PROTOCOL_HTTP' | 'PROXY_PROTOCOL_HTTPS';

export interface OriginRow {
  server: string;
  protocol: Protocol;
}

export interface ResourceFormValue {
  resourceName: string;
  resourceDomain: string;
  plan: TenantPlan;
  status: TenantStatus;
  /** Path the origins serve, e.g. "/" */
  location: string;
  /** Empty means the backend applies its default edge setting */
  origins: OriginRow[];
}

interface Option<T extends string> {
  label: string;
  value: T;
  description?: string;
}

const PLAN_OPTIONS: Option<TenantPlan>[] = [
  { label: 'Free', value: 'PLAN_FREE', description: 'For testing and personal projects.' },
  { label: 'Standard', value: 'PLAN_STANDARD', description: 'For production sites.' },
  { label: 'Pro', value: 'PLAN_PRO', description: 'For high-traffic sites.' },
];

const STATUS_OPTIONS: Option<TenantStatus>[] = [
  {
    label: 'Active',
    value: 'STATUS_ACTIVE',
    description: 'The edge serves traffic for this domain.',
  },
  {
    label: 'Disable',
    value: 'STATUS_DISABLE',
    description: 'The resource is kept but the edge does not serve it.',
  },
];

const PROTOCOL_OPTIONS: Option<Protocol>[] = [
  { label: 'HTTPS', value: 'PROXY_PROTOCOL_HTTPS' },
  { label: 'HTTP', value: 'PROXY_PROTOCOL_HTTP' },
];

// Patterns from tenant.proto / setting.proto (buf.validate)
const DOMAIN_RE = /^[a-zA-Z][a-zA-Z0-9._]{2,29}$/;
const LOCATION_RE = /^\/.*$/;
const UPSTREAM_RE =
  /^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})|(([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}))$/;

export function createEmptyResourceForm(): ResourceFormValue {
  return {
    resourceName: '',
    resourceDomain: '',
    plan: 'PLAN_FREE',
    status: 'STATUS_ACTIVE',
    location: '/',
    origins: [],
  };
}

/** TenantResource (API) -> form value */
export function resourceFormOf(r: TenantResource): ResourceFormValue {
  const loc = r.resourceSetting?.server?.locations?.[0];
  return {
    resourceName: r.resourceName ?? '',
    resourceDomain: r.resourceDomain ?? '',
    plan: r.plan ?? 'PLAN_FREE',
    status: r.status ?? 'STATUS_ACTIVE',
    location: loc?.location ?? '/',
    origins: (loc?.proxyPass ?? []).map((u: any) => ({
      server: u.server ?? '',
      protocol:
        u.protocol === 'PROXY_PROTOCOL_HTTP' ? 'PROXY_PROTOCOL_HTTP' : 'PROXY_PROTOCOL_HTTPS',
    })),
  };
}

/** Form value -> CreateResourceRequest (API) */
export function resourceOf(f: ResourceFormValue): CreateResourceRequest {
  const name = f.resourceName.trim();
  const origins = f.origins.filter((o) => o.server.trim());
  return {
    resourceName: name,
    resourceDomain: f.resourceDomain.trim(),
    plan: f.plan,
    status: f.status,
    resourceSetting:
      origins.length === 0
        ? undefined
        : {
            server: {
              name,
              locations: [
                {
                  location: f.location.trim(),
                  proxyRewrite: '/',
                  proxyPass: origins.map((o) => ({
                    server: o.server.trim(),
                    protocol: o.protocol,
                  })),
                  balanceStrategy: 'BALANCE_STRATEGY_ROUND_ROBIN',
                  proxySetHeaders: {},
                },
              ],
            },
          },
  };
}

/** Returns an error message, or null when the form is valid */
export function validateResourceForm(f: ResourceFormValue): string | null {
  if (!DOMAIN_RE.test(f.resourceName.trim())) {
    return 'Name must start with a letter and be 3-30 characters of letters, digits, "." or "_"';
  }
  if (!DOMAIN_RE.test(f.resourceDomain.trim())) {
    return 'Domain must start with a letter and be 3-30 characters of letters, digits, "." or "_"';
  }
  const origins = f.origins.filter((o) => o.server.trim());
  if (origins.length === 0) return null;
  if (!LOCATION_RE.test(f.location.trim())) {
    return 'Location must start with "/"';
  }
  const bad = origins.find((o) => !UPSTREAM_RE.test(o.server.trim()));
  if (bad) {
    return `Origin "${bad.server}" must be an IPv4 address or a hostname`;
  }
  return null;
}

function OptionSelect<T extends string>({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: T;
  options: Option<T>[];
  onChange: (v: T) => void;
}) {
  const selected = options.find((o) => o.value === value);
  return (
    <div className="grid gap-2 content-start">
      <Label>{label}</Label>
      <Select value={value} onValueChange={(v) => onChange(v as T)}>
        <SelectTrigger className="w-full">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {options.map((o) => (
            <SelectItem key={o.value} value={o.value}>
              {o.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {selected?.description && (
        <p className="text-sm text-muted-foreground">{selected.description}</p>
      )}
    </div>
  );
}

function OriginRows({
  rows,
  onChange,
  idPrefix,
}: {
  rows: OriginRow[];
  onChange: (rows: OriginRow[]) => void;
  idPrefix: string;
}) {
  const patchRow = (i: number, patch: Partial<OriginRow>) =>
    onChange(rows.map((r, j) => (j === i ? { ...r, ...patch } : r)));

  return (
    <div className="flex flex-col gap-2">
      {rows.map((row, i) => (
        <div key={i} className="flex gap-2">
          <Select
            value={row.protocol}
            onValueChange={(v) => patchRow(i, { protocol: v as Protocol })}
          >
            <SelectTrigger className="w-28 shrink-0" aria-label="Protocol">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {PROTOCOL_OPTIONS.map((o) => (
                <SelectItem key={o.value} value={o.value}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Input
            id={`${idPrefix}-origin-${i}`}
            aria-label="Origin server"
            placeholder="10.0.0.10 or origin.example.com"
            spellCheck={false}
            autoComplete="off"
            value={row.server}
            onChange={(e) => patchRow(i, { server: e.target.value })}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label="Remove origin"
            onClick={() => onChange(rows.filter((_, j) => j !== i))}
          >
            <Trash2Icon className="w-4 h-4" />
          </Button>
        </div>
      ))}
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="w-fit"
        onClick={() => onChange([...rows, { server: '', protocol: 'PROXY_PROTOCOL_HTTPS' }])}
      >
        <PlusIcon className="w-4 h-4" />
        Add origin
      </Button>
    </div>
  );
}

interface Props {
  value: ResourceFormValue;
  onChange: (patch: Partial<ResourceFormValue>) => void;
  idPrefix?: string;
}

/** Form fields matching v1CreateResourceRequest in tenant.swagger.json */
export function ResourceFields({ value, onChange, idPrefix = 'resource' }: Props) {
  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div className="grid gap-2 content-start">
          <Label htmlFor={`${idPrefix}-name`}>Name</Label>
          <Input
            id={`${idPrefix}-name`}
            placeholder="my_site"
            spellCheck={false}
            autoComplete="off"
            value={value.resourceName}
            onChange={(e) => onChange({ resourceName: e.target.value })}
          />
        </div>
        <div className="grid gap-2 content-start">
          <Label htmlFor={`${idPrefix}-domain`}>Domain</Label>
          <Input
            id={`${idPrefix}-domain`}
            placeholder="example.com"
            spellCheck={false}
            autoComplete="off"
            value={value.resourceDomain}
            onChange={(e) => onChange({ resourceDomain: e.target.value })}
          />
          <p className="text-sm text-muted-foreground">The host name the edge serves.</p>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <OptionSelect
          label="Plan"
          value={value.plan}
          options={PLAN_OPTIONS}
          onChange={(plan) => onChange({ plan })}
        />
        <OptionSelect
          label="Status"
          value={value.status}
          options={STATUS_OPTIONS}
          onChange={(status) => onChange({ status })}
        />
      </div>

      <div className="grid gap-2 mt-2">
        <Label>Origin</Label>
        <p className="text-sm text-muted-foreground">
          Upstream servers the edge proxies to (round robin). Leave empty to start from the default
          edge setting.
        </p>
        <div className="grid gap-2 sm:w-1/2">
          <Label htmlFor={`${idPrefix}-location`}>Location</Label>
          <Input
            id={`${idPrefix}-location`}
            placeholder="/"
            spellCheck={false}
            autoComplete="off"
            value={value.location}
            onChange={(e) => onChange({ location: e.target.value })}
          />
        </div>
        <OriginRows
          rows={value.origins}
          onChange={(origins) => onChange({ origins })}
          idPrefix={idPrefix}
        />
      </div>
    </div>
  );
}
