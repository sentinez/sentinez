import { Metadata } from 'next';
import { cookies } from 'next/headers';
import { getTranslations } from 'next-intl/server';

import { ResizableSidebarProvider } from '@/components/resizable-sidebar';
import { RouteGuard } from '@/components/session-redirect';
import { NotificationsProvider } from '@/hooks/use-notifications';
import { SIDEBAR_WIDTH_COOKIE } from '@/lib/layout-bars';

const SIDEBAR_DEFAULT_WIDTH = 300;

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('Metadata');
  return { title: t('consoleTitle'), description: t('consoleDescription') };
}

export default async function Layout({ children }: { children: React.ReactNode }) {
  const saved = Number((await cookies()).get(SIDEBAR_WIDTH_COOKIE)?.value);

  return (
    // Drag the sidebar edge (ResizableSidebarRail) to resize it.
    <ResizableSidebarProvider
      defaultWidth={saved > 0 ? saved : SIDEBAR_DEFAULT_WIDTH}
      minWidth={240}
      maxWidth={560}
    >
      <RouteGuard>
        {/* above the section layouts: the stream stays open between them */}
        <NotificationsProvider>{children}</NotificationsProvider>
      </RouteGuard>
    </ResizableSidebarProvider>
  );
}
