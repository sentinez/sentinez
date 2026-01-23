import PageLayout from '@/components/page-layout';
import ResourceLoading from './components/resources';
import { ResourceTable } from './components/table';

export default async function Page() {
  return (
    // <>
    //   {Array.from({ length: 24 }).map((_, index) => (
    //     <ResourceLoading key={index}/>
    //   ))}
    // </>
    <PageLayout>
      <ResourceTable />
    </PageLayout>
  );
}
