'use client';

import { useMemo, useState } from 'react';
import { useFormatter } from 'next-intl';
import { Bar, BarChart, CartesianGrid, Cell, ReferenceArea, XAxis, YAxis } from 'recharts';

import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@sentinez/ui/components/chart';
import type { HistogramPoint } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import type { HistogramResolution } from '@sentinez/proto/sentinez/types/v1/known';
import { DATE_TIME_ZONE, tickFormatOf } from './time-range';
import type { SeriesDef } from './types';

// Bucket the events of the bucket starting at time (millis), per series
export type Bucket = { time: number; total: number; [series: string]: number };

// The buckets of the points of a histogram, oldest first: the totals of
// every resource and key summed per series
export function bucketsOf(points: HistogramPoint[], seriesOf: (key: number) => string): Bucket[] {
  const buckets = new Map<number, Bucket>();
  for (const p of points) {
    const time = p.time?.getTime() ?? 0;
    const bucket = buckets.get(time) ?? { time, total: 0 };
    // an empty bucket has a point of total 0
    if (p.total > 0) {
      const series = seriesOf(p.key);
      bucket[series] = (bucket[series] ?? 0) + p.total;
      bucket.total += p.total;
    }
    buckets.set(time, bucket);
  }
  return [...buckets.values()].sort((a, b) => a.time - b.time);
}

interface Props {
  data: Bucket[];
  series: SeriesDef[];
  resolution: HistogramResolution;
  /** The width of a bucket, in millis */
  bucketMillis: number;
  /** The start of the selected bucket, its bar highlighted */
  selected?: number;
  /** A click on a bar: the start of its bucket */
  onSelect: (time: number) => void;
  /** A drag over bars: from the start of the first bucket to the end of the last */
  onSelectRange: (from: number, to: number) => void;
}

// Drag the buckets a drag went over, from the one it started on (millis)
type Drag = { start: number; end: number };

// The state of a chart event, the bucket under the pointer
type ChartState = { activeTooltipIndex?: number | string | null } | null | undefined;

/**
 * Stacked bars of the events per bucket and series (Kibana Discover): a
 * click on a bar selects its bucket, a drag over bars zooms the time range
 * to them.
 */
export function EventHistogram({
  data,
  series,
  resolution,
  bucketMillis,
  selected,
  onSelect,
  onSelectRange,
}: Props) {
  const format = useFormatter();
  const [drag, setDrag] = useState<Drag | null>(null);
  const timeAt = (state: ChartState) => data[Number(state?.activeTooltipIndex)]?.time;
  const dragged = drag && {
    from: Math.min(drag.start, drag.end),
    to: Math.max(drag.start, drag.end),
  };
  const endDrag = () => {
    setDrag(null);
    if (!dragged) return;
    if (dragged.from === dragged.to) onSelect(dragged.from);
    else onSelectRange(dragged.from, dragged.to + bucketMillis);
  };
  const config = useMemo(
    () =>
      Object.fromEntries(
        series.map((s) => [s.id, { label: s.label, color: s.color }]),
      ) satisfies ChartConfig,
    [series],
  );
  const tickFormat = tickFormatOf(bucketMillis, resolution);
  const tickOf = (time: number) => format.dateTime(new Date(time), tickFormat);
  const labelOf = (_: unknown, payload: readonly { payload?: Bucket }[]) => {
    const time = payload?.[0]?.payload?.time;
    return time ? format.dateTime(new Date(time), DATE_TIME_ZONE) : '';
  };
  // Only the series with events get bars, so the legend stays short
  const shown = series.filter((s) => data.some((d) => (d[s.id] ?? 0) > 0));

  return (
    <ChartContainer config={config} className="h-56 w-full select-none">
      <BarChart
        data={data}
        margin={{ top: 8, bottom: 0, left: 0, right: 8 }}
        onMouseDown={(state) => {
          const time = timeAt(state);
          if (time !== undefined) setDrag({ start: time, end: time });
        }}
        onMouseMove={(state) => {
          const time = timeAt(state);
          if (drag && time !== undefined && time !== drag.end) setDrag({ ...drag, end: time });
        }}
        onMouseUp={endDrag}
        onMouseLeave={() => setDrag(null)}
      >
        <CartesianGrid vertical={false} />
        <XAxis
          dataKey="time"
          tickFormatter={tickOf}
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          minTickGap={32}
        />
        <YAxis tickLine={false} axisLine={false} width={40} allowDecimals={false} />
        <ChartTooltip
          cursor={{ fill: 'var(--muted)', opacity: 0.6 }}
          content={<ChartTooltipContent labelFormatter={labelOf} />}
        />
        {shown.length > 1 && <ChartLegend content={<ChartLegendContent />} />}
        {dragged && dragged.from !== dragged.to && (
          <ReferenceArea
            x1={dragged.from}
            x2={dragged.to}
            fill="var(--primary)"
            fillOpacity={0.15}
            stroke="var(--primary)"
            strokeOpacity={0.4}
            ifOverflow="visible"
          />
        )}
        {shown.map((s, i) => (
          <Bar
            key={s.id}
            dataKey={s.id}
            stackId="events"
            fill={`var(--color-${s.id})`}
            radius={i === shown.length - 1 ? [2, 2, 0, 0] : 0}
            className="cursor-pointer"
            isAnimationActive={false}
          >
            {data.map((d) => (
              <Cell
                key={d.time}
                fillOpacity={selected === undefined || selected === d.time ? 1 : 0.35}
              />
            ))}
          </Bar>
        ))}
      </BarChart>
    </ChartContainer>
  );
}
