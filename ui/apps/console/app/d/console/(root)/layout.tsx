'use client';

import { ConsoleSidebarInset, RootSidebar } from '@/components/console-sidebar';

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <RootSidebar />
      <ConsoleSidebarInset>{children}</ConsoleSidebarInset>
    </>
  );
}
