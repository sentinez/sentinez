'use client';

import { useEffect, useMemo, useRef, useState, type MouseEvent, type RefObject } from 'react';
import { useFormatter } from 'next-intl';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  DefaultZIndexes,
  XAxis,
  ZIndexLayer,
  YAxis,
  getRelativeCoordinate,
  usePlotArea,
} from 'recharts';

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
import { SECOND_MILLIS } from '@/lib/api/analytic';
import { DATE_TIME_ZONE, SECONDS_TICK, tickFormatOf } from './time-range';
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
  /** A drag over bars: the range it went over, to the second */
  onSelectRange: (from: number, to: number) => void;
}

// Drag the pointer positions a drag went from and to (px from the chart
// left), and the bucket it started on, the one a click selects
type Drag = { startX: number; endX: number; bucket?: number };

// The plot area of the chart (px)
type Plot = { x: number; y: number; width: number; height: number };

// A drag shorter than this (px) is a click
const CLICK_PX = 4;

// The x of a chart mouse event from the chart left (px)
const xOf = (event: MouseEvent<SVGGraphicsElement>) => getRelativeCoordinate(event).relativeX;

/**
 * The layer of a drag in the chart: reports its plot area to plotRef, and
 * shades the dragged span, left to right (px), labeled with its times.
 * Both are drawn over the bars, the label over everything else too.
 */
function DragLayer({
  plotRef,
  span,
  label,
}: {
  plotRef: RefObject<Plot | null>;
  span?: { left: number; right: number };
  label: string;
}) {
  const plot = usePlotArea();
  useEffect(() => {
    plotRef.current = plot ?? null;
  }, [plot, plotRef]);
  if (!plot || !span) return null;
  return (
    <>
      <ZIndexLayer zIndex={DefaultZIndexes.activeBar + 1}>
        <rect
          x={span.left}
          y={plot.y}
          width={span.right - span.left}
          height={plot.height}
          className="pointer-events-none fill-primary/15 stroke-primary/40"
        />
      </ZIndexLayer>
      <ZIndexLayer zIndex={DefaultZIndexes.label + 1}>
        <foreignObject
          x={(span.left + span.right) / 2 - 80}
          y={plot.y}
          width={160}
          height={24}
          className="pointer-events-none overflow-visible"
        >
          <div className="flex justify-center">
            <span className="bg-popover text-popover-foreground rounded border px-1.5 py-0.5 text-xs whitespace-nowrap tabular-nums shadow-sm">
              {label}
            </span>
          </div>
        </foreignObject>
      </ZIndexLayer>
    </>
  );
}

// The state of a chart event, the bucket under the pointer
type ChartState = { activeTooltipIndex?: number | string | null } | null | undefined;

/**
 * Stacked bars of the events per bucket and series (Kibana Discover): a
 * click on a bar selects its bucket, a drag zooms the time range to the
 * time it went over, to the second, however wide the buckets.
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
  const plotRef = useRef<Plot | null>(null);
  const timeAt = (state: ChartState) => data[Number(state?.activeTooltipIndex)]?.time;
  // An x within the plot area (px)
  const clampX = (x: number) => {
    const plot = plotRef.current;
    return plot ? Math.min(Math.max(x, plot.x), plot.x + plot.width) : x;
  };
  // The time at a plot area x, to the second: the buckets share the plot
  // width evenly (a band axis), every one of them present (bucketsOf)
  const timeAtX = (x: number) => {
    const plot = plotRef.current;
    const first = data[0]?.time;
    if (!plot || first === undefined || plot.width <= 0) return undefined;
    const time = first + ((x - plot.x) / plot.width) * data.length * bucketMillis;
    return Math.round(time / SECOND_MILLIS) * SECOND_MILLIS;
  };
  const span = drag && {
    left: clampX(Math.min(drag.startX, drag.endX)),
    right: clampX(Math.max(drag.startX, drag.endX)),
  };
  // The span of a drag, none for a click
  const dragged = span && span.right - span.left >= CLICK_PX ? span : undefined;
  const times = dragged && { from: timeAtX(dragged.left), to: timeAtX(dragged.right) };
  const endDrag = () => {
    setDrag(null);
    if (!drag) return;
    if (!times) {
      if (drag.bucket !== undefined) onSelect(drag.bucket);
      return;
    }
    const { from, to } = times;
    if (from === undefined || to === undefined) return;
    onSelectRange(from, Math.max(to, from + SECOND_MILLIS));
  };
  const config = useMemo(
    () =>
      Object.fromEntries(
        series.map((s) => [s.id, { label: s.label, color: s.color }]),
      ) satisfies ChartConfig,
    [series],
  );
  const secondOf = (time?: number) =>
    time === undefined ? '' : format.dateTime(new Date(time), SECONDS_TICK);

  // The bars, axes and legend: kept while a drag re-renders the chart
  const content = useMemo(() => {
    const tickFormat = tickFormatOf(bucketMillis, resolution);
    const tickOf = (time: number) => format.dateTime(new Date(time), tickFormat);
    const labelOf = (_: unknown, payload: readonly { payload?: Bucket }[]) => {
      const time = payload?.[0]?.payload?.time;
      return time ? format.dateTime(new Date(time), DATE_TIME_ZONE) : '';
    };
    // Only the series with events get bars, so the legend stays short
    const shown = series.filter((s) => data.some((d) => (d[s.id] ?? 0) > 0));
    return (
      <>
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
      </>
    );
  }, [data, series, selected, bucketMillis, resolution, format]);

  return (
    <ChartContainer config={config} className="h-56 w-full select-none">
      <BarChart
        data={data}
        margin={{ top: 8, bottom: 0, left: 0, right: 8 }}
        onMouseDown={(state, event) =>
          setDrag({ startX: xOf(event), endX: xOf(event), bucket: timeAt(state) })
        }
        onMouseMove={(_, event) => {
          const x = xOf(event);
          if (drag && x !== drag.endX) setDrag({ ...drag, endX: x });
        }}
        onMouseUp={endDrag}
        onMouseLeave={() => setDrag(null)}
      >
        {content}
        <DragLayer
          plotRef={plotRef}
          span={dragged}
          label={times ? `${secondOf(times.from)} – ${secondOf(times.to)}` : ''}
        />
      </BarChart>
    </ChartContainer>
  );
}
