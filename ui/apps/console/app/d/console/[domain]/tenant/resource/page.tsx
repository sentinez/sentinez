import { ResourceView } from '../../../components/resource-view';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { getTranslations } from 'next-intl/server';

export default async function Page() {
  const t = await getTranslations('Resource');

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')}></PageLayoutHeader>
      <PageLayoutContent>
        <ResourceView />
      </PageLayoutContent>
    </PageLayout>
  );
}
