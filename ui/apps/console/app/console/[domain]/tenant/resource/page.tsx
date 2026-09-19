import Title from '@/components/title';
import { ResourceView } from '../../../components/resource-view';
import PageLayout from '@/components/page-layout';

type Props = {
  params: Promise<{
    domain: string;
  }>;
};

export default async function ResourcePage({ params }: Props) {
  const { domain } = await params;
  return (
    <>
      <Title title="Resource" subtitle="Manage resources for this domain"></Title>
      <PageLayout>
        <ResourceView domain={domain} />;
      </PageLayout>
    </>
  );
}
