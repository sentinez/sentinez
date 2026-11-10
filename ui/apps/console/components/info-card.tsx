import { ReactNode } from 'react';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { cn } from '@sentinez/ui/lib/utils';

/** Read-only card: a heading (any node, e.g. an avatar row) over InfoRows */
export function InfoCard({
  title,
  description,
  children,
}: Readonly<{ title: ReactNode; description?: ReactNode; children: ReactNode }>) {
  return (
    <Card>
      <CardHeader className="grid-cols-[minmax(0,1fr)] border-b">
        <CardTitle className="leading-normal">{title}</CardTitle>
        {description && <CardDescription>{description}</CardDescription>}
      </CardHeader>
      <CardContent>
        <dl className="divide-y">{children}</dl>
      </CardContent>
    </Card>
  );
}

/**
 * A label / value line of a `<dl>`, e.g. of an InfoCard. `mono` is for ids,
 * `compact` for dense lists such as the detail of an event: a narrow label
 * column beside the value at every width, each line ruled off.
 */
export function InfoRow({
  label,
  mono,
  compact,
  children,
}: Readonly<{
  label: ReactNode;
  mono?: boolean;
  compact?: boolean;
  children: ReactNode;
}>) {
  return (
    <div
      className={
        compact
          ? 'grid grid-cols-[6.5rem_minmax(0,1fr)] gap-3 border-b py-1.5 last:border-0 sm:grid-cols-[9rem_minmax(0,1fr)]'
          : 'grid gap-1 py-3 first:pt-0 last:pb-0 sm:grid-cols-[12rem_minmax(0,1fr)] sm:gap-4'
      }
    >
      <dt className={cn('text-muted-foreground', compact ? 'text-xs font-medium' : 'text-sm')}>
        {label}
      </dt>
      <dd
        className={cn(
          compact ? 'text-xs' : 'text-sm',
          mono ? 'font-mono break-all' : 'break-words',
        )}
      >
        {children}
      </dd>
    </div>
  );
}
