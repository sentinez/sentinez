'use client';

import { DomainSidebar, DomainSidebarInset } from '@/components/domain-sidebar';

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <DomainSidebar />
      <DomainSidebarInset>{children}</DomainSidebarInset>
    </>
  );
}
