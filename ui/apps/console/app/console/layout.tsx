import { SidebarProvider } from '@sentinez/ui/components/sidebar';
import { Metadata } from 'next';

import styles from '@/app/console/console.module.scss';
import LoadingEffect from '@/components/loading-effect';

export const metadata: Metadata = {
  title: 'Console | Sentinez',
  description: 'Sentinez Console',
};

export default async function ConsoleLayout({ children }: { children: React.ReactNode }) {
  return (
    <LoadingEffect>
      <SidebarProvider
        // style={{'--sidebar-width': '300px'} as React.CSSProperties}
        className={styles.sidebar_provider}
      >
        {children}
      </SidebarProvider>
    </LoadingEffect>
  );
}
