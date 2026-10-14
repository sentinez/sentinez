import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon } from 'lucide-react';
import { getTranslations } from 'next-intl/server';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));
  const t = await getTranslations('Members');

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')}>
        <Button size="sm">
          <PlusIcon className="w-4 h-4" />
          {t('invite')}
        </Button>
      </PageLayoutHeader>
      <PageLayoutContent>
        <span></span>
      </PageLayoutContent>
    </PageLayout>
  );
}
