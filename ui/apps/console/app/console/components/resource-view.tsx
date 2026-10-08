'use client';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Badge } from '@sentinez/ui/components/badge';
import { Separator } from '@sentinez/ui/components/separator';
import { Skeleton } from '@sentinez/ui/components/skeleton';
import { getResourceByDomain } from '@/lib/api/tenant';
import { useApi } from '@/hooks/use-api';
import { Loader2 } from 'lucide-react';
import IsLoading from '@sentinez/ui/components/common/loading';
import { Button } from '@sentinez/ui/components/button';
import { ResourceVisitorChart } from './chart';
import { useState } from 'react';
import { useFormatter, useTranslations } from 'next-intl';

type Props = {
  domain: string;
};

export function ResourceView({ domain }: Props) {
  const { data: resource, isLoading, error } = useApi(getResourceByDomain, domain);
  const [showJson, setShowJson] = useState(false);
  const t = useTranslations('Resource');
  const tc = useTranslations('Common');
  const format = useFormatter();
  // Rendered after the client-side fetch, so the browser's time zone is safe to use
  const formatDate = (iso?: string) =>
    iso
      ? format.dateTime(new Date(iso), {
          dateStyle: 'medium',
          timeStyle: 'medium',
          timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        })
      : tc('notAvailable');

  if (isLoading) {
    return <IsLoading />;
  }

  if (error || !resource) {
    return (
      <div className="p-8 flex flex-col items-center justify-center text-center space-y-4">
        <div className="text-muted-foreground">{error?.message || t('unavailable')}</div>
        <p className="text-sm text-muted-foreground max-w-xs">
          {t.rich('unavailableHint', {
            domain,
            strong: (chunks) => <strong>{chunks}</strong>,
          })}
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6 w-full max-w-6xl p-4 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{resource.resourceName}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {resource.resourceDomain} &mdash; {resource.id}
          </p>
        </div>
        <Badge
          variant={resource.status === 'STATUS_ACTIVE' ? 'default' : 'secondary'}
          className="text-sm"
        >
          {resource.status}
        </Badge>
      </div>

      <Separator />

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>{t('overview')}</CardTitle>
            <CardDescription>{t('overviewDescription')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">{t('id')}</span>
              <span className="col-span-2 text-sm break-all">{resource.id}</span>
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">{t('name')}</span>
              <span className="col-span-2 text-sm">{resource.resourceName}</span>
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">{t('domain')}</span>
              <span className="col-span-2 text-sm">{resource.resourceDomain}</span>
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">{t('plan')}</span>
              <span className="col-span-2 text-sm">{resource.plan}</span>
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">{t('status')}</span>
              <span className="col-span-2 text-sm">{resource.status}</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('metadata')}</CardTitle>
            <CardDescription>{t('metadataDescription')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-3 gap-4">
              <span className="font-semibold text-sm">{t('createdAt')}</span>
              <span className="col-span-2 text-sm">{formatDate(resource.metadata?.createdAt)}</span>
            </div>
            <div className="grid grid-cols-3 gap-4">
              <span className="font-semibold text-sm">{t('updatedAt')}</span>
              <span className="col-span-2 text-sm">{formatDate(resource.metadata?.updatedAt)}</span>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('uniqueVisitors')}</CardTitle>
          <CardDescription>{t('uniqueVisitorsDescription')}</CardDescription>
        </CardHeader>
        <CardContent>
          <ResourceVisitorChart />
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div className="grid gap-1.5">
            <CardTitle>{t('settings')}</CardTitle>
            <CardDescription>{t('settingsDescription')}</CardDescription>
          </div>
          {resource.resourceSetting && (
            <Button
              variant="outline"
              size="sm"
              aria-expanded={showJson}
              onClick={() => setShowJson((v) => !v)}
            >
              {showJson ? t('hideJson') : t('showJson')}
            </Button>
          )}
        </CardHeader>
        <CardContent>
          {!resource.resourceSetting ? (
            <div className="text-sm text-muted-foreground p-4 border rounded-md">
              {t('noSettings')}
            </div>
          ) : showJson ? (
            <pre className="bg-muted text-foreground p-4 rounded-md overflow-x-auto text-xs min-h-[300px]">
              {JSON.stringify(resource.resourceSetting, null, 2)}
            </pre>
          ) : (
            <div className="text-sm text-muted-foreground">{t('hiddenJson')}</div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
