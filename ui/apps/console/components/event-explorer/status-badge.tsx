import type { ReactNode } from 'react';

import { Badge } from '@sentinez/ui/components/badge';
import { cn } from '@sentinez/ui/lib/utils';
import { STATUS_COLORS, statusClassOf } from './status';

/** A mono badge in color, like the histogram series it belongs to */
export function ColoredBadge({
  color,
  className,
  children,
}: {
  color: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <Badge
      variant="outline"
      className={cn('font-mono', className)}
      style={{
        color,
        borderColor: color,
        backgroundColor: `color-mix(in oklch, ${color} 12%, transparent)`,
      }}
    >
      {children}
    </Badge>
  );
}

/** An HTTP status, colored by its class like the histogram series */
export function StatusBadge({ status }: { status: number }) {
  return (
    <ColoredBadge color={STATUS_COLORS[statusClassOf(status)]} className="tabular-nums">
      {status || '—'}
    </ColoredBadge>
  );
}
