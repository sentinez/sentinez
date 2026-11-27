'use client';

import { Button } from '@sentinez/ui/components/button';
import { ChevronLeft } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { ReactNode } from 'react';
import { cn } from '@sentinez/ui/lib/utils';
import { BreadcrumbBar } from './breadcrumb-bar';
import { PAGE_HEADER_MIN_HEIGHT } from '@/lib/layout-bars';

export function PageLayout({ children }: Readonly<{ children: ReactNode }>) {
  return <>{children}</>;
}

export interface PageLayoutHeaderProps {
  children?: ReactNode;
  title: string;
  subtitle: string;
}

export function PageLayoutHeader({ children, title, subtitle }: PageLayoutHeaderProps) {
  const router = useRouter();
  const t = useTranslations('Common');

  return (
    // stuck under the top bar of ConsoleSidebarInset while it shows, at the
    // top of the window when it hides (0 outside of it)
    <div className="sticky top-[var(--console-preview-offset,0px)] z-1 bg-background transition-[top] duration-150 ease-out">
      <BreadcrumbBar />
      <div
        className={cn(
          'flex w-full max-w-7xl items-center gap-1 border-b px-1 md:px-2',
          PAGE_HEADER_MIN_HEIGHT,
        )}
      >
        <Button
          variant="ghost"
          size="icon"
          className="shrink-0"
          aria-label={t('back')}
          onClick={() => router.back()}
        >
          <ChevronLeft className="w-5 h-5" />
        </Button>
        <div className="flex min-w-0 flex-1 flex-wrap items-center justify-between gap-x-4 gap-y-2 px-1 py-2 sm:p-2">
          <div className="flex min-w-0 flex-col">
            <h1 className="truncate text-base font-semibold tracking-tight">{title}</h1>
            <p className="line-clamp-2 text-sm text-muted-foreground sm:text-balance">{subtitle}</p>
          </div>

          {children && <div className="flex flex-wrap items-center gap-2">{children}</div>}
        </div>
      </div>
    </div>
  );
}

/** The widths of `PageLayoutContent`, from large to small. */
export type PageLayoutContentSize = 'xl' | 'lg' | 'md' | 'sm';

const contentSizes: Record<PageLayoutContentSize, string> = {
  xl: 'max-w-7xl',
  lg: 'max-w-5xl',
  md: 'max-w-3xl',
  sm: 'max-w-xl',
};

/**
 * The body of a page, centred under the header (`max-w-7xl`), padded and
 * spacing its children; the whole width when the page is narrower. `size`
 * sets its width, from large to small: `xl` the whole width of the header
 * (list pages, event logs and their histogram), `lg` the default (details),
 * `md`, and `sm` for the create and update forms. Put the content in it directly: no wrapper with
 * its own `max-w-*`, `mx-auto` or vertical padding, they break the alignment.
 */
export function PageLayoutContent({
  children,
  size = 'lg',
}: Readonly<{ children: ReactNode; size?: PageLayoutContentSize }>) {
  return (
    <div className="w-full max-w-7xl">
      <div className={cn('mx-auto flex w-full flex-col gap-6 p-4', contentSizes[size])}>
        {children}
      </div>
    </div>
  );
}

/**
 * The row of controls (filter, refresh, columns) above a list table, spaced
 * from it; its top is the one of `PageLayoutContent`.
 */
export function PageLayoutToolbar({
  children,
  className,
}: Readonly<{ children: ReactNode; className?: string }>) {
  return <div className={cn('flex items-center gap-2 pb-4', className)}>{children}</div>;
}
