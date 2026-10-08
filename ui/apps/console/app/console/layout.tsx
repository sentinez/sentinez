import { SidebarProvider } from '@sentinez/ui/components/sidebar';
import { Metadata } from 'next';
import { getTranslations } from 'next-intl/server';

import styles from '@/app/console/console.module.scss';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('Metadata');
  return { title: t('consoleTitle'), description: t('consoleDescription') };
}

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarProvider
      // style={{'--sidebar-width': '300px'} as React.CSSProperties}
      className={styles.sidebar_provider}
    >
      {children}
    </SidebarProvider>
  );
}
