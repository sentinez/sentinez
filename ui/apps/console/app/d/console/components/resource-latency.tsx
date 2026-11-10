'use client';

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
import { useApi } from '@/hooks/use-api';
import { getEgressLatency } from '@/lib/api/analytic';
import { latencyMs } from '@/lib/latency';

// The columns after the upstream: the label key, and the field of the row
const LATENCY_COLUMNS = [
  { key: 'avg', field: 'avgMicros' },
  { key: 'p50', field: 'p50Micros' },
  { key: 'p95', field: 'p95Micros' },
  { key: 'p99', field: 'p99Micros' },
  { key: 'max', field: 'maxMicros' },
] as const;

/**
 * The latency of the requests the edge proxied to each upstream of a
 * resource over the last 24 hours: how many, and their average, median,
 * p95, p99 and maximum.
 */
export function ResourceLatency({ resourceId }: { resourceId: string }) {
  const t = useTranslations('Resource');
  const format = useFormatter();

  // Without a range the API measures the last 24 hours
  const { data, isLoading } = useApi(getEgressLatency, {
    selection: { resourceIds: [resourceId] },
  });
  const upstreams = data?.upstreams ?? [];

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('latency')}</CardTitle>
        <CardDescription>{t('latencyDescription')}</CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-40 w-full" />
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('upstream')}</TableHead>
                <TableHead className="text-right">{t('requests')}</TableHead>
                {LATENCY_COLUMNS.map((c) => (
                  <TableHead key={c.key} className="text-right">
                    {t(c.key)}
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {upstreams.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={2 + LATENCY_COLUMNS.length} className="h-24 text-center">
                    {t('noLatency')}
                  </TableCell>
                </TableRow>
              ) : (
                upstreams.map((u) => (
                  <TableRow key={u.upstream}>
                    <TableCell className="font-mono break-all">{u.upstream}</TableCell>
                    <TableCell className="text-right tabular-nums">
                      {format.number(u.total)}
                    </TableCell>
                    {LATENCY_COLUMNS.map((c) => (
                      <TableCell key={c.key} className="text-right tabular-nums">
                        {latencyMs(format, u[c.field])}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
