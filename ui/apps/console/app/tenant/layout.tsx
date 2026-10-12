import { Metadata } from 'next';
import { getTranslations } from 'next-intl/server';

import { RouteGuard } from '@/components/session-redirect';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('Metadata');
  return { title: t('tenantTitle'), description: t('tenantDescription') };
}

export default function Layout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return <RouteGuard>{children}</RouteGuard>;
}
