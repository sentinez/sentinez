import { redirect } from 'next/navigation';
export default async function Page({ params }) {
    const { domain } = await params;
    redirect(`/console/${domain}/tenant/resource`);
}
