import { DomainSidebar, DomainSidebarInset } from '@/components/domain-sidebar';

export default function DomainLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <DomainSidebar />
      <DomainSidebarInset>{children}</DomainSidebarInset>
    </>
  );
}
