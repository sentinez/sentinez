'use client';

import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import IsLoading from '@sentinez/ui/components/common/loading';
import { ArrowLeftRight } from 'lucide-react';
import Link from 'next/link';
import { useFormatter, useTranslations } from 'next-intl';

import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useApi } from '@/hooks/use-api';
import { usePlanLabel } from '@/hooks/use-plan-label';
import { useTenantId } from '@/hooks/use-tenant-id';
import { getTenant } from '@/lib/api/tenant';

import { InfoCard, InfoRow } from '../components';

export default function View() {
  const t = useTranslations('Setting');
  const tt = useTranslations('Tenant');
  const tc = useTranslations('Common');
  const format = useFormatter();
  const planLabel = usePlanLabel();
  const [tenantId] = useTenantId();
  // null skips the call when no tenant was picked
  const { data, isLoading } = useApi(getTenant, tenantId ? { id: tenantId } : null);

  // A tenant switched in the sidebar shows the spinner, not the previous one
  const tenant = data?.tenant?.id === tenantId ? data.tenant : undefined;
  const na = tc('notAvailable');

  const dateOf = (date?: Date) =>
    date ? format.dateTime(date, { dateStyle: 'medium', timeStyle: 'short' }) : na;

  if (isLoading && !tenant) return <IsLoading />;

  return (
    <PageLayout>
      <PageLayoutHeader title={t('tenantTitle')} subtitle={t('tenantSubtitle')}>
        <Button size="sm" variant="outline" asChild>
          <Link href="/tenant">
            <ArrowLeftRight className="w-4 h-4" />
            {t('switchTenant')}
          </Link>
        </Button>
      </PageLayoutHeader>
      <PageLayoutContent>
        <div className="mx-auto w-full max-w-3xl py-4">
          {tenant ? (
            <InfoCard
              title={
                <div className="flex items-center gap-2">
                  <span className="truncate">{tenant.name || tenant.id}</span>
                  <Badge variant="secondary">{planLabel(tenant.plan)}</Badge>
                </div>
              }
              description={tenant.description || undefined}
            >
              <InfoRow label={tt('name')}>{tenant.name || na}</InfoRow>
              <InfoRow label={tt('description')}>{tenant.description || na}</InfoRow>
              <InfoRow label={tt('plan')}>{planLabel(tenant.plan)}</InfoRow>
              <InfoRow label={t('tenantId')} mono>
                {tenant.id}
              </InfoRow>
              <InfoRow label={t('createdAt')}>{dateOf(tenant.metadata?.createdAt)}</InfoRow>
              <InfoRow label={t('updatedAt')}>{dateOf(tenant.metadata?.updatedAt)}</InfoRow>
            </InfoCard>
          ) : (
            <p className="text-muted-foreground py-10 text-center text-sm">
              {tenantId ? t('tenantNotFound') : t('noTenant')}
            </p>
          )}
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
