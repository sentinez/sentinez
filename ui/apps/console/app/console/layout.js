import { SidebarProvider } from '@sentinez/ui/components/sidebar';
import styles from '@/app/console/console.module.scss';
export const metadata = {
    title: 'Console | Sentinez',
    description: 'Sentinez Console',
};
export default async function ConsoleLayout({ children }) {
    return (<SidebarProvider 
    // style={{'--sidebar-width': '300px'} as React.CSSProperties}
    className={styles.sidebar_provider}>
      {children}
    </SidebarProvider>);
}
