import type { ChartConfig } from '@sentinez/ui/components/chart';
import type { LineProps, XAxisProps, YAxisProps } from 'recharts';

// The line between the points of a chart: 'linear' straight with sharp
// corners, 'monotone' smooth, 'step' in steps
type Curve = NonNullable<LineProps['type']>;

// The chart palette, the --chart-N tokens of packages/ui globals.css: the
// shades of the brand blue. Series of a meaning of their own
// (HTTP status, rule action) take STATUS_COLORS (event-explorer/status).
export const CHART_COLORS = [
  'var(--chart-1)',
  'var(--chart-2)',
  'var(--chart-3)',
  'var(--chart-4)',
  'var(--chart-5)',
] as const;

// The margin of a chart with axes: room on the right for the last tick
export const CHART_MARGIN = { top: 8, bottom: 0, left: 0, right: 8 } as const;

// A sparkline (no axes) in a table row: a line only, 40px tall (h-10), in a
// column of columnSize (its cells padded p-2: width px wide). Its first
// render takes initialDimension before the container is measured: the
// default (320x200) would stretch the row until then.
const SPARKLINE_COLUMN = 200;
const SPARKLINE_WIDTH = SPARKLINE_COLUMN - 16;
export const SPARKLINE = {
  columnSize: SPARKLINE_COLUMN,
  className: 'aspect-auto h-10 w-full',
  // the box of the chart cell, and of its "no data" text: the chart height
  boxClassName: 'flex h-10 w-full items-center',
  // the placeholder of the chart while it loads: the same box
  skeletonClassName: 'h-10 w-full',
  initialDimension: { width: SPARKLINE_WIDTH, height: 40 },
  margin: { top: 4, bottom: 4, left: 2, right: 2 },
  curve: 'linear' satisfies Curve,
  activeDot: { r: 3 },
  // The line draws itself left to right when the chart shows (and again
  // when its data changes); isAnimationActive: false turns it off
  animation: {
    isAnimationActive: true,
    animationBegin: 0,
    animationDuration: 800,
    animationEasing: 'ease-out',
  },
  // The tooltip floats above the chart, as wide as it (tooltipWidth) and on one line (the
  // hour left, the count right, no series name: it would not fit), so the
  // hovered point stays in sight
  tooltip: {
    position: { x: 0, y: -32 },
    allowEscapeViewBox: { x: true, y: true },
    offset: 0,
  },
  tooltipWidth: SPARKLINE_WIDTH,
  tooltipClassName:
    'flex items-center justify-between gap-3 rounded-lg border border-border/50 bg-background px-2 py-1 text-xs whitespace-nowrap shadow-md',
} as const;

// The time axis of a chart, without line or ticks: its labels left out as
// they would overlap, at least minTickGap (px) apart whatever the width of
// the chart and the number of buckets
export const X_AXIS = {
  dataKey: 'time',
  tickLine: false,
  axisLine: false,
  tickMargin: 8,
  minTickGap: 32,
} as const satisfies Partial<XAxisProps>;

// The count axis of a chart: whole numbers only
export const Y_AXIS = {
  tickLine: false,
  axisLine: false,
  width: 40,
  allowDecimals: false,
} as const satisfies Partial<YAxisProps>;

// The unique visitors chart of a resource (tenant/resource page): an area
// per hour, its axes, and the skeleton of its card while it loads: as tall
// as the card content, the stats row (~4.5rem with its gap) and the chart
export const RESOURCE_CHART = {
  className: 'aspect-auto h-64 w-full',
  skeletonClassName: 'h-80 w-full',
  margin: CHART_MARGIN,
  // the first and last labels kept
  xAxis: { ...X_AXIS, interval: 'preserveStartEnd' },
  yAxis: { ...Y_AXIS, width: 36 },
  curve: 'linear' satisfies Curve,
  strokeWidth: 4,
  fillOpacity: 0,
} as const;

// The tooltip cursor of a bar chart: a shade over the hovered bucket
export const BAR_CURSOR = { fill: 'var(--muted)', opacity: 0.6 } as const;

// The opacity of the bars of the buckets not selected
export const DIMMED_OPACITY = 0.35;

// A series of a chart: its key in the data, label and color
export type ChartSeries = { id: string; label: string; color: string };

// The chart config of series, keyed by their id: each color is then
// var(--color-<id>) inside the ChartContainer
export function seriesConfig(series: readonly ChartSeries[]): ChartConfig {
  return Object.fromEntries(series.map((s) => [s.id, { label: s.label, color: s.color }]));
}
