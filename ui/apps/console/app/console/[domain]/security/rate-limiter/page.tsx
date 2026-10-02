import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 15000));
  return (
    <PageLayout>
      <PageLayoutHeader
        title="Rate Limiter Rule"
        subtitle="Manage active security rules."
      ></PageLayoutHeader>

      <PageLayoutContent>
        <span></span>
      </PageLayoutContent>
    </PageLayout>
  );
}
