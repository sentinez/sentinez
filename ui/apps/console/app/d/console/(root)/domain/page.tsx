import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { ResourceTable } from '@/app/d/console/components/table';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon } from 'lucide-react';
import Link from 'next/link';
import { getTranslations } from 'next-intl/server';
import { pageDelay } from '@/lib/delays';

export default async function Page() {
  const [t, tc] = await pageDelay(
    Promise.all([getTranslations('Domain'), getTranslations('Common')]),
  );

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')}>
        <Button size="sm" asChild>
          <Link href="/d/console/domain/new">
            <PlusIcon className="w-4 h-4" />
            {tc('create')}
          </Link>
        </Button>
      </PageLayoutHeader>

      <PageLayoutContent size="xl">
        <ResourceTable />
      </PageLayoutContent>
    </PageLayout>
  );
}
