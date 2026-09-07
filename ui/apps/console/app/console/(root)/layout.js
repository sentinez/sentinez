import { RootSidebar, RootSidebarInset } from '@/components/root-sidebar';
export default function RootLayout({ children }) {
    return (<>
      <RootSidebar />
      <RootSidebarInset>{children}</RootSidebarInset>
    </>);
}
