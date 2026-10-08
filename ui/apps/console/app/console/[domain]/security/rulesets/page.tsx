'use client';

import { useTranslations } from 'next-intl';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';

export default function Page() {
  const t = useTranslations('Rulesets');

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')}></PageLayoutHeader>
      <PageLayoutContent>
        <span></span>
      </PageLayoutContent>
    </PageLayout>
  );
}
