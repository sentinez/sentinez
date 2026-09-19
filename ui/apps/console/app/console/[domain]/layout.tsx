'use client';

import { DomainSidebar, DomainSidebarInset } from '@/components/domain-sidebar';
import IsLoading from '@/components/main-loading';
import { useEffect, useState } from 'react';

export default function DomainLayout({ children }: { children: React.ReactNode }) {
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
      <DomainSidebar />
      <DomainSidebarInset>{children}</DomainSidebarInset>
    </>
  );
}
