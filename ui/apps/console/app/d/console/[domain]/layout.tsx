'use client';

import { ConsoleSidebarInset, DomainSidebar } from '@/components/console-sidebar';

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <DomainSidebar />
      <ConsoleSidebarInset>{children}</ConsoleSidebarInset>
    </>
  );
}
