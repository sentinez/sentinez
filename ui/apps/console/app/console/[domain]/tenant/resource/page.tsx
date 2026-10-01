import { ResourceView } from '../../../components/resource-view';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';

type Props = {
  params: Promise<{
    domain: string;
  }>;
};

export default async function Page({ params }: Props) {
  const { domain } = await params;
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
