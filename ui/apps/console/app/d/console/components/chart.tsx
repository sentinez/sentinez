'use client';

import { Area, AreaChart, CartesianGrid, Line, LineChart, XAxis, YAxis } from 'recharts';

import { ChartContainer, ChartTooltip } from '@sentinez/ui/components/chart';
import type { HistogramPoint } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { useFormatter, useTranslations } from 'next-intl';
import { useCallback, useMemo, useState } from 'react';
import { compactNumber } from '@/lib/number';
import { CompactTooltipContent } from '@/components/chart-tooltip';
import { CompactNumber } from '@/components/compact-number';
import { CHART_COLORS, RESOURCE_CHART, SPARKLINE, seriesConfig } from '@/lib/chart';

// VisitorBucket the unique visitors of the bucket starting at time (millis)
export type VisitorBucket = { time: number; visitors: number };

// The bucket of a point of the unique visitor histogram
export const bucketOf = (p: HistogramPoint): VisitorBucket => ({
  time: p.time?.getTime() ?? 0,
  visitors: p.total,
});

// The start of an hourly bucket, as "14:00"
export const HOUR_FORMAT = { hour: '2-digit', minute: '2-digit' } as const;

// The visitors series: the line of the sparkline, the area of the resource
// chart
const VISITOR_AREA = {
  dataKey: 'visitors',
  stroke: 'var(--color-visitors)',
  fill: 'var(--color-visitors)',
} as const;

function useVisitorConfig() {
  const t = useTranslations('Resource');
  return useMemo(
    () => seriesConfig([{ id: 'visitors', label: t('visitors'), color: CHART_COLORS[3] }]),
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
// The tooltip of the sparkline, on one line: the hour of the bucket (the
// buckets are the last 24 hours, the hour is enough) and its visitors
function SparklineTooltip({
  active,
  payload,
}: {
  active?: boolean;
  payload?: readonly { payload?: VisitorBucket }[];
}) {
  const t = useTranslations('Resource');
  const format = useFormatter();
  const bucket = payload?.[0]?.payload;
  if (!active || !bucket) return null;
  return (
    <div
      className={SPARKLINE.tooltipClassName}
      style={{ width: SPARKLINE.tooltipWidth }}
      aria-label={t('visitors')}
    >
      <span className="text-muted-foreground">
        {format.dateTime(new Date(bucket.time), HOUR_FORMAT)}
      </span>
      <span className="truncate font-mono font-medium tabular-nums">
        {compactNumber(bucket.visitors)}
      </span>
    </div>
  );
}

export function UniqueVisitorChart({
  data,
  strokeWidth = 1.5,
}: {
  data: VisitorBucket[];
  /** The width of the line (px) */
  strokeWidth?: number;
}) {
  const t = useTranslations('Domain');
  const config = useVisitorConfig();
  // The line is hidden until its animation starts: on the first render its
  // path is not measured yet, and recharts would draw it whole for a frame
  // (a dasharray of zeros is a solid stroke) before growing it from nothing
  const [started, setStarted] = useState(!SPARKLINE.animation.isAnimationActive);
  // stable: a new callback would restart the animation on every render
  const onAnimationStart = useCallback(() => setStarted(true), []);
  if (data.length === 0) {
    return (
      <div className={`${SPARKLINE.boxClassName} text-xs text-muted-foreground`}>
        {t('noVisitorData')}
      </div>
    );
  }

  return (
    <ChartContainer
      config={config}
      className={SPARKLINE.className}
      initialDimension={SPARKLINE.initialDimension}
    >
      <LineChart data={data} margin={SPARKLINE.margin}>
        <ChartTooltip {...SPARKLINE.tooltip} cursor={false} content={<SparklineTooltip />} />
        <Line
          dataKey={VISITOR_AREA.dataKey}
          type={SPARKLINE.curve}
          stroke={VISITOR_AREA.stroke}
          strokeWidth={strokeWidth}
          activeDot={SPARKLINE.activeDot}
          dot={false}
          {...SPARKLINE.animation}
          strokeOpacity={started ? 1 : 0}
          onAnimationStart={onAnimationStart}
        />
      </LineChart>
    </ChartContainer>
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
          <CompactNumber value={total} className="block text-2xl font-bold tabular-nums" />
        </div>
        <div>
          <div className="text-sm text-muted-foreground">{t('peakHour')}</div>
          <div className="text-2xl font-bold tabular-nums">
            {peak ? hourOf(peak.time) : '—'}{' '}
            {peak && (
              <span className="text-sm font-normal text-muted-foreground">
                (<CompactNumber value={peak.visitors} />)
              </span>
            )}
          </div>
        </div>
      </div>
      <ChartContainer config={config} className={RESOURCE_CHART.className}>
        <AreaChart data={data} margin={RESOURCE_CHART.margin}>
          <CartesianGrid vertical={false} />
          <XAxis {...RESOURCE_CHART.xAxis} tickFormatter={hourOf} />
          <YAxis {...RESOURCE_CHART.yAxis} tickFormatter={compactNumber} />
          <ChartTooltip
            cursor={false}
            content={
              <CompactTooltipContent
                config={config}
                indicator="line"
                labelFormatter={bucketLabel}
              />
            }
          />
          <Area
            {...VISITOR_AREA}
            type={RESOURCE_CHART.curve}
            strokeWidth={RESOURCE_CHART.strokeWidth}
            fillOpacity={RESOURCE_CHART.fillOpacity}
          />
        </AreaChart>
      </ChartContainer>
    </div>
  );
}
