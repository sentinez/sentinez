import { AppSidebar, AppSidebarInset } from '@/components/app-sidebar';
import { SidebarProvider } from '@sentinez/ui/components/sidebar';
import {Metadata} from 'next';

import styles from '@/app/console/console.module.scss';


export const metadata: Metadata = {
  title: 'Console | Sentinez',
  description: 'Sentinez Console',
};

export default async function ConsoleLayout({ children }: { children: React.ReactNode }) {
  await new Promise((resolve) => setTimeout(resolve, 1000));

  return (
    <SidebarProvider
      // style={{'--sidebar-width': '300px'} as React.CSSProperties}
      className={styles.sidebar_provider}
    >
      <AppSidebar />
      <AppSidebarInset>{children}</AppSidebarInset>
    </SidebarProvider>
  );
}
