import { Badge } from '@sentinez/ui/components/badge';
import { STATUS_COLORS, statusClassOf } from './status';

/** An HTTP status, colored by its class like the histogram series */
export function StatusBadge({ status }: { status: number }) {
  const color = STATUS_COLORS[statusClassOf(status)];
  return (
    <Badge
      variant="outline"
      className="font-mono tabular-nums"
      style={{
        color,
        borderColor: color,
        backgroundColor: `color-mix(in oklch, ${color} 12%, transparent)`,
      }}
    >
      {status || '—'}
    </Badge>
  );
}
