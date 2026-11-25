import type { ComponentProps, CSSProperties } from 'react';
import { type ChartConfig, ChartTooltipContent } from '@sentinez/ui/components/chart';
import { compactNumber } from '@/lib/number';

type Indicator = 'line' | 'dot';

// CompactTooltipContent a ChartTooltipContent with its counts compact
// (1.2K, 3.4M): its rows drawn as the default ones (indicator, the label of
// the series in config, value), as ChartTooltipContent only formats the
// value by replacing the whole row
export function CompactTooltipContent({
  config,
  indicator = 'dot',
  ...props
}: Omit<ComponentProps<typeof ChartTooltipContent>, 'indicator' | 'formatter'> & {
  config: ChartConfig;
  indicator?: Indicator;
}) {
  return (
    <ChartTooltipContent
      {...props}
      indicator={indicator}
      formatter={(value, name, item) => (
        <>
          <div
            className={
              indicator === 'dot'
                ? 'h-2.5 w-2.5 shrink-0 rounded-[2px] bg-(--color-bg)'
                : 'w-1 shrink-0 rounded-[2px] bg-(--color-bg)'
            }
            style={{ '--color-bg': item.payload?.fill ?? item.color } as CSSProperties}
          />
          <div className="flex flex-1 items-center justify-between gap-2 leading-none">
            <span className="text-muted-foreground">
              {config[String(item.dataKey ?? name)]?.label ?? name}
            </span>
            <span className="font-mono font-medium text-foreground tabular-nums">
              {typeof value === 'number' ? compactNumber(value) : String(value)}
            </span>
          </div>
        </>
      )}
    />
  );
}
