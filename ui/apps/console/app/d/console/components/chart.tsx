'use client';

import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from 'recharts';

import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@sentinez/ui/components/chart';
import type { HistogramPoint } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { useFormatter, useTranslations } from 'next-intl';
import { useMemo } from 'react';

// VisitorBucket the unique visitors of the bucket starting at time (millis)
export type VisitorBucket = { time: number; visitors: number };

// The bucket of a point of the unique visitor histogram
export const bucketOf = (p: HistogramPoint): VisitorBucket => ({
  time: p.time?.getTime() ?? 0,
  visitors: p.total,
});

// The start of an hourly bucket, as "14:00"
export const HOUR_FORMAT = { hour: '2-digit', minute: '2-digit' } as const;

// The area of the visitors series, shared by both charts
const VISITOR_AREA = {
  dataKey: 'visitors',
  type: 'monotone',
  stroke: 'var(--color-visitors)',
  fill: 'var(--color-visitors)',
  strokeWidth: 2,
} as const;

function useVisitorConfig() {
  const t = useTranslations('Resource');
  return useMemo(
    () => ({ visitors: { label: t('visitors'), color: 'var(--chart-1)' } }) satisfies ChartConfig,
    [t],
  );
}

// The tooltip label of a bucket: its start, date and time
function useBucketLabel() {
  const format = useFormatter();
  return (_: unknown, payload: readonly { payload?: VisitorBucket }[]) => {
    const time = payload?.[0]?.payload?.time;
    return time ? format.dateTime(new Date(time), { dateStyle: 'short', timeStyle: 'short' }) : '';
  };
}

// Sparkline of the unique visitors of a resource, oldest bucket first
export function UniqueVisitorChart({ data }: { data: VisitorBucket[] }) {
  const t = useTranslations('Domain');
  const config = useVisitorConfig();
  const bucketLabel = useBucketLabel();
  if (data.length === 0) {
    return <div className="text-xs text-muted-foreground">{t('noVisitorData')}</div>;
  }

  return (
    <div className="w-full h-16 overflow-hidden">
      <ChartContainer config={config} className="h-16 w-full">
        <AreaChart data={data} margin={{ top: 4, bottom: 4, left: 0, right: 0 }}>
          <ChartTooltip
            cursor={false}
            content={<ChartTooltipContent labelFormatter={bucketLabel} />}
          />
          <Area {...VISITOR_AREA} dot={false} fillOpacity={0.4} />
        </AreaChart>
      </ChartContainer>
    </div>
  );
}

// ResourceVisitorChart the unique visitors per hour of a resource, oldest
// first, with the unique visitors of the whole range (total: an ip seen in
// several hours counted once) and the busiest hour
export function ResourceVisitorChart({ data, total }: { data: VisitorBucket[]; total: number }) {
  const t = useTranslations('Resource');
  const format = useFormatter();
  const config = useVisitorConfig();
  const bucketLabel = useBucketLabel();
  const hourOf = (time: number) => format.dateTime(new Date(time), HOUR_FORMAT);
  const peak = data.reduce<VisitorBucket | null>(
    (max, d) => (d.visitors > (max?.visitors ?? 0) ? d : max),
    null,
  );

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-8">
        <div>
          <div className="text-sm text-muted-foreground">{t('total24h')}</div>
          <div className="text-2xl font-bold tabular-nums">{format.number(total)}</div>
        </div>
        <div>
          <div className="text-sm text-muted-foreground">{t('peakHour')}</div>
          <div className="text-2xl font-bold tabular-nums">
            {peak ? hourOf(peak.time) : '—'}{' '}
            {peak && (
              <span className="text-sm font-normal text-muted-foreground">
                ({format.number(peak.visitors)})
              </span>
            )}
          </div>
        </div>
      </div>
      <ChartContainer config={config} className="h-64 w-full">
        <AreaChart data={data} margin={{ top: 8, bottom: 0, left: 0, right: 8 }}>
          <CartesianGrid vertical={false} />
          <XAxis
            dataKey="time"
            tickFormatter={hourOf}
            tickLine={false}
            axisLine={false}
            tickMargin={8}
            interval={3}
          />
          <YAxis tickLine={false} axisLine={false} width={36} allowDecimals={false} />
          <ChartTooltip
            cursor={false}
            content={<ChartTooltipContent indicator="line" labelFormatter={bucketLabel} />}
          />
          <Area {...VISITOR_AREA} fillOpacity={0.3} />
        </AreaChart>
      </ChartContainer>
    </div>
  );
}
