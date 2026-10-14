'use client';

import { useTranslations } from 'next-intl';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';

export default function Page() {
  const t = useTranslations('Delivery');

  return (
    <PageLayout>
      <PageLayoutHeader title={t('cdnTitle')} subtitle={t('cdnSubtitle')}></PageLayoutHeader>
      <PageLayoutContent>
        <span></span>
      </PageLayoutContent>
    </PageLayout>
  );
}
