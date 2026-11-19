'use client';

import { useMemo, type ReactNode } from 'react';
import { useFormatter, useTranslations } from 'next-intl';

import { AuditLog } from '@sentinez/proto/sentinez/types/audit/v1/audit';
import { controlPlaneToJSON } from '@sentinez/proto/sentinez/types/v1/known';
import {
  getAuditLogHistogram,
  getAuditLogTopValues,
  listAuditLogs,
  type ApiOptions,
} from '@/lib/api/analytic';
import { latencyMs } from '@/lib/latency';
import { FieldChangeValue } from './field-change';
import { MonoCell } from '@/components/event-explorer/kinds/shared';
import { idsMatching } from '@/components/event-explorer/query';
import { STATUS_COLORS } from '@/components/event-explorer/status';
import { ColoredBadge } from '@/components/event-explorer/status-badge';
import type {
  EventConfig,
  FieldDef,
  HistogramRequest,
  ListRequest,
  SeriesDef,
  TopValuesRequest,
} from '@/components/event-explorer/types';

// The names of the grpc codes, by their number
const GRPC_CODES = [
  'OK',
  'CANCELLED',
  'UNKNOWN',
  'INVALID_ARGUMENT',
  'DEADLINE_EXCEEDED',
  'NOT_FOUND',
  'ALREADY_EXISTS',
  'PERMISSION_DENIED',
  'RESOURCE_EXHAUSTED',
  'FAILED_PRECONDITION',
  'ABORTED',
  'OUT_OF_RANGE',
  'UNIMPLEMENTED',
  'INTERNAL',
  'UNAVAILABLE',
  'DATA_LOSS',
  'UNAUTHENTICATED',
];

const codeName = (code: number) => GRPC_CODES[code] ?? String(code);

// CodeClass the class of a grpc code, a series of the histogram: the call
// succeeded, was refused for what the caller sent, or failed in the server
type CodeClass = 'ok' | 'client' | 'server';

const CLIENT_CODES = new Set([1, 3, 5, 6, 7, 9, 10, 11, 16]);

const codeClassOf = (code: number): CodeClass =>
  code === 0 ? 'ok' : CLIENT_CODES.has(code) ? 'client' : 'server';

const CODE_COLORS: Record<CodeClass, string> = {
  ok: STATUS_COLORS.s2xx,
  client: STATUS_COLORS.s4xx,
  server: STATUS_COLORS.s5xx,
};

// The log ids a filter names, which the list asks the API for
const idsOf = idsMatching(/^aud_\S+$/);

// The request of the API of an explorer request: within the tenant, else
// of the logs of no tenant
const auditRequest = <R,>(req: R, tenantId: string) =>
  tenantId ? { ...req, tenantIds: [tenantId] } : { ...req, withoutTenant: true };

/** The grpc code of a call, colored by its class like the histogram */
function CodeBadge({ code }: { code: number }) {
  return (
    <ColoredBadge color={CODE_COLORS[codeClassOf(code)]} className="text-[11px]">
      {codeName(code)}
    </ColoredBadge>
  );
}

// The service and method of a full grpc method, "/pkg.Service/Method"
function methodParts(method: string): { service: string; name: string } {
  const [, full = '', name = method] = method.split('/');
  return { service: full.split('.').pop() ?? full, name };
}

/** The method of a call, its service muted */
function MethodCell({ method }: { method: string }) {
  const { service, name } = methodParts(method);
  return (
    <div className="truncate font-mono text-xs" title={method}>
      <span className="text-muted-foreground">{service}/</span>
      {name}
    </div>
  );
}

// The leaves of a JSON value as dotted keys, e.g. secrule.owner.resourceId
function leavesOf(value: unknown, prefix = ''): [string, ReactNode][] {
  if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
    return Object.entries(value).flatMap(([k, v]) => leavesOf(v, prefix ? `${prefix}.${k}` : k));
  }
  const text = typeof value === 'string' ? value : JSON.stringify(value);
  return [
    [
      prefix || '—',
      <span key={prefix} className="break-all">
        {text}
      </span>,
    ],
  ];
}

