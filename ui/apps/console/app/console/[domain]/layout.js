import { DomainSidebar, DomainSidebarInset } from '@/components/domain-sidebar';
export default function DomainLayout({ children }) {
    return (<>
      <DomainSidebar />
      <DomainSidebarInset>{children}</DomainSidebarInset>
    </>);
}
