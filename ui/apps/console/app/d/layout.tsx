import { SidebarProvider } from '@sentinez/ui/components/sidebar';
import { Metadata } from 'next';
import { cookies } from 'next/headers';
import { getTranslations } from 'next-intl/server';

import { RouteGuard } from '@/components/session-redirect';

// Written by SidebarProvider when the sidebar edge is dragged
const SIDEBAR_WIDTH_COOKIE = 'sidebar_width';
const SIDEBAR_DEFAULT_WIDTH = 300;

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('Metadata');
  return { title: t('consoleTitle'), description: t('consoleDescription') };
}

export default async function Layout({ children }: { children: React.ReactNode }) {
  const saved = Number((await cookies()).get(SIDEBAR_WIDTH_COOKIE)?.value);

  return (
    // Drag the sidebar edge (SidebarRail) to resize it.
    <SidebarProvider
      defaultWidth={saved > 0 ? saved : SIDEBAR_DEFAULT_WIDTH}
      minWidth={240}
      maxWidth={560}
    >
      <RouteGuard>{children}</RouteGuard>
    </SidebarProvider>
  );
}
