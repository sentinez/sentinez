'use client';

import { useMemo } from 'react';
import { useFormatter, useTranslations } from 'next-intl';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Skeleton } from '@sentinez/ui/components/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@sentinez/ui/components/table';
import {
  HistogramResolution,
  type VisitorPoint,
} from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { useApi } from '@/hooks/use-api';
import { getUniqueVisitorHistogram, listVisitors } from '@/lib/api/analytic';
import { HOUR_FORMAT, ResourceVisitorChart, bucketOf } from './chart';

// The client ips listed in the top visitors card
const TOP_VISITORS = 10;

// TopVisitor the requests of a client ip over the last 24 hours
type TopVisitor = { clientIp: string; requests: number; lastSeen: number };

// The client ips of visitors, the most requests first
function topVisitorsOf(visitors: VisitorPoint[]): TopVisitor[] {
  const byIp = new Map<string, TopVisitor>();
  for (const v of visitors) {
    const time = v.time?.getTime() ?? 0;
    const top = byIp.get(v.clientIp) ?? { clientIp: v.clientIp, requests: 0, lastSeen: 0 };
    top.requests += v.total;
    top.lastSeen = Math.max(top.lastSeen, time);
    byIp.set(v.clientIp, top);
  }
  return [...byIp.values()].sort((a, b) => b.requests - a.requests);
}

/**
 * The visitors of a resource over the last 24 hours, per hour: the unique
 * visitor chart and the client ips of the most requests.
 */
export function ResourceVisitors({ resourceId }: { resourceId: string }) {
  const t = useTranslations('Resource');
  const tc = useTranslations('Common');
  const format = useFormatter();
  const req = {
    resolution: HistogramResolution.HISTOGRAM_RESOLUTION_HOUR,
    resourceIds: [resourceId],
  };
  const histogram = useApi(getUniqueVisitorHistogram, req);
  const visitors = useApi(listVisitors, req);

  const buckets = useMemo(() => (histogram.data?.points ?? []).map(bucketOf), [histogram.data]);
  const top = useMemo(() => topVisitorsOf(visitors.data?.visitors ?? []), [visitors.data]);

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>{t('uniqueVisitors')}</CardTitle>
          <CardDescription>{t('uniqueVisitorsDescription')}</CardDescription>
        </CardHeader>
        <CardContent>
          {histogram.isLoading || visitors.isLoading ? (
            <Skeleton className="h-80 w-full" />
          ) : (
            <ResourceVisitorChart data={buckets} total={top.length} />
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('topVisitors')}</CardTitle>
          <CardDescription>
            {t('topVisitorsDescription', { count: TOP_VISITORS })}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {visitors.isLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('clientIp')}</TableHead>
                  <TableHead className="text-right">{t('requests')}</TableHead>
                  <TableHead className="text-right">{t('lastSeen')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {top.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={3} className="h-24 text-center">
                      {tc('noResults')}
                    </TableCell>
                  </TableRow>
                ) : (
                  top.slice(0, TOP_VISITORS).map((v) => (
                    <TableRow key={v.clientIp}>
                      <TableCell className="font-mono">{v.clientIp}</TableCell>
                      <TableCell className="text-right tabular-nums">
                        {format.number(v.requests)}
                      </TableCell>
                      <TableCell className="text-right">
                        {format.dateTime(new Date(v.lastSeen), HOUR_FORMAT)}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </>
  );
}
