'use client';

import {
  Area,
  AreaChart,
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
} from 'recharts';

import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@sentinez/ui/components/chart';
import { Card } from '@sentinez/ui/components/card';

export const description = 'A line chart';

const chartData = [
  { month: '0', desktop: 1 },
  { month: '1', desktop: 2 },
  { month: '2', desktop: 0 },
  { month: '3', desktop: 3 },
  { month: '4', desktop: 1 },
  { month: '5', desktop: 1 },
  { month: '6', desktop: 3 },
  { month: '7', desktop: 2 },
  { month: '8', desktop: 4 },
  { month: '9', desktop: 12 },
  { month: '10', desktop: 1 },
  { month: '11', desktop: 1 },
  { month: '12', desktop: 1 },
  { month: '13', desktop: 0 },
  { month: '14', desktop: 0 },
  { month: '15', desktop: 1 },
  { month: '16', desktop: 10 },
  { month: '17', desktop: 0 },
  { month: '18', desktop: 0 },
  { month: '19', desktop: 0 },
  { month: '20', desktop: 0 },
  { month: '21', desktop: 0 },
  { month: '22', desktop: 0 },
  { month: '23', desktop: 0 },
];

const chartConfig = {
  desktop: {
    label: 'Desktop',
    color: 'var(--chart-1)',
  },
} satisfies ChartConfig;

export function UniqueVisitorChart() {
  return (
    <div className="w-full h-16 overflow-hidden">
      <ChartContainer config={chartConfig} className=" h-16 w-full">
        <AreaChart
          // accessibilityLayer
          data={chartData}
          margin={{ top: 4, bottom: 4, left: 0, right: 0 }}
        >
          <ChartTooltip cursor={false} content={<ChartTooltipContent hideLabel />} />
          <Area
            dataKey="desktop"
            // type="monotone"
            strokeWidth={2}
            dot={false}
            // fill="var(--color-desktop)"
            fillOpacity={0.4}
            // stroke="var(--color-desktop)"
          />
        </AreaChart>
      </ChartContainer>
    </div>
  );
}

// Mock: unique visitors per hour over the last 24h (no analytic API yet)
const hourlyVisitors = [
  42, 35, 28, 21, 18, 16, 24, 51, 88, 124, 142, 156, 171, 163, 149, 158, 177, 196, 214, 187, 152,
  118, 83, 57,
].map((visitors, hour) => ({ hour: `${String(hour).padStart(2, '0')}:00`, visitors }));

const visitorConfig = {
  visitors: {
    label: 'Unique visitors',
    color: 'var(--chart-1)',
  },
} satisfies ChartConfig;

export function ResourceVisitorChart() {
  const total = hourlyVisitors.reduce((sum, d) => sum + d.visitors, 0);
  const peak = hourlyVisitors.reduce((max, d) => (d.visitors > max.visitors ? d : max));

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-8">
        <div>
          <div className="text-sm text-muted-foreground">Total (24h)</div>
          <div className="text-2xl font-bold tabular-nums">{total.toLocaleString()}</div>
        </div>
        <div>
          <div className="text-sm text-muted-foreground">Peak hour</div>
          <div className="text-2xl font-bold tabular-nums">
            {peak.hour}{' '}
            <span className="text-sm font-normal text-muted-foreground">
              ({peak.visitors.toLocaleString()})
            </span>
          </div>
        </div>
      </div>
      <ChartContainer config={visitorConfig} className="h-64 w-full">
        <AreaChart data={hourlyVisitors} margin={{ top: 8, bottom: 0, left: 0, right: 8 }}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="hour" tickLine={false} axisLine={false} tickMargin={8} interval={3} />
          <YAxis tickLine={false} axisLine={false} width={36} />
          <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
          <Area
            dataKey="visitors"
            type="monotone"
            stroke="var(--color-visitors)"
            fill="var(--color-visitors)"
            fillOpacity={0.3}
            strokeWidth={2}
          />
        </AreaChart>
      </ChartContainer>
    </div>
  );
}
