'use client';

import type { ReactNode } from 'react';
import { useFormatter, useTranslations } from 'next-intl';

import { Card, CardContent } from '@sentinez/ui/components/card';
import { cn } from '@sentinez/ui/lib/utils';
import type { Resource } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import BadgeStatus from '@/components/badge-status';
import { statusLabel } from '@/lib/type/security';

/** The ids, names, status and timestamps of the resource, as a grid of short fields */
export function ResourceOverview({ resource }: { resource: Resource }) {
  const t = useTranslations('Resource');
  const tc = useTranslations('Common');
  const format = useFormatter();
  // Rendered after the client-side fetch, so the browser's time zone is safe to use
  const formatDate = (date?: Date) =>
    date
      ? format.dateTime(date, {
          dateStyle: 'medium',
          timeStyle: 'short',
          timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        })
      : tc('notAvailable');
  const status = statusLabel(resource.status);
  // The id the edge stamps its events with, when not the resource's
  const eventsId = resource.server?.resourceId;

  return (
    <OverviewCard>
      <OverviewField label={t('name')}>{resource.server?.name}</OverviewField>
      <OverviewField label={t('domain')}>{resource.server?.resourceDomain}</OverviewField>
      <OverviewField label={t('status')}>
        <BadgeStatus status={status} value={status} />
      </OverviewField>
      <OverviewField label={t('updatedAt')}>{formatDate(resource.metadata?.updatedAt)}</OverviewField>
      <OverviewField label={t('id')} mono>
        {resource.id}
      </OverviewField>
      <OverviewField label={t('tenantId')} mono>
        {resource.tenantId}
      </OverviewField>
      {eventsId && eventsId !== resource.id && (
        <OverviewField label={t('eventsResourceId')} mono>
          {eventsId}
        </OverviewField>
      )}
      <OverviewField label={t('createdAt')}>{formatDate(resource.metadata?.createdAt)}</OverviewField>
    </OverviewCard>
  );
}

/** A card of short label / value fields, in a grid of up to four columns */
export function OverviewCard({ children }: { children: ReactNode }) {
  return (
    <Card className="py-4">
      <CardContent className="px-4">
        <dl className="grid grid-cols-1 gap-x-6 gap-y-3 sm:grid-cols-2 lg:grid-cols-4">
          {children}
        </dl>
      </CardContent>
    </Card>
  );
}

/** One field of an OverviewCard; `className` is for a field that spans columns */
export function OverviewField({
  label,
  mono,
  className,
  children,
}: {
  label: string;
  mono?: boolean;
  className?: string;
  children: ReactNode;
}) {
  return (
    <div className={cn('min-w-0', className)}>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className={cn('text-sm', mono ? 'font-mono break-all' : 'break-words')}>{children}</dd>
    </div>
  );
}
