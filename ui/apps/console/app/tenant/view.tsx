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
import { useTranslations } from 'next-intl';

import { useApi } from '@/hooks/use-api';
import { useDebounced } from '@/hooks/use-debounced';
import { listTenants } from '@/lib/api/tenant';

import { Tenant, TeamJoin } from './components/card';

// Tenant cards visible at once; more than this scrolls
const VISIBLE_CARDS = 3;
// Card width (300px) plus the 16px gutter of a CarouselItem
const SLIDE_WIDTH = 316;
const SLIDE_GUTTER = 16;

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
    <div className="flex min-h-svh flex-col items-center justify-center gap-10 p-10">
      <div className="relative w-full max-w-sm">
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
      <div className="flex w-full flex-wrap items-center justify-center gap-10">
        {tenants.length > 0 && (
          <Carousel
            opts={{ align: 'start' }}
            className="mx-12 max-w-[calc(100%-6rem)]"
            style={{ width: visible * SLIDE_WIDTH - SLIDE_GUTTER }}
          >
            <CarouselContent>
              {tenants.map((tenant) => (
                <CarouselItem key={tenant.id} className="basis-auto">
                  <Tenant tenant={tenant} />
                </CarouselItem>
              ))}
            </CarouselContent>
            {tenants.length > VISIBLE_CARDS && (
              <>
                <CarouselPrevious />
                <CarouselNext />
              </>
            )}
          </Carousel>
        )}
        {tenants.length === 0 && name && !isLoading && (
          <p className="text-muted-foreground w-[300px] text-center text-sm">{tc('noResults')}</p>
        )}
        <TeamJoin onCreated={() => refresh()} />
      </div>
    </div>
  );
}
