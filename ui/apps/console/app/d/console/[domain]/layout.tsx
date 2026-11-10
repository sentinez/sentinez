'use client';

import { ConsoleSidebarInset, DomainSidebar } from '@/components/console-sidebar';
import { useDomain } from '@/hooks/use-domain';
import { ResourceProvider } from './resource-context';

export default function Layout({ children }: { children: React.ReactNode }) {
  const domain = useDomain();

  return (
    <ResourceProvider key={domain} domain={domain}>
      <DomainSidebar />
      <ConsoleSidebarInset>{children}</ConsoleSidebarInset>
    </ResourceProvider>
  );
}
