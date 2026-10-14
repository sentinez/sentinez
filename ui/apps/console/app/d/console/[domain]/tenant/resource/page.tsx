import { ResourceView } from '../../../components/resource-view';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { getTranslations } from 'next-intl/server';

type Props = {
  params: Promise<{
    domain: string;
  }>;
};

export default async function Page({ params }: Props) {
  const { domain } = await params;
  await new Promise((resolve) => setTimeout(resolve, 1500));
  const t = await getTranslations('Resource');

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')}></PageLayoutHeader>
      <PageLayoutContent>
        <ResourceView domain={domain} />
      </PageLayoutContent>
    </PageLayout>
  );
}
