'use client';

import type { ReactNode } from 'react';
import { useTranslations } from 'next-intl';
import type { ColumnDef } from '@tanstack/react-table';

import { cn } from '@sentinez/ui/lib/utils';
import type messages from '@/messages/en.json';
import { RowLink } from '../request-link';
import { STATUS_CLASSES, STATUS_COLORS } from '../status';
import type { FieldDef, RequestRef, SeriesDef } from '../types';

export type LogsTranslator = ReturnType<typeof useTranslations<'Logs'>>;
export type ColumnKey = keyof (typeof messages)['Logs']['column'];

// The label of a column of the Logs messages
export const columnOf = (t: LogsTranslator) => (key: ColumnKey) => t(`column.${key}`);

// The series of the histograms keyed by an http status: one per class
export const statusSeries = (t: LogsTranslator): SeriesDef[] =>
  STATUS_CLASSES.map((id) => ({
    id,
    label: t(`series.${id}`),
    color: STATUS_COLORS[id],
  }));

/** A cell of monospace text, truncated */
export function MonoCell({ value, muted = false }: { value: string; muted?: boolean }) {
  return (
    <div
      className={cn('truncate font-mono text-xs', muted && 'text-muted-foreground')}
      title={value}
    >
      {value || '—'}
    </div>
  );
}

/** The host and path of a request, its query muted */
export function PathCell({ host, path, query }: { host?: string; path: string; query?: string }) {
  const full = `${host ?? ''}${path}${query ? `?${query}` : ''}`;
  return (
    <div className="truncate font-mono text-xs" title={full}>
      {host && <span className="text-muted-foreground">{host}</span>}
      {path}
      {query && <span className="text-muted-foreground">?{query}</span>}
    </div>
  );
}

/**
 * Links to the rules of ids, the page of each under the security section
 * of the domain (`sec-rule` or `rate-limiter`); a click does not open the
 * row it is in.
 */
export function RuleLinks({ ids, href }: { ids: string[]; href: (id: string) => string }) {
  if (ids.length === 0) return <span>—</span>;
  return (
    <span className="flex min-w-0 flex-wrap gap-x-2">
      {ids.map((id) => (
        <RowLink key={id} href={href(id)} className="truncate">
          {id}
        </RowLink>
      ))}
    </span>
  );
}

// RuleRequest the fields of the request a rule event (SecRuleEvent,
// CoreRuleEvent) carries
export type RuleRequest = {
  requestId: string;
  requestIp: string;
  requestDomain: string;
  requestScheme: string;
  requestPath: string;
  requestQuery: string;
  matchedTime: number;
};

type ColumnOf = ReturnType<typeof columnOf>;

// The columns of the request of a rule event: its path, its client ip
export const ruleRequestColumns = <T extends RuleRequest>(c: ColumnOf): ColumnDef<T>[] => [
  {
    id: 'path',
    size: 0,
    header: c('path'),
    cell: ({ row }) => (
      <PathCell
        host={row.original.requestDomain}
        path={row.original.requestPath}
        query={row.original.requestQuery}
      />
    ),
  },
  {
    id: 'clientIp',
    size: 140,
    header: c('clientIp'),
    cell: ({ row }) => <MonoCell value={row.original.requestIp} />,
  },
];

// The fields of the request of a rule event
export const ruleRequestFields = <T extends RuleRequest>(c: ColumnOf): FieldDef<T>[] => [
  { name: 'request_ip', label: c('clientIp'), get: (e) => e.requestIp },
  { name: 'request_domain', label: c('domain'), get: (e) => e.requestDomain },
  { name: 'request_path', label: c('path'), get: (e) => e.requestPath },
];

// The detail rows of the request of a rule event
export const ruleRequestRows = (c: ColumnOf, e: RuleRequest): [string, ReactNode][] => [
  [c('domain'), e.requestDomain],
  [c('scheme'), e.requestScheme],
  [c('path'), e.requestPath],
  [c('query'), e.requestQuery],
  [c('clientIp'), e.requestIp],
];

// The request a rule event is about
export const ruleRequestOf = (e: RuleRequest): RequestRef => ({
  id: e.requestId,
  time: e.matchedTime,
});
