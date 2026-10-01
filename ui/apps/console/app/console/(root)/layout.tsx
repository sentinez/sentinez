'use client';

import LoadingEffect from '@/components/loading-effect';
import { RootSidebar, RootSidebarInset } from '@/components/root-sidebar';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <RootSidebar />
      <RootSidebarInset>{children}</RootSidebarInset>
    </>
  );
}
