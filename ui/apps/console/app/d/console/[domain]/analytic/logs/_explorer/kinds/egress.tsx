'use client';

import { useMemo } from 'react';
import { useFormatter, useTranslations } from 'next-intl';

import { EventKind } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { EgressEvent } from '@sentinez/proto/sentinez/types/net/http/v1/event';
import { listHttpEgressEvents, type ApiOptions } from '@/lib/api/analytic';
import { StatusBadge } from '../status-badge';
import { queryParamsOf, statusClassOf } from '../status';
import type { EventConfig, ListRequest } from '../types';
import { MonoCell, PathCell, columnOf, statusSeries } from './shared';

const list = async (req: ListRequest, options?: ApiOptions) =>
  (await listHttpEgressEvents(req, options)).events;

/** The explorer of the requests the edge proxied to the origins */
export function useEgressConfig(): EventConfig<EgressEvent> {
  const t = useTranslations('Logs');
  const format = useFormatter();
  return useMemo(() => {
    const c = columnOf(t);
    const latencyOf = (e: EgressEvent) =>
      `${format.number(e.latencyMicros / 1000, { maximumFractionDigits: 1 })} ms`;
    const lengthOf = (e: EgressEvent) =>
      e.contentLength < 0 ? t('streamed') : format.number(e.contentLength);
    return {
      kind: EventKind.EVENT_KIND_HTTP_EGRESS,
      title: t('egress.title'),
      subtitle: t('egress.subtitle'),
      list,
      idOf: (e) => e.id,
      timeOf: (e) => e.timestamp,
      columns: [
        {
          id: 'status',
          size: 80,
          header: c('status'),
          cell: ({ row }) => <StatusBadge status={row.original.status} />,
        },
        {
          id: 'method',
          size: 80,
          header: c('method'),
          cell: ({ row }) => <MonoCell value={row.original.method} />,
        },
        {
          id: 'upstream',
          size: 200,
          header: c('upstream'),
          cell: ({ row }) => <MonoCell value={row.original.upstream} />,
        },
        {
          id: 'path',
          size: 0,
          header: c('path'),
          cell: ({ row }) => <PathCell path={row.original.path} query={row.original.queries} />,
        },
        {
          id: 'latency',
          size: 100,
          header: () => <div className="text-right">{c('latency')}</div>,
          cell: ({ row }) => (
            <div className="text-right font-mono text-xs tabular-nums">
              {latencyOf(row.original)}
            </div>
          ),
        },
        {
          id: 'contentLength',
          size: 120,
          header: () => <div className="text-right">{c('contentLength')}</div>,
          cell: ({ row }) => (
            <div className="text-right font-mono text-xs tabular-nums">
              {lengthOf(row.original)}
            </div>
          ),
        },
      ],
      series: statusSeries(t),
      seriesOf: statusClassOf,
      fields: [
        { name: 'upstream', label: c('upstream'), get: (e) => e.upstream },
        { name: 'status', label: c('status'), get: (e) => String(e.status) },
        { name: 'method', label: c('method'), get: (e) => e.method },
        { name: 'path', label: c('path'), get: (e) => e.path },
        { name: 'content_type', label: c('contentType'), get: (e) => e.contentType },
        { name: 'protocol', label: c('protocol'), get: (e) => e.protocol },
      ],
      // the leg to the origin of the request of the same id
      requestOf: (e) => ({ id: e.id, time: e.timestamp }),
      text: (e) => [e.id, String(e.status), e.method, e.upstream, e.path, e.queries, e.contentType],
      detail: {
        title: (e) => (
          <>
            <StatusBadge status={e.status} />
            <span className="min-w-0 truncate font-mono text-sm">
              {e.method} {e.upstream}
              {e.path}
            </span>
          </>
        ),
        rows: (e) => [
          [c('status'), String(e.status)],
          [c('method'), e.method],
          [c('upstream'), e.upstream],
          [c('scheme'), e.scheme],
          [c('path'), e.path],
          [c('protocol'), e.protocol],
          [c('latency'), latencyOf(e)],
          [c('contentType'), e.contentType],
          [c('contentLength'), lengthOf(e)],
        ],
        sections: (e) => [{ title: t('detail.queries'), pairs: queryParamsOf(e.queries) }],
        toJSON: (e) => EgressEvent.toJSON(e),
      },
    };
  }, [t, format]);
}
