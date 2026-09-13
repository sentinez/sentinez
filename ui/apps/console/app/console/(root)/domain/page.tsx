import PageLayout from '@/components/page-layout';
import { ResourceTable } from '@/app/console/[domain]/tenant/resource/components/table';
import Title from '@/components/title';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon } from 'lucide-react';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));

  return (
    <PageLayout>
      <Title title="Domains" subtitle="Manage active resource domains">
        <Button size="sm">
          <PlusIcon className="w-4 h-4" />
          Create
        </Button>
      </Title>

      <ResourceTable />
    </PageLayout>
  );
}
