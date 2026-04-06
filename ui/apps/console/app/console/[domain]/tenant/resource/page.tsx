import { ResourceView } from './components/resource-view';

type Props = {
  params: Promise<{
    domain: string;
  }>;
};

export default async function ResourcePage({ params }: Props) {
  const { domain } = await params;

  return <ResourceView domain={domain} />;
}
