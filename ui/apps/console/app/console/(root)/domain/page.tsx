import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { ResourceTable } from '@/app/console/components/table';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon } from 'lucide-react';
import Link from 'next/link';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));

  return (
    <PageLayout>
      <PageLayoutHeader title="Domains" subtitle="Manage active resource domains">
        <Button size="sm" asChild>
          <Link href="/console/domain/new">
            <PlusIcon className="w-4 h-4" />
            Create
          </Link>
        </Button>
      </PageLayoutHeader>

      <PageLayoutContent>
        <ResourceTable />
      </PageLayoutContent>
    </PageLayout>
  );
}
