'use client';

import { PlusIcon, Trash2Icon } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { Textarea } from '@sentinez/ui/components/textarea';
import {
  BalanceStrategy,
  ProxyProtocol,
  type Location,
  type Server,
} from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { OptionSelect, OriginRows, type OriginRow } from '@/app/d/console/(root)/domain/components';
import {
  BALANCE_STRATEGIES,
  HEADER_NAME_RE,
  LOCATION_RE,
  NAME_RE,
  UPSTREAM_RE,
  balanceLabelKey,
  headerLinesOf,
  protocolOf,
} from '@/lib/type/edge';
import { FieldGrid } from '@/components/field-grid';

export interface LocationFormValue {
  location: string;
  /** Blank when the path is not rewritten: the edge then forwards it as is */
  proxyRewrite: string;
  /** Picks between several upstreams, not sent for a single one */
  balanceStrategy: BalanceStrategy;
  upstreams: OriginRow[];
  /** The request headers sent to the origin, one "Name: value" per line */
  headers: string;
}

export interface ServerFormValue {
  name: string;
  locations: LocationFormValue[];
}

type ValidationTranslator = ReturnType<typeof useTranslations<'Validation'>>;

/** A location proxying "/" to nothing yet */
function createEmptyLocation(): LocationFormValue {
  return {
    location: '/',
    proxyRewrite: '',
    balanceStrategy: BalanceStrategy.BALANCE_STRATEGY_ROUND_ROBIN,
    upstreams: [{ server: '', protocol: ProxyProtocol.PROXY_PROTOCOL_HTTPS }],
    headers: '',
  };
}

function locationFormOf(l: Location): LocationFormValue {
  return {
    location: l.location,
    proxyRewrite: l.proxyRewrite === l.location ? '' : l.proxyRewrite,
    balanceStrategy: l.balanceStrategy,
    upstreams: l.proxyPass.map((u) => ({ server: u.server, protocol: protocolOf(u.protocol) })),
    headers: Object.entries(l.proxySetHeaders)
      .map(([name, value]) => `${name}: ${value}`)
      .join('\n'),
  };
}

/** Server (API) -> form value */
export function serverFormOf(server?: Server): ServerFormValue {
  return {
    name: server?.name ?? '',
    locations: (server?.locations ?? []).map(locationFormOf),
  };
}

// The location as the edge takes it. Fields left at what they default to
// are filled the way the edge would: the rewrite is the path (the API
// requires one), the balancing is the default for a single upstream.
function locationOf(f: LocationFormValue): Location {
  const location = f.location.trim();
  const proxyPass = f.upstreams
    .filter((u) => u.server.trim())
    .map((u) => ({ server: u.server.trim(), protocol: u.protocol }));
  return {
    location,
    proxyRewrite: f.proxyRewrite.trim() || location,
    proxyPass,
    balanceStrategy:
      proxyPass.length > 1 ? f.balanceStrategy : BalanceStrategy.BALANCE_STRATEGY_UNSPECIFIED,
    proxySetHeaders: Object.fromEntries(headerLinesOf(f.headers).map((h) => [h.name, h.value])),
  };
}

/**
 * Form value -> Server (API). The ids and the domain are the stored
 * server's: they are not edited here.
 */
export function serverOf(f: ServerFormValue, stored?: Server): Server {
  return {
    name: f.name.trim(),
    resourceId: stored?.resourceId ?? '',
    resourceDomain: stored?.resourceDomain ?? '',
    locations: f.locations.map(locationOf),
  };
}

// The headers of a location as the edge takes them: "Name: value" lines
// with a token name, each name once (names are not case sensitive)
function validateHeaders(l: LocationFormValue, t: ValidationTranslator): string | null {
  const seen = new Set<string>();
  for (const h of headerLinesOf(l.headers)) {
    if (!h.line.includes(':') || !h.value) return t('headerLine', { line: h.line });
    if (!HEADER_NAME_RE.test(h.name)) return t('headerName', { name: h.name });
    if (seen.has(h.name.toLowerCase())) {
      return t('headerTwice', { name: h.name, location: l.location });
    }
    seen.add(h.name.toLowerCase());
  }
  return null;
}

function validateLocation(l: LocationFormValue, t: ValidationTranslator): string | null {
  if (!LOCATION_RE.test(l.location.trim())) return t('location');
  const rewrite = l.proxyRewrite.trim();
  if (rewrite && !LOCATION_RE.test(rewrite)) return t('rewrite');
  const upstreams = l.upstreams.filter((u) => u.server.trim());
  if (upstreams.length === 0) return t('locationOrigin', { location: l.location });
  const bad = upstreams.find((u) => !UPSTREAM_RE.test(u.server.trim()));
  if (bad) return t('origin', { server: bad.server });
  return validateHeaders(l, t);
}

