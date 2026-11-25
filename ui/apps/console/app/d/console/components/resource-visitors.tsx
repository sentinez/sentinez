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
import type { VisitorPoint } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { useApi } from '@/hooks/use-api';
import {
  VISITORS_RESOLUTION,
  bucketMillisOf,
  durationOf,
  getUniqueVisitorHistogram,
  listVisitors,
} from '@/lib/api/analytic';
import { RESOURCE_CHART } from '@/lib/chart';
import { CompactNumber } from '@/components/compact-number';
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
 * The visitors of a resource over the last 24 hours, per bucket of the
 * interval the API answers with: the unique visitor chart and the client
 * ips of the most requests.
 */
export function ResourceVisitors({ resourceId }: { resourceId: string }) {
  const t = useTranslations('Resource');
  const req = { resolution: VISITORS_RESOLUTION, resourceIds: [resourceId] };
  const histogram = useApi(getUniqueVisitorHistogram, req);
  const visitors = useApi(listVisitors, req);

  const buckets = useMemo(() => (histogram.data?.points ?? []).map(bucketOf), [histogram.data]);
  const top = useMemo(() => topVisitorsOf(visitors.data?.visitors ?? []), [visitors.data]);
  const interval = durationOf(bucketMillisOf(VISITORS_RESOLUTION, histogram.data?.interval));

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('uniqueVisitors')}</CardTitle>
        <CardDescription>{t('uniqueVisitorsDescription', interval)}</CardDescription>
      </CardHeader>
      {/* two parts: the chart, then the top visitors, side by side on a
          wide page (ruled off by a border), one under the other otherwise */}
      <CardContent className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,22rem)]">
        <div className="min-w-0">
          {histogram.isLoading || visitors.isLoading ? (
            <Skeleton className={RESOURCE_CHART.skeletonClassName} />
          ) : (
            <ResourceVisitorChart data={buckets} total={top.length} />
          )}
        </div>

        {/* The chart sets the height of the card: on a wide page this part
            is laid over its column (absolute), so it adds no height, and the
            table scrolls within the chart height; stacked, the table scrolls
            past max-h-80 */}
        <section className="relative min-w-0 border-t pt-6 lg:border-t-0 lg:border-l lg:pt-0">
          <div className="flex flex-col lg:absolute lg:inset-0 lg:pl-6">
            <h3 className="font-semibold leading-none">{t('topVisitors')}</h3>
            <p className="mt-1.5 text-sm text-muted-foreground">
              {t('topVisitorsDescription', { count: TOP_VISITORS })}
            </p>
            <div className="mt-4 max-h-80 min-h-0 flex-1 overflow-y-auto lg:max-h-none">
              {visitors.isLoading ? (
                <Skeleton className="h-full min-h-40 w-full" />
              ) : (
                <TopVisitorTable top={top.slice(0, TOP_VISITORS)} />
              )}
            </div>
          </div>
        </section>
      </CardContent>
    </Card>
  );
}

// The client ips of the most requests, their requests and the hour they
// were last seen
function TopVisitorTable({ top }: { top: TopVisitor[] }) {
  const t = useTranslations('Resource');
  const tc = useTranslations('Common');
  const format = useFormatter();
  return (
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
          top.map((v) => (
            <TableRow key={v.clientIp}>
              <TableCell className="font-mono">{v.clientIp}</TableCell>
              <TableCell className="text-right tabular-nums">
                <CompactNumber value={v.requests} />
              </TableCell>
              <TableCell className="text-right">
                {format.dateTime(new Date(v.lastSeen), HOUR_FORMAT)}
              </TableCell>
            </TableRow>
          ))
        )}
      </TableBody>
    </Table>
  );
}
