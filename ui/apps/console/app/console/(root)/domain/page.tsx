import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { ResourceTable } from '@/app/console/components/table';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon } from 'lucide-react';

export default async function Page() {
  return (
    <PageLayout>
      <PageLayoutHeader title="Domains" subtitle="Manage active resource domains">
        <Button size="sm">
          <PlusIcon className="w-4 h-4" />
          Create
        </Button>
      </PageLayoutHeader>

      <PageLayoutContent>
        <ResourceTable />
      </PageLayoutContent>
    </PageLayout>
  );
}
