import { redirect } from 'next/navigation';

export default async function Page({ params }: { params: Promise<{ domain: string }> }) {
  const { domain } = await params;
  redirect(`/console/${domain}/tenant/resource`);
}
