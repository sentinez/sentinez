import { cn } from '@sentinez/ui/lib/utils';

// The bars of the Sentinéz mark (apps/console/public/assets/sz.png), x, y and
// height; the viewBox crops its margin so the bars line up with the text
const BARS: [number, number, number][] = [
  [120, 120, 800],
  [240, 280, 600],
  [360, 360, 480],
  [480, 120, 800],
  [600, 360, 480],
  [720, 280, 600],
  [840, 120, 800],
];

/** The Sentinéz mark, in the brand color (lighter in dark mode) */
export function LogoMark({ className }: { className?: string }) {
  return (
    <svg viewBox="120 120 800 800" aria-hidden className={cn('text-brand', className)}>
      {BARS.map(([x, y, h]) => (
        <rect key={x} x={x} y={y} width={80} height={h} fill="currentColor" />
      ))}
    </svg>
  );
}

/** The mark, SENTINÉZ in Phudu, and the app name */
export function Logo() {
  return (
    <span className="flex items-center gap-1.5">
      <LogoMark className="size-5" />
      <span className="font-brand text-xl font-bold text-brand">SENTINÉZ</span>
      <span className="ml-1 border-l pl-2.5 text-sm font-medium text-muted-foreground">recept</span>
    </span>
  );
}
