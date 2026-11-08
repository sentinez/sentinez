'use client';

import { useMemo } from 'react';
import { useTranslations } from 'next-intl';

import { EventKind } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { IngressEvent } from '@sentinez/proto/sentinez/types/net/http/v1/event';
import { listHttpIngressEvents, type ApiOptions } from '@/lib/api/analytic';
import { StatusBadge } from '../status-badge';
import { headerLinesOf, queryParamsOf, statusClassOf } from '../status';
import type { EventConfig, ListRequest } from '../types';
import { MonoCell, PathCell, columnOf, statusSeries } from './shared';

const list = async (req: ListRequest, options?: ApiOptions) =>
  (await listHttpIngressEvents(req, options)).events;

/** The explorer of the requests received by the edge */
export function useIngressConfig(): EventConfig<IngressEvent> {
  const t = useTranslations('Logs');
  return useMemo(() => {
    const c = columnOf(t);
    return {
      kind: EventKind.EVENT_KIND_HTTP_INGRESS,
      title: t('ingress.title'),
      subtitle: t('ingress.subtitle'),
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
          id: 'path',
          size: 0,
          header: c('path'),
          cell: ({ row }) => (
            <PathCell
              host={row.original.host}
              path={row.original.path}
              query={row.original.queries}
            />
          ),
        },
        {
          id: 'clientIp',
          size: 140,
          header: c('clientIp'),
          cell: ({ row }) => <MonoCell value={row.original.clientIp} />,
        },
        {
          id: 'country',
          size: 80,
          header: c('country'),
          cell: ({ row }) => <MonoCell value={row.original.country} />,
        },
        {
          id: 'userAgent',
          size: 240,
          header: c('userAgent'),
          cell: ({ row }) => <MonoCell value={row.original.userAgent} muted />,
        },
      ],
      series: statusSeries(t),
      seriesOf: statusClassOf,
      fields: [
        { name: 'method', label: c('method'), get: (e) => e.method },
        { name: 'status', label: c('status'), get: (e) => String(e.status) },
        { name: 'host', label: c('host'), get: (e) => e.host },
        { name: 'client_ip', label: c('clientIp'), get: (e) => e.clientIp },
        { name: 'country', label: c('country'), get: (e) => e.country },
        { name: 'protocol', label: c('protocol'), get: (e) => e.protocol },
        { name: 'path', label: c('path'), get: (e) => e.path },
        { name: 'user_agent', label: c('userAgent'), get: (e) => e.userAgent },
      ],
      byId: async ({ id, resourceId, from, to }, options) => {
        const res = await listHttpIngressEvents(
          {
            requestIds: [id],
            resourceIds: [resourceId],
            from,
            to,
            page: { index: 1, size: 1, total: false },
          },
          options,
        );
        return res.events[0] ?? null;
      },
      text: (e) => [
        e.id,
        String(e.status),
        e.method,
        e.host,
        e.path,
        e.queries,
        e.clientIp,
        e.userAgent,
        e.country,
      ],
      detail: {
        title: (e) => (
          <>
            <StatusBadge status={e.status} />
            <span className="min-w-0 truncate font-mono text-sm">
              {e.method} {e.path}
            </span>
          </>
        ),
        rows: (e) => [
          [c('status'), String(e.status)],
          [c('method'), e.method],
          [c('scheme'), e.scheme],
          [c('host'), e.host],
          [c('path'), e.path],
          [c('protocol'), e.protocol],
          [c('clientIp'), e.clientIp],
          [c('remoteAddress'), e.remoteAddress],
          [c('country'), e.country],
          [c('asn'), e.asn ? `AS${e.asn}` : ''],
          [c('contentType'), e.contentType],
          [c('userAgent'), e.userAgent],
          [c('fingerprint'), e.fingerprint],
        ],
        sections: (e) => [
          { title: t('detail.queries'), pairs: queryParamsOf(e.queries) },
          { title: t('detail.headers'), pairs: headerLinesOf(e.headers) },
        ],
        toJSON: (e) => IngressEvent.toJSON(e),
      },
    };
  }, [t]);
}