/** The explorer of the audit logs of the tenant of tenantId, of no tenant when empty */
export function useAuditLogConfig(tenantId: string): EventConfig<AuditLog> {
  const t = useTranslations('AuditLog');
  const format = useFormatter();
  // apart from the rest: the explorer refetches when its fetchers change
  const apis = useMemo(
    () => ({
      list: async (req: ListRequest, options?: ApiOptions) =>
        (await listAuditLogs(auditRequest(req, tenantId), options)).logs,
      histogram: (req: HistogramRequest, options?: ApiOptions) =>
        getAuditLogHistogram(auditRequest(req, tenantId), options),
      topValues: (req: TopValuesRequest, options?: ApiOptions) =>
        getAuditLogTopValues(auditRequest(req, tenantId), options),
    }),
    [tenantId],
  );
  return useMemo(() => {
    const series: SeriesDef[] = (['ok', 'client', 'server'] as const).map((id) => ({
      id,
      label: t(`series.${id}`),
      color: CODE_COLORS[id],
    }));
    const fields: FieldDef<AuditLog>[] = [
      { name: 'method', label: t('column.method'), get: (l) => l.method },
      {
        name: 'code',
        label: t('column.code'),
        get: (l) => String(l.code),
        format: (v) => codeName(Number(v)),
      },
      { name: 'user_id', label: t('column.userId'), get: (l) => l.userId },
      { name: 'remote_addr', label: t('column.remoteAddr'), get: (l) => l.remoteAddr },
      { name: 'user_agent', label: t('column.userAgent'), get: (l) => l.userAgent },
      {
        name: 'control_plane',
        label: t('column.controlPlane'),
        get: (l) => String(l.controlPlane),
        format: (v) => controlPlaneToJSON(Number(v)),
      },
    ];
    const latencyOf = (l: AuditLog) =>
      l.latency ? latencyMs(format, l.latency.seconds * 1_000_000 + l.latency.nanos / 1000) : '—';
    const userOf = (l: AuditLog) => l.name || l.userId;
    return {
      title: t('title'),
      subtitle: t('subtitle'),
      ...apis,
      idsOf,
      series,
      seriesOf: codeClassOf,
      fields,
      text: auditLogText,
      idOf: (l) => l.id,
      timeOf: (l) => l.timestamp,
      columns: [
        {
          id: 'code',
          size: 150,
          header: t('column.code'),
          cell: ({ row }) => <CodeBadge code={row.original.code} />,
        },
        {
          id: 'method',
          size: 0,
          header: t('column.method'),
          cell: ({ row }) => <MethodCell method={row.original.method} />,
        },
        {
          id: 'user',
          size: 160,
          header: t('column.user'),
          cell: ({ row }) => <MonoCell value={userOf(row.original)} title={row.original.userId} />,
        },
        {
          id: 'resource',
          size: 180,
          header: t('column.resource'),
          cell: ({ row }) => <MonoCell value={row.original.resourceId} muted />,
        },
        {
          id: 'remoteAddr',
          size: 170,
          header: t('column.remoteAddr'),
          cell: ({ row }) => <MonoCell value={row.original.remoteAddr} />,
        },
        {
          id: 'latency',
          size: 90,
          header: () => <div className="text-right">{t('column.latency')}</div>,
          cell: ({ row }) => (
            <div className="text-right font-mono text-xs tabular-nums">
              {latencyOf(row.original)}
            </div>
          ),
        },
      ],
      detail: {
        title: (l) => (
          <>
            <CodeBadge code={l.code} />
            <span className="min-w-0 truncate font-mono text-sm">{methodParts(l.method).name}</span>
          </>
        ),
        rows: (l) => [
          [t('column.method'), l.method],
          [t('column.code'), `${codeName(l.code)} (${l.code})`],
          [t('column.error'), l.error],
          [t('column.userId'), l.userId],
          [t('column.name'), l.name],
          [t('column.controlPlane'), controlPlaneToJSON(l.controlPlane)],
          [t('column.remoteAddr'), l.remoteAddr],
          [t('column.userAgent'), l.userAgent],
          [t('column.tenant'), l.tenantId],
          [t('column.resource'), l.resourceId],
          [t('column.latency'), latencyOf(l)],
        ],
        sections: (l) => [
          {
            title: t('changes'),
            pairs: (l.change?.fields ?? []).map((c): [string, ReactNode] => [
              c.field,
              <FieldChangeValue key={c.field} change={c} />,
            ]),
          },
          { title: t('request'), pairs: l.request ? leavesOf(l.request) : [] },
        ],
        toJSON: (l) => AuditLog.toJSON(l),
      },
    };
  }, [t, format, apis]);
}

/** The texts a free-text term of the filter is searched in */
const auditLogText = (l: AuditLog): string[] => [
  l.id,
  l.method,
  codeName(l.code),
  l.error,
  l.userId,
  l.name,
  l.remoteAddr,
  l.userAgent,
  l.resourceId,
];
