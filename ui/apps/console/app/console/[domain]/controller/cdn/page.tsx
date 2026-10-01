'use client';

import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';

export default function Page() {
  return (
    <PageLayout>
      <PageLayoutHeader
        title="CDN Rule"
        subtitle="Manage active content delivery network rule"
      ></PageLayoutHeader>
      <PageLayoutContent>
        <span></span>
      </PageLayoutContent>
    </PageLayout>
  );
}
