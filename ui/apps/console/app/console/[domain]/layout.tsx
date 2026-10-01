'use client';

import { DomainSidebar, DomainSidebarInset } from '@/components/domain-sidebar';
import LoadingEffect from '@/components/loading-effect';

export default function DomainLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <DomainSidebar />
      <DomainSidebarInset>{children}</DomainSidebarInset>
    </>
  );
}
