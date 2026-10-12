'use client';

import { ReactNode, useEffect, useState } from 'react';

import { usePathname, useRouter } from 'next/navigation';
import IsLoading from '@sentinez/ui/components/common/loading';

import { redirectFor } from '@/lib/session';

/**
 * Applies the redirects of redirectFor (see lib/session.ts). It sits in the
 * root layout, outside LoadingEffect, so the navigation starts while the
 * splash is still up instead of after the page was painted.
 */
export function SessionRedirect() {
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    const target = redirectFor(pathname);
    if (target) router.replace(target);
  }, [pathname, router]);

  return null;
}

/**
 * Renders a page only for a visitor allowed to see it. One who is on the way
 * out (see SessionRedirect) gets a spinner, never a glimpse of the page.
 */
export function RouteGuard({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  // unknown until mounted: localStorage is not readable while rendering
  const [allowed, setAllowed] = useState<boolean>();

  useEffect(() => {
    setAllowed(redirectFor(pathname) === null);
  }, [pathname]);

  if (allowed === undefined) return null;
  if (!allowed) return <IsLoading />;

  return <>{children}</>;
}
