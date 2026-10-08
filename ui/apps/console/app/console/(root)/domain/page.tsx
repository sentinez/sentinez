import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { ResourceTable } from '@/app/console/components/table';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon } from 'lucide-react';
import Link from 'next/link';
import { getTranslations } from 'next-intl/server';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));
  const t = await getTranslations('Domain');
  const tc = await getTranslations('Common');

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')}>
        <Button size="sm" asChild>
          <Link href="/console/domain/new">
            <PlusIcon className="w-4 h-4" />
            {tc('create')}
          </Link>
        </Button>
      </PageLayoutHeader>

      <PageLayoutContent>
        <ResourceTable />
      </PageLayoutContent>
    </PageLayout>
  );
}
