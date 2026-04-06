import PageLayout from '@/components/page-layout';
import { ResourceTable } from '@/app/console/[domain]/tenant/resource/components/table';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 2000));

  return (
    <PageLayout>
      <ResourceTable />
    </PageLayout>
  );
}
