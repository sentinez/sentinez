'use client';

import { Button } from '@sentinez/ui/components/button';
import IsLoading from '@sentinez/ui/components/common/loading';
import { ArrowLeftRight } from 'lucide-react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';

import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useApi } from '@/hooks/use-api';
import { useTenantId } from '@/hooks/use-tenant-id';
import { getTenant } from '@/lib/api/tenant';

import { DangerCard, OverviewCard, ProfileCard } from './cards';

export default function View() {
  const t = useTranslations('Setting');
  const [tenantId] = useTenantId();
  // null skips the call when no tenant was picked
  const { data, isLoading, refresh } = useApi(getTenant, tenantId ? { id: tenantId } : null);

  // A tenant switched in the sidebar shows the spinner, not the previous one
  const tenant = data?.tenant?.id === tenantId ? data.tenant : undefined;

  if (isLoading && !tenant) return <IsLoading />;

  return (
    <PageLayout>
      <PageLayoutHeader title={t('tenantTitle')} subtitle={t('tenantSubtitle')}>
        <Button size="sm" variant="outline" asChild>
          <Link href="/tenant">
            <ArrowLeftRight />
            {t('switchTenant')}
          </Link>
        </Button>
      </PageLayoutHeader>
      <PageLayoutContent size="md">
        {tenant ? (
          <>
            {/* keyed: another tenant starts from its own name and description */}
            <ProfileCard key={tenant.id} tenant={tenant} onSaved={refresh} />
            <OverviewCard tenant={tenant} />
            <DangerCard tenant={tenant} />
          </>
        ) : (
          <div className="flex flex-col items-center gap-3 rounded-lg border border-dashed p-10 text-center">
            <p className="text-sm text-muted-foreground">
              {tenantId ? t('tenantNotFound') : t('noTenant')}
            </p>
            <Button size="sm" asChild>
              <Link href="/tenant">{t('pickTenant')}</Link>
            </Button>
          </div>
        )}
      </PageLayoutContent>
    </PageLayout>
  );
}
