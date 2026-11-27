'use client';

import { Fragment, useMemo } from 'react';
import { usePathname } from 'next/navigation';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbSeparator,
} from '@sentinez/ui/components/breadcrumb';
import { Separator } from '@sentinez/ui/components/separator';
import { SidebarTrigger } from '@sentinez/ui/components/sidebar';
import { cn } from '@sentinez/ui/lib/utils';
import { BAR_HEIGHT } from '@/lib/layout-bars';

/** The bar with the sidebar trigger and the breadcrumb of the current path. */
export function BreadcrumbBar() {
  const pathname = usePathname();
  const segments = useMemo(() => pathname.split('/').filter((path) => path !== ''), [pathname]);

  return (
    // the padding of PageLayoutHeader, and the trigger centred in a slot as
    // wide as its back button (size-9): the two icons line up
    <div className={cn('flex shrink-0 items-center gap-2 border-b px-1 md:px-2', BAR_HEIGHT)}>
      <div className="grid w-9 shrink-0 place-items-center">
        <SidebarTrigger className="cursor-pointer" />
      </div>
      <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
      <Breadcrumb>
        <BreadcrumbList>
          {segments.map((path, index) => (
            <Fragment key={index}>
              <BreadcrumbItem
                // a phone shows only the current page
                className={cn(
                  'text-xs',
                  index === segments.length - 1
                    ? 'truncate text-black dark:text-white'
                    : 'hidden md:block',
                )}
              >
                {path === 'd' ? 'dashboard' : path}
              </BreadcrumbItem>
              {index < segments.length - 1 && (
                <BreadcrumbSeparator className="hidden md:block text-xs" />
              )}
            </Fragment>
          ))}
        </BreadcrumbList>
      </Breadcrumb>
    </div>
  );
}
