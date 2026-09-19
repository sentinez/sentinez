import PageLayout from '@/components/page-layout';
import Title from '@/components/title';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon } from 'lucide-react';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));

  return (
    <>
      <Title title="Members" subtitle="Manage members of the organization.">
        <Button size="sm">
          <PlusIcon className="w-4 h-4" />
          Invite
        </Button>
      </Title>
      <PageLayout>
        <span></span>
      </PageLayout>
    </>
  );
}
