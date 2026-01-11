import Nothing from '@sentinez/ui/components/common/nothing';
import { redirect } from 'next/navigation';

export default async function Page() {
  redirect('/console/tenant/resource');
}
