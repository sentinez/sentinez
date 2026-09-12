import Title from '@/components/title';
import { ResourceView } from './components/resource-view';

type Props = {
  params: Promise<{
    domain: string;
  }>;
};

export default async function ResourcePage({ params }: Props) {
  const { domain } = await params;
  return (
    <div className="flex flex-col gap-6 w-full">
      <Title title="Resource" subtitle="Manage resources for this domain"></Title>
      <ResourceView domain={domain} />;
    </div>
  );
}
