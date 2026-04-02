import PageLayout from '@/components/page-layout';
import { ResourceTable } from '@/app/console/[domain]/tenant/resource/components/table';

export default function Page() {
  return (
    <PageLayout>
      <ResourceTable />
    </PageLayout>
  );
}
