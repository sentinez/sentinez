'use client';

import { Button } from '@sentinez/ui/components/button';
import { ChevronLeft } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { ReactNode } from 'react';

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
    <div className="sticky top-11.25 z-2 flex w-full max-w-7xl items-center gap-1 border-b bg-background px-1 md:p-2.25">
      <Button
        variant="ghost"
        size="icon"
        className="shrink-0"
        aria-label={t('back')}
        onClick={() => router.back()}
      >
        <ChevronLeft className="w-5 h-5" />
      </Button>
      <div className="flex min-w-0 flex-1 flex-wrap items-center justify-between gap-x-4 gap-y-2 px-1 py-2 sm:p-3">
        <div className="flex min-w-0 flex-col">
          <h1 className="truncate text-base font-semibold tracking-tight sm:text-lg">{title}</h1>
          <p className="line-clamp-2 text-sm text-muted-foreground sm:text-base sm:text-balance">
            {subtitle}
          </p>
        </div>

        {children && <div className="flex flex-wrap items-center gap-2">{children}</div>}
      </div>
    </div>
  );
}

export function PageLayoutContent({ children }: Readonly<{ children: ReactNode }>) {
  return <div className="flex w-full max-w-7xl flex-col gap-6 p-4">{children}</div>;
}
