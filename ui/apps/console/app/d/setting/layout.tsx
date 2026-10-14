'use client';

import { RootSidebar, RootSidebarInset } from '@/components/root-sidebar';

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <RootSidebar />
      <RootSidebarInset>{children}</RootSidebarInset>
    </>
  );
}
