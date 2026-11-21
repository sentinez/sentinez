'use client';

import { type ReactNode, useEffect, useState } from 'react';
import { usePathname } from 'next/navigation';
import IsLoading from '@sentinez/ui/components/common/loading';
import { TAB_LOADING_MILLIS } from '@/lib/delays';
import { isUnder } from '@/lib/session';

// A move into a page's own sub page (new/, [id]/) or back out of it is not a
// tab switch: those pages show their own loading.
const related = (a: string, b: string) => isUnder(a, b) || isUnder(b, a);

/**
 * Covers the page with the spinner for TAB_LOADING_MILLIS when the sidebar
 * tab changes (and when the sidebar layout mounts). The page renders and
 * starts fetching underneath, so the wait runs alongside the fetch instead of
 * before it: when the cover lifts, the data or the table skeleton is there.
 */
export function TabLoading({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  // The pathname whose cover has lifted; null until the first one has.
  const [settled, setSettled] = useState<string | null>(null);
  const covering = settled === null || (settled !== pathname && !related(settled, pathname));

  useEffect(() => {
    if (settled === pathname) return;
    const id = setTimeout(() => setSettled(pathname), covering ? TAB_LOADING_MILLIS : 0);
    return () => clearTimeout(id);
  }, [settled, pathname, covering]);

  return (
    <div className="relative flex flex-1 flex-col gap-4">
      <div className="contents" inert={covering}>
        {children}
      </div>
      {covering && (
        <div className="absolute inset-0 z-10 overflow-hidden bg-background">
          <IsLoading />
        </div>
      )}
    </div>
  );
}
