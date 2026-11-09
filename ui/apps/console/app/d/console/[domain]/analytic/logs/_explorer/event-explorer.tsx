'use client';

import { useMemo, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { useFormatter, useLocale, useTranslations } from 'next-intl';
import type { PaginationState } from '@tanstack/react-table';
import { RotateCcw, Search } from 'lucide-react';

import { Button } from '@sentinez/ui/components/button';
import { Card, CardContent, CardHeader } from '@sentinez/ui/components/card';
import { Input } from '@sentinez/ui/components/input';
import { Skeleton } from '@sentinez/ui/components/skeleton';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { RefreshButton } from '@/components/refresh-button';
import { useApi } from '@/hooks/use-api';
import { useResourceId } from '@/hooks/use-resource-id';
import {
  RESOLUTION_BUCKETS,
  getEventHistogramPages,
  getEventTopValues,
  type ApiOptions,
} from '@/lib/api/analytic';
import { dimmed } from './dimmed';
import { EventDetail } from './event-detail';
import { EventFields } from './event-fields';
import { EventHistogram, bucketsOf } from './event-histogram';
import { DEFAULT_PAGE_SIZE, EventTable } from './event-table';
import { matchesTerms, requestIdsOf, termsOf } from './query';
import { FieldFilters } from './field-filters';
import { RemovableBadge } from './removable-badge';
import { AT_PARAM, REQUEST_PARAM } from './request-link';
import {
  DATE_TIME_ZONE,
  clippedOf,
  resolutionOf,
  timeZoneOf,
  windowOf,
  zoomOf,
  type TimeRange,
} from './time-range';
import { TimeRangePicker } from './time-range-picker';
import { filterParam, type EventConfig, type FieldFilter } from './types';

const FIRST_PAGE: PaginationState = { pageIndex: 0, pageSize: DEFAULT_PAGE_SIZE };

const dateOf = (millis?: number) => (millis === undefined ? undefined : new Date(millis));

// Keyed a request and the refresh generation: useApi fetches it again when
// either changes
type Keyed<R> = { req: R; key: number };

// The fetch of keyed requests of an API function
const keyed =
  <R, T>(fn: (req: R, options?: ApiOptions) => Promise<T>) =>
  (k: Keyed<R>, options?: ApiOptions) =>
    fn(k.req, options);

const fetchHistogram = keyed(getEventHistogramPages);
const fetchTopValues = keyed(getEventTopValues);
// The fetch of a config without byId, never called
const noFetch = async () => null;

// The range around the time of a request a link opens the page on
const AROUND_MILLIS = 30 * 1000;

/**
 * Kibana Discover-like explorer of the events of a kind, for the resource
 * of the `[domain]` segment: a time range, the histogram of the events per
 * series, the top values of their fields (counted by the API over the
 * range), a filter of the page and the page of events, each opened in a
 * detail sheet. A click on a bar narrows the list and the top values to
 * its bucket.
 */
export function EventExplorer<T>({ config }: { config: EventConfig<T> }) {
  const t = useTranslations('Explorer');
  const format = useFormatter();
  const locale = useLocale();
  const { resourceId } = useResourceId();

  // A link to a request (RequestLink) opens the page around its time, the
  // filter on its id and its detail
  const params = useSearchParams();
  const focusId = params.get(REQUEST_PARAM) ?? '';
  const focusAt = Number(params.get(AT_PARAM)) || 0;

  const [range, setRange] = useState<TimeRange>(() =>
    focusAt ? { from: focusAt - AROUND_MILLIS, to: focusAt + AROUND_MILLIS } : { quick: 'last60m' },
  );
  // When the range was picked or refreshed: a quick range ends then
  const [now, setNow] = useState(() => Date.now());
  // The start of the bucket a click on the histogram selected
  const [bucket, setBucket] = useState<number | undefined>();
  const [pagination, setPagination] = useState<PaginationState>(FIRST_PAGE);
  const [query, setQuery] = useState(focusId);
  // The field filters a click on a top value adds: the API applies them to
  // the list and the top values
  const [filters, setFilters] = useState<FieldFilter[]>([]);
  const filterParams = useMemo(() => filters.map(filterParam), [filters]);
  // The event a click opened, null when closed; undefined until then, the
  // event of the request of a link opening
  const [selected, setSelected] = useState<T | null | undefined>();

  // The refresh generation, in every request: a refresh fetches them all
  // again, an absolute range not moving
  const [reloadKey, setReloadKey] = useState(0);

  const { kind, list, seriesOf } = config;
  const resolution = resolutionOf(range, now);
  const bucketMillis = RESOLUTION_BUCKETS[resolution].millis;
  const bounds = useMemo(() => windowOf(range, now), [range, now]);
  // The request ids of the filter are asked to the API, the other terms
  // filter the page; the histogram counts those of the API filters
  const requestIds = useMemo(() => requestIdsOf(query), [query]);
  // An absolute range ends its buckets with it: the API ends them before
  // to, in pages back to its start
  const histogram = useApi(
    fetchHistogram,
    resourceId
      ? {
          req: {
            kind,
            resolution,
            resourceIds: [resourceId],
            to: dateOf(bounds.to),
            from: dateOf(bounds.from),
            requestIds,
            filters: filterParams,
          },
          key: reloadKey,
        }
      : null,
  );
  // The API returns a fixed number of buckets: keep those within the range
  const buckets = useMemo(
    () =>
      bucketsOf(histogram.data ?? [], seriesOf).filter(
        (b) => b.time + bucketMillis > bounds.from && b.time < (bounds.to ?? Infinity),
      ),
    [histogram.data, seriesOf, bucketMillis, bounds],
  );
  const hits = bucket
    ? (buckets.find((b) => b.time === bucket)?.total ?? 0)
    : buckets.reduce((sum, b) => sum + b.total, 0);

  // The range of the list and of the top values: the selected bucket, else
  // the time range
  const listRange = useMemo(
    () => ({
      from: bucket ?? bounds.from,
      to: bucket === undefined ? bounds.to : bucket + bucketMillis,
    }),
    [bucket, bounds, bucketMillis],
  );
  const fieldNames = useMemo(() => config.fields.map((f) => f.name), [config.fields]);
  const topValues = useApi(
    fetchTopValues,
    resourceId
      ? {
          req: {
            kind,
            resourceIds: [resourceId],
            from: new Date(listRange.from),
            to: dateOf(listRange.to),
            fields: fieldNames,
            filters: filterParams,
          },
          key: reloadKey,
        }
      : null,
  );

  const fetchList = useMemo(() => keyed(list), [list]);
  const page = useApi(
    fetchList,
    resourceId
      ? {
          req: {
            resourceIds: [resourceId],
            requestIds,
            filters: filterParams,
            page: { index: pagination.pageIndex + 1, size: pagination.pageSize, total: false },
            from: new Date(listRange.from),
            to: dateOf(listRange.to),
          },
          key: reloadKey,
        }
      : null,
  );
  const events = useMemo(() => page.data ?? [], [page.data]);

  const { byId } = config;
  const fetchFocus = useMemo(() => (byId ? keyed(byId) : noFetch), [byId]);
  // Looked up around its time: the table has no index on the request id
  const focus = useApi(
    fetchFocus,
    byId && focusId && resourceId
      ? {
          req: {
            id: focusId,
            resourceId,
            from: focusAt ? new Date(focusAt - AROUND_MILLIS) : undefined,
            to: focusAt ? new Date(focusAt + AROUND_MILLIS) : undefined,
          },
          key: 0,
        }
      : null,
  );
  const loading = page.isLoading;

  const shown = useMemo(() => {
    const terms = termsOf(query);
    return events.filter((e) => matchesTerms(e, terms, config));
  }, [events, query, config]);

  const firstPage = () => setPagination((p) => (p.pageIndex === 0 ? p : { ...p, pageIndex: 0 }));
  const pickRange = (r: TimeRange) => {
    setRange(r);
    setNow(Date.now());
    setBucket(undefined);
    firstPage();
  };
  const refresh = () => {
    setNow(Date.now());
    setReloadKey((k) => k + 1);
    firstPage();
  };
  // A new request id of the filter changes the list: back to its first page
  const changeQuery = (next: string) => {
    if (requestIdsOf(next).join() !== requestIds.join()) firstPage();
    setQuery(next);
  };
  const filtered = query.trim() !== '' || bucket !== undefined || filters.length > 0;
  const resetFilters = () => {
    setQuery('');
    setFilters([]);
    setBucket(undefined);
    firstPage();
  };
  const pickBucket = (time?: number) => {
    setBucket(time === bucket ? undefined : time);
    firstPage();
  };
  const addFilter = (field: string, value: string) => {
    setFilters((fs) =>
      fs.some((f) => f.field === field && f.value === value) ? fs : [...fs, { field, value }],
    );
    firstPage();
  };
  const removeFilter = (index: number) => {
    setFilters((fs) => fs.filter((_, i) => i !== index));
    firstPage();
  };
  const clipped = clippedOf(range, now);
  const bucketLabel = (time: number) =>
    format.dateTimeRange(new Date(time), new Date(time + bucketMillis), DATE_TIME_ZONE);

  return (
    <PageLayout>
      <PageLayoutHeader title={config.title} subtitle={config.subtitle}>
        <div className="flex min-w-0 max-w-full items-center gap-2">
          <TimeRangePicker value={range} onChange={pickRange} />
          <RefreshButton
            loading={loading || histogram.isLoading || topValues.isLoading}
            onRefresh={refresh}
          />
        </div>
      </PageLayoutHeader>

      <PageLayoutContent>
        <Card className="gap-2 py-4">
          <CardHeader className="flex flex-wrap items-center gap-x-3 gap-y-2 px-3 sm:px-4">
            <div className="flex shrink-0 items-baseline gap-2">
              <span className="text-2xl font-bold tabular-nums">{format.number(hits)}</span>
              <span className="text-sm text-muted-foreground">{t('hits')}</span>
            </div>
            {bucket !== undefined && (
              <RemovableBadge
                className="min-w-0 shrink"
                removeLabel={t('clearBucket')}
                onRemove={() => pickBucket(undefined)}
              >
                {bucketLabel(bucket)}
              </RemovableBadge>
            )}
            <span className="ml-auto text-xs text-muted-foreground">
              <span className="hidden md:inline">{t('histogramHint')} · </span>
              {t('timeZone', { zone: timeZoneOf(locale) })}
            </span>
            {clipped !== undefined && (
              <p className="w-full text-xs text-muted-foreground">
                {t('histogramClipped', {
                  count: clipped,
                  unit: RESOLUTION_BUCKETS[resolution].unit,
                })}
              </p>
            )}
          </CardHeader>
          <CardContent className="px-2 sm:px-4">
            {/* A refetch keeps the chart, dimmed: a skeleton until the first data */}
            {!histogram.data ? (
              <Skeleton className="h-56 w-full" />
            ) : (
              <div className={dimmed(histogram.isLoading)}>
                <EventHistogram
                  data={buckets}
                  series={config.series}
                  resolution={resolution}
                  bucketMillis={bucketMillis}
                  selected={bucket}
                  onSelect={pickBucket}
                  onSelectRange={(from, to) => pickRange(zoomOf(resolution, from, to))}
                />
              </div>
            )}
          </CardContent>
        </Card>

        <div className="flex w-full max-w-2xl items-center gap-2">
          <div className="relative min-w-0 flex-1">
            <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => changeQuery(e.target.value)}
              placeholder={t('filterPlaceholder')}
              aria-label={t('filterPlaceholder')}
              className="pl-8 font-mono text-sm"
            />
          </div>
          <Button
            variant="outline"
            className="shrink-0"
            disabled={!filtered}
            title={t('resetFiltersHint')}
            onClick={resetFilters}
          >
            <RotateCcw />
            <span className="hidden sm:inline">{t('resetFilters')}</span>
            <span className="sr-only sm:hidden">{t('resetFilters')}</span>
          </Button>
        </div>
        <FieldFilters fields={config.fields} filters={filters} onRemove={removeFilter} />

        {/* Sized by the content width (sidebar, zoom), not the viewport */}
        <div className="@container">
          <div className="grid grid-cols-1 gap-6 @4xl:grid-cols-[14rem_minmax(0,1fr)]">
            <div className="hidden min-w-0 @4xl:block">
              <EventFields
                fields={config.fields}
                data={topValues.data}
                loading={topValues.isLoading}
                onFilter={addFilter}
              />
            </div>
            <EventTable
              config={config}
              events={shown}
              loading={loading}
              initial={page.data === null}
              pagination={pagination}
              onPaginationChange={setPagination}
              hasNext={events.length >= pagination.pageSize}
              onSelect={setSelected}
            />
          </div>
        </div>
      </PageLayoutContent>

      <EventDetail
        config={config}
        event={selected === undefined ? (focus.data ?? null) : selected}
        onClose={() => setSelected(null)}
      />
    </PageLayout>
  );
}
