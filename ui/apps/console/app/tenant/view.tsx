'use client';

import { useState } from 'react';
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from '@sentinez/ui/components/carousel';
import IsLoading from '@sentinez/ui/components/common/loading';
import { Input } from '@sentinez/ui/components/input';
import { Search } from 'lucide-react';
import Image from 'next/image';
import { useTranslations } from 'next-intl';

import Language from '@/components/language';
import { RefreshButton } from '@/components/refresh-button';
import { UserMenu } from '@/components/user-menu';
import { useApi } from '@/hooks/use-api';
import { useDebounced } from '@/hooks/use-debounced';
import { listTenants } from '@/lib/api/tenant';

import { Tenant, TeamJoin } from './components/card';

// Tenant cards visible at once; more than this scrolls
const VISIBLE_CARDS = 3;
// Card width (300px) plus the 16px gutter of a CarouselItem
const SLIDE_WIDTH = 316;
const SLIDE_GUTTER = 16;
// Below Tailwind's `sm` the carousel is off and the cards stack vertically
const CAROUSEL_OPTS = {
  align: 'start',
  active: false,
  breakpoints: { '(min-width: 640px)': { active: true } },
} as const;

export default function View() {
  const t = useTranslations('Tenant');
  const tc = useTranslations('Common');
  const [search, setSearch] = useState('');
  const name = useDebounced(search.trim());
  const { data, isLoading, refresh } = useApi(listTenants, { name });

  if (isLoading && !data) return <IsLoading />;

  const tenants = data?.tenants ?? [];
  const visible = Math.min(tenants.length, VISIBLE_CARDS);

  return (
    <div className="flex min-h-svh flex-col gap-6 p-4 sm:p-10">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1 font-bold [font-family:'Phudu',sans-serif]">
          <div className="flex size-6 items-center justify-center rounded-md">
            <Image width={600} height={600} src="/assets/sntz.png" alt="" loading="eager" />
          </div>
          <span className="items-center text-xl text-[#0504aa]">SENTINÉZ</span>
        </div>
        <div className="flex items-center gap-2">
          <Language />
          <UserMenu />
        </div>
      </div>
      <div className="flex flex-1 flex-col items-center justify-center gap-6 sm:gap-10">
        <div className="flex w-full max-w-sm items-center gap-2">
          <div className="relative flex-1">
            <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
            <Input
              type="search"
              value={search}
              placeholder={t('searchPlaceholder')}
              aria-label={t('searchPlaceholder')}
              className="pl-9"
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <RefreshButton loading={isLoading} onRefresh={refresh} />
        </div>
        <div className="flex w-full flex-col items-center justify-center gap-4 sm:flex-row sm:flex-wrap sm:gap-10">
          {tenants.length > 0 && (
            <Carousel
              opts={CAROUSEL_OPTS}
              className="w-full max-w-sm sm:mx-12 sm:w-(--carousel-width) sm:max-w-[calc(100%-6rem)]"
              style={
                {
                  '--carousel-width': `${visible * SLIDE_WIDTH - SLIDE_GUTTER}px`,
                } as React.CSSProperties
              }
            >
              <CarouselContent className="flex-col gap-4 sm:flex-row sm:gap-0">
                {tenants.map((tenant) => (
                  <CarouselItem key={tenant.id} className="basis-auto">
                    <Tenant tenant={tenant} />
                  </CarouselItem>
                ))}
              </CarouselContent>
              {tenants.length > VISIBLE_CARDS && (
                <>
                  <CarouselPrevious className="hidden sm:inline-flex" />
                  <CarouselNext className="hidden sm:inline-flex" />
                </>
              )}
            </Carousel>
          )}
          {tenants.length === 0 && name && !isLoading && (
            <p className="text-muted-foreground w-full text-center text-sm sm:w-[300px]">
              {tc('noResults')}
            </p>
          )}
          <TeamJoin onCreated={() => refresh()} />
        </div>
      </div>
    </div>
  );
}
