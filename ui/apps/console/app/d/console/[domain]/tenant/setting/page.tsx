import { getTranslations } from 'next-intl/server';

import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { SettingView } from './view';

type Props = {
  params: Promise<{
    domain: string;
  }>;
};

export default async function Page({ params }: Props) {
  const { domain } = await params;
  await new Promise((resolve) => setTimeout(resolve, 1500));
  const t = await getTranslations('ResourceSetting');

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')} />
      <PageLayoutContent>
        <SettingView domain={domain} />
      </PageLayoutContent>
    </PageLayout>
  );
}
