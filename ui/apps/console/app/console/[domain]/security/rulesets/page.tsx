'use client';

import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';

export default function Page() {
  return (
    <PageLayout>
      <PageLayoutHeader
        title="Security Rulesets"
        subtitle="Manage active owasp core rulesets."
      ></PageLayoutHeader>
      <PageLayoutContent>
        <span></span>
      </PageLayoutContent>
    </PageLayout>
  );
}
