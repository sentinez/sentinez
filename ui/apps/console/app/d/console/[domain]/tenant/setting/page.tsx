import { getTranslations } from 'next-intl/server';

import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { SettingView } from './view';
import { pageDelay } from '@/lib/delays';

export default async function Page() {
  const t = await pageDelay(getTranslations('ResourceSetting'));

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')} />
      <PageLayoutContent>
        <SettingView />
      </PageLayoutContent>
    </PageLayout>
  );
}
