import { ResourceView } from '../../../components/resource-view';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';

type Props = {
  params: Promise<{
    domain: string;
  }>;
};

export default async function Page({ params }: Props) {
  const { domain } = await params;
  await new Promise((resolve) => setTimeout(resolve, 1500));

  return (
    <PageLayout>
      <PageLayoutHeader
        title="Resource"
        subtitle="Manage resources for this domain"
      ></PageLayoutHeader>
      <PageLayoutContent>
        <ResourceView domain={domain} />
      </PageLayoutContent>
    </PageLayout>
  );
}
