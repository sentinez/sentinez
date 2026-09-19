'use client';

import IsLoading from '@/components/main-loading';
import { RootSidebar, RootSidebarInset } from '@/components/root-sidebar';
import { useEffect, useState } from 'react';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const run = async () => {
      await new Promise((resolve) => setTimeout(resolve, 1500));

      setLoading(false);
    };

    run();
  }, []);

  if (loading) return <IsLoading timeout={1000} />;

  return (
    <>
      <RootSidebar />
      <RootSidebarInset>{children}</RootSidebarInset>
    </>
  );
}
