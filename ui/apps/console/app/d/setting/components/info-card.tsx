import { ReactNode } from 'react';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';

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

/** A label / value line of an InfoCard; `mono` is for ids */
export function InfoRow({
  label,
  mono,
  children,
}: Readonly<{ label: string; mono?: boolean; children: ReactNode }>) {
  return (
    <div className="grid gap-1 py-3 first:pt-0 last:pb-0 sm:grid-cols-[12rem_minmax(0,1fr)] sm:gap-4">
      <dt className="text-muted-foreground text-sm">{label}</dt>
      <dd className={mono ? 'font-mono text-sm break-all' : 'text-sm break-words'}>{children}</dd>
    </div>
  );
}
