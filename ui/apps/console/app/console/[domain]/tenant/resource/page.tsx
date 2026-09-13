import Title from '@/components/title';
import { ResourceView } from './components/resource-view';
import PageLayout from '@/components/page-layout';

type Props = {
  params: Promise<{
    domain: string;
  }>;
};

export default async function ResourcePage({ params }: Props) {
  const { domain } = await params;
  return (
    <PageLayout>
      <Title title="Resource" subtitle="Manage resources for this domain"></Title>
      <ResourceView domain={domain} />;
    </PageLayout>
  );
}
