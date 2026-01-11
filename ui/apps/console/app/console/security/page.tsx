import { redirect } from 'next/navigation';

export default async function Page() {
  redirect('/console/security/logs');
}
