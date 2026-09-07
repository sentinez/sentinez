'use client';
import { Area, AreaChart, } from 'recharts';
import { ChartContainer, ChartTooltip, ChartTooltipContent, } from '@sentinez/ui/components/chart';
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
};
export function UniqueVisitorChart() {
    return (<div className="w-full h-16 overflow-hidden">
      <ChartContainer config={chartConfig} className=" h-16 w-full">
        <AreaChart 
    // accessibilityLayer
    data={chartData} margin={{ top: 4, bottom: 4, left: 0, right: 0 }}>
          <ChartTooltip cursor={false} content={<ChartTooltipContent hideLabel/>}/>
          <Area dataKey="desktop" 
    // type="monotone"
    strokeWidth={2} dot={false} 
    // fill="var(--color-desktop)"
    fillOpacity={0.4}/>
        </AreaChart>
      </ChartContainer>
    </div>);
}
