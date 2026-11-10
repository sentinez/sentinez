import { getTranslations } from 'next-intl/server';

import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { SettingView } from './view';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));
  const t = await getTranslations('ResourceSetting');

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')} />
      <PageLayoutContent>
        <SettingView />
      </PageLayoutContent>
    </PageLayout>
  );
}
