'use client';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Separator } from '@sentinez/ui/components/separator';
import { getResourceByDomain } from '@/lib/api/tenant';
import { statusToJSON } from '@sentinez/proto/sentinez/types/v1/known';
import { useApi } from '@/hooks/use-api';
import IsLoading from '@sentinez/ui/components/common/loading';
import { ResourceUnavailable } from './resource-unavailable';
import { ResourceVisitors } from './resource-visitors';
import { useFormatter, useTranslations } from 'next-intl';
import BadgeStatus from '@/components/badge-status';
import { statusLabel } from '@/lib/type/security';

type Props = {
  domain: string;
};

export function ResourceView({ domain }: Props) {
  const { data, isLoading, error } = useApi(getResourceByDomain, { resourceDomain: domain });
  const resource = data?.resource;
  const t = useTranslations('Resource');
  const tc = useTranslations('Common');
  const format = useFormatter();
  // Rendered after the client-side fetch, so the browser's time zone is safe to use
  const formatDate = (date?: Date) =>
    date
      ? format.dateTime(date, {
          dateStyle: 'medium',
          timeStyle: 'medium',
          timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        })
      : tc('notAvailable');

  if (isLoading) {
    return <IsLoading />;
  }

  if (error || !resource) {
    return <ResourceUnavailable domain={domain} error={error} />;
  }

  return (
    <div className="flex flex-col gap-6 w-full max-w-6xl p-4 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{resource.server?.name}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {resource.server?.resourceDomain} &mdash; {resource.id}
          </p>
        </div>
        <BadgeStatus status={statusLabel(resource.status)} value={statusLabel(resource.status)} />
      </div>

      <Separator />

      <Card>
        <CardHeader>
          <CardTitle>{t('overview')}</CardTitle>
          <CardDescription>{t('overviewDescription')}</CardDescription>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-4">
            <Field label={t('id')}>{resource.id}</Field>
            <Field label={t('tenantId')}>{resource.tenantId}</Field>
            <Field label={t('name')}>{resource.server?.name}</Field>
            <Field label={t('domain')}>{resource.server?.resourceDomain}</Field>
            <Field label={t('status')}>{statusToJSON(resource.status)}</Field>
            {/* The id the edge stamps its events with, when not the resource's */}
            {resource.server?.resourceId && resource.server.resourceId !== resource.id && (
              <Field label={t('eventsResourceId')}>{resource.server.resourceId}</Field>
            )}
            <Field label={t('createdAt')}>{formatDate(resource.metadata?.createdAt)}</Field>
            <Field label={t('updatedAt')}>{formatDate(resource.metadata?.updatedAt)}</Field>
          </dl>
        </CardContent>
      </Card>

      <ResourceVisitors resourceId={resource.id} />
    </div>
  );
}

// A label and its value of the overview
function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-3 items-center gap-4">
      <dt className="font-semibold text-sm">{label}</dt>
      <dd className="col-span-2 text-sm break-all">{children}</dd>
    </div>
  );
}
