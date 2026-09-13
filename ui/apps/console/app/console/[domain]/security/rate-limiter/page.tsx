import PageLayout from '@/components/page-layout';
import Title from '@/components/title';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));
  return (
    <PageLayout>
      <Title title="Rate Limiter Rule" subtitle="Manage active security rules."></Title>
    </PageLayout>
  );
}
