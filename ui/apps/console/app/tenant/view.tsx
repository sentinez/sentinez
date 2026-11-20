'use client';

import { useEffect, useState } from 'react';
import {
  type CarouselApi,
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from '@sentinez/ui/components/carousel';
import { Input } from '@sentinez/ui/components/input';
import { cn } from '@sentinez/ui/lib/utils';
import { Search } from 'lucide-react';
import Image from 'next/image';
import { useTheme } from 'next-themes';
import { useTranslations } from 'next-intl';

import LetterGlitch from '@/components/letter-glitch/letter-glitch';
import Language from '@/components/language';
import { RefreshButton } from '@/components/refresh-button';
import { UserMenu } from '@/components/user-menu';
import { useApi } from '@/hooks/use-api';
import { useDebounced } from '@/hooks/use-debounced';
import { listTenants } from '@/lib/api/tenant';
import { withMinLoading } from '@/lib/min-loading';

import { Tenant, TeamJoin, TenantSkeleton } from './components/card';

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

// Keeps the skeleton cards on screen for at least MIN_LOADING_MILLIS
const listTenantsMinLoading = (...args: Parameters<typeof listTenants>) =>
  withMinLoading(listTenants(...args));

// Fade-up entrance (tw-animate-css), staggered by the card's position
const ENTER = 'animate-in fade-in slide-in-from-bottom-4 fill-mode-both duration-400 ease-out';
const enterDelay = (index: number) => ({ animationDelay: `${index * 70}ms` });

/**
 * Whether the carousel has cards out of view, on either side; kept in step
 * as it is resized or its cards change.
 */
function useScrollable() {
  const [api, setApi] = useState<CarouselApi>();
  const [value, setValue] = useState(false);

  useEffect(() => {
    if (!api) return;
    const update = () => setValue(api.canScrollPrev() || api.canScrollNext());
    update();
    api.on('reInit', update).on('select', update);
    return () => {
      api.off('reInit', update).off('select', update);
    };
  }, [api]);

  return { value, setApi };
}

export default function View() {
  const t = useTranslations('Tenant');
  const tc = useTranslations('Common');
  const [search, setSearch] = useState('');
  const name = useDebounced(search.trim());
  const { data, isLoading, refresh } = useApi(listTenantsMinLoading, { name });
  const scrollable = useScrollable();

  const tenants = data?.tenants ?? [];
  // While loading, keep as many cards as before (a full row on first load)
  const visible = Math.min(
    isLoading ? tenants.length || VISIBLE_CARDS : tenants.length,
    VISIBLE_CARDS,
  );

  return (
    <div className="relative isolate flex min-h-svh flex-col gap-6 p-4 sm:p-10">
      <TenantBackdrop />
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1 font-bold font-['Phudu',sans-serif]">
          <div className="flex size-6 items-center justify-center rounded-md">
            <Image width={600} height={600} src="/assets/sz.png" alt="" loading="eager" />
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
          <RefreshButton
            loading={isLoading}
            onRefresh={refresh}
            className="border-input bg-transparent"
          />
        </div>
        <div className="flex w-full flex-col items-center justify-center gap-4 sm:flex-row sm:flex-wrap sm:gap-10">
          {visible > 0 && (
            <Carousel
              opts={CAROUSEL_OPTS}
              setApi={scrollable.setApi}
              className="w-full max-w-sm sm:mx-12 sm:w-(--carousel-width) sm:max-w-[calc(100%-6rem)]"
              style={
                {
                  '--carousel-width': `${visible * SLIDE_WIDTH - SLIDE_GUTTER}px`,
                } as React.CSSProperties
              }
            >
              <CarouselContent
                // room for the hover lift and shadow of the cards
                viewportClassName="-my-8 py-8"
                className="flex-col gap-4 sm:flex-row sm:gap-0"
                aria-busy={isLoading}
              >
                {isLoading
                  ? Array.from({ length: visible }, (_, i) => (
                      <CarouselItem key={i} className="basis-auto">
                        <TenantSkeleton />
                      </CarouselItem>
                    ))
                  : tenants.map((tenant, i) => (
                      <CarouselItem key={tenant.id} className="basis-auto">
                        <div className={ENTER} style={enterDelay(i)}>
                          <Tenant tenant={tenant} />
                        </div>
                      </CarouselItem>
                    ))}
              </CarouselContent>
              {/* also when a narrow screen cuts the row of cards */}
              {!isLoading && scrollable.value && (
                <>
                  <CarouselPrevious className="hidden sm:inline-flex" />
                  <CarouselNext className="hidden sm:inline-flex" />
                </>
              )}
            </Carousel>
          )}
          {tenants.length === 0 && name && !isLoading && (
            <p className="text-muted-foreground w-full text-center text-sm sm:w-75">
              {tc('noResults')}
            </p>
          )}
          <div
            className={cn(ENTER, 'w-full max-w-sm sm:w-auto')}
            style={enterDelay(isLoading ? 0 : visible)}
          >
            <TeamJoin onCreated={() => refresh()} />
          </div>
        </div>
      </div>
    </div>
  );
}

/**
 * The glitching letters of the auth page, toned down behind the cards: dark
 * ink on the light theme, light letters on the dark one. The cursor still
 * stirs them: they listen on the window, not on this layer. Nothing renders
 * until the theme is known, so the dark theme does not start light.
 */
function TenantBackdrop() {
  const { resolvedTheme } = useTheme();

  return (
    <div
      aria-hidden
      className="pointer-events-none absolute inset-0 -z-10 overflow-hidden opacity-60"
    >
      {resolvedTheme && <LetterGlitch lightMode={resolvedTheme !== 'dark'} />}
    </div>
  );
}
