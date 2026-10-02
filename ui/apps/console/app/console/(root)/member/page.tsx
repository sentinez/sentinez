import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon } from 'lucide-react';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));

  return (
    <PageLayout>
      <PageLayoutHeader title="Members" subtitle="Manage members of the organization.">
        <Button size="sm">
          <PlusIcon className="w-4 h-4" />
          Invite
        </Button>
      </PageLayoutHeader>
      <PageLayoutContent>
        <span></span>
      </PageLayoutContent>
    </PageLayout>
  );
}