/** Returns an error message, or null when the form is valid */
export function validateServerForm(f: ServerFormValue, t: ValidationTranslator): string | null {
  if (!NAME_RE.test(f.name.trim())) return t('resourceName');
  const seen = new Set<string>();
  for (const l of f.locations) {
    const error = validateLocation(l, t);
    if (error) return error;
    if (seen.has(l.location.trim())) return t('locationTwice', { location: l.location });
    seen.add(l.location.trim());
  }
  return null;
}

function LocationFields({
  index,
  value,
  onChange,
  onRemove,
}: {
  index: number;
  value: LocationFormValue;
  onChange: (patch: Partial<LocationFormValue>) => void;
  onRemove: () => void;
}) {
  const t = useTranslations('ResourceSetting');
  const idPrefix = `location-${index}`;
  const balances = BALANCE_STRATEGIES.map((value) => ({ value, label: t(balanceLabelKey(value)) }));
  // The strategy only picks between several upstreams
  const balanced = value.upstreams.filter((u) => u.server.trim()).length > 1;

  return (
    <fieldset className="flex flex-col gap-3 rounded-md border p-3 min-w-0">
      <legend className="sr-only">{t('locationN', { n: index + 1 })}</legend>

      <div className="flex items-end gap-2">
        <FieldGrid className="flex-1 gap-3">
          <div className="grid gap-1.5 content-start">
            <Label htmlFor={`${idPrefix}-path`}>{t('location')}</Label>
            <Input
              id={`${idPrefix}-path`}
              placeholder="/"
              spellCheck={false}
              autoComplete="off"
              className="font-mono"
              value={value.location}
              onChange={(e) => onChange({ location: e.target.value })}
            />
          </div>
          <div className="grid gap-1.5 content-start">
            <Label htmlFor={`${idPrefix}-rewrite`}>{t('rewriteTo')}</Label>
            <Input
              id={`${idPrefix}-rewrite`}
              placeholder={t('rewriteSame')}
              spellCheck={false}
              autoComplete="off"
              className="font-mono"
              value={value.proxyRewrite}
              onChange={(e) => onChange({ proxyRewrite: e.target.value })}
            />
          </div>
        </FieldGrid>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label={t('removeLocation')}
          title={t('removeLocation')}
          onClick={onRemove}
        >
          <Trash2Icon className="w-4 h-4" />
        </Button>
      </div>

      <div className="grid gap-1.5">
        <Label>{t('upstreams')}</Label>
        <OriginRows
          rows={value.upstreams}
          onChange={(upstreams) => onChange({ upstreams })}
          idPrefix={idPrefix}
        />
      </div>

      {balanced && (
        <div className="sm:w-1/2">
          <OptionSelect
            label={t('balance')}
            value={value.balanceStrategy}
            options={balances}
            onChange={(balanceStrategy) => onChange({ balanceStrategy })}
          />
        </div>
      )}

      <div className="grid gap-1.5">
        <Label htmlFor={`${idPrefix}-headers`}>{t('headers')}</Label>
        <Textarea
          id={`${idPrefix}-headers`}
          placeholder={'X-Real-IP: $remote_addr\nX-Env: prod'}
          spellCheck={false}
          autoComplete="off"
          rows={2}
          className="font-mono min-h-0"
          value={value.headers}
          onChange={(e) => onChange({ headers: e.target.value })}
        />
      </div>
    </fieldset>
  );
}

interface Props {
  value: ServerFormValue;
  onChange: (patch: Partial<ServerFormValue>) => void;
}

/** The fields of the edge Server of a resource: its name and locations */
export function ServerFields({ value, onChange }: Props) {
  const t = useTranslations('ResourceSetting');
  const patchLocation = (i: number, patch: Partial<LocationFormValue>) =>
    onChange({ locations: value.locations.map((l, j) => (j === i ? { ...l, ...patch } : l)) });

  return (
    <div className="flex flex-col gap-3">
      <div className="grid gap-1.5 sm:w-1/2">
        <Label htmlFor="server-name">{t('serverName')}</Label>
        <Input
          id="server-name"
          placeholder="my_site"
          spellCheck={false}
          autoComplete="off"
          value={value.name}
          onChange={(e) => onChange({ name: e.target.value })}
        />
      </div>

      {value.locations.length === 0 && (
        <div className="text-sm text-muted-foreground p-4 border rounded-md">
          {t('noLocationsEdit')}
        </div>
      )}
      {value.locations.map((l, i) => (
        <LocationFields
          key={i}
          index={i}
          value={l}
          onChange={(patch) => patchLocation(i, patch)}
          onRemove={() => onChange({ locations: value.locations.filter((_, j) => j !== i) })}
        />
      ))}

      <div className="flex flex-wrap items-start justify-between gap-3">
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="w-fit"
          onClick={() => onChange({ locations: [...value.locations, createEmptyLocation()] })}
        >
          <PlusIcon className="w-4 h-4" />
          {t('addLocation')}
        </Button>
        {/* Once for every location, not under each header field */}
        <p className="text-xs text-muted-foreground max-w-2xl">{t('headerHint')}</p>
      </div>
    </div>
  );
}
