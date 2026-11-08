'use client';

import { useMemo, useRef } from 'react';
import { useFormatter, useLocale, useTranslations } from 'next-intl';
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
  type PaginationState,
  type VisibilityState,
} from '@tanstack/react-table';
import {
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@sentinez/ui/components/table';
import { FixedTable, columnMinWidth } from '@/components/fixed-table';
import { useElementWidth } from '@/hooks/use-element-width';
import { TablePagination } from '@/components/table-pagination';
import { TableSkeletonRows } from '@/components/table-skeleton';
import { RequestLink } from './request-link';
import { dimmed } from './dimmed';
import { timeZoneOf } from './time-range';
import type { EventConfig, RequestRef } from './types';

const PAGE_SIZES = [20, 50, 100];
export const DEFAULT_PAGE_SIZE = 20;

// The columns that always show: the time, the link to the request, and
// the auto-width ones (the path)
const KEPT_COLUMNS = new Set(['time', 'request']);

/**
 * The visibility of columns so they fit width: the fixed-width ones hidden
 * from the right until their min widths sum to width at most; all of them
 * until the width is measured.
 */
function fitColumns<T>(columns: ColumnDef<T>[], width: number): VisibilityState {
  const visibility: VisibilityState = {};
  if (width <= 0) return visibility;
  let total = columns.reduce((sum, c) => sum + columnMinWidth(c.size), 0);
  for (let i = columns.length - 1; i >= 0 && total > width; i--) {
    const { id, size } = columns[i] ?? {};
    if (!id || !size || KEPT_COLUMNS.has(id)) continue;
    visibility[id] = false;
    total -= size;
  }
  return visibility;
}

// The column linking an event to its request on the ingress page
function requestColumn<T>(
  requestOf: (e: T) => RequestRef | undefined,
  label: string,
): ColumnDef<T> {
  return {
    id: 'request',
    size: 48,
    header: () => <span className="sr-only">{label}</span>,
    cell: ({ row }) => {
      const request = requestOf(row.original);
      return request?.id ? <RequestLink request={request} iconOnly /> : null;
    },
  };
}

interface Props<T> {
  config: EventConfig<T>;
  events: T[];
  /** Fetching the page: the last rows stay, dimmed */
  loading: boolean;
  /** No page fetched yet: skeleton rows */
  initial: boolean;
  pagination: PaginationState;
  onPaginationChange: (p: PaginationState) => void;
  /** The page is full, so a next one may exist (the API has no total) */
  hasNext: boolean;
  onSelect: (e: T) => void;
}

/** The events of a page, the latest first; a click opens one */
export function EventTable<T>({
  config,
  events,
  loading,
  initial,
  pagination,
  onPaginationChange,
  hasNext,
  onSelect,
}: Props<T>) {
  const t = useTranslations('Explorer');
  const tc = useTranslations('Common');
  const format = useFormatter();
  const zone = timeZoneOf(useLocale());
  const { timeOf, idOf, requestOf } = config;
  const columns = useMemo<ColumnDef<T>[]>(
    () => [
      {
        id: 'time',
        size: 170,
        // The zone once in the header, not in every row
        header: () => (
          <span className="whitespace-nowrap">
            {t('time')} <span className="font-normal text-muted-foreground">({zone})</span>
          </span>
        ),
        accessorFn: timeOf,
        cell: ({ row }) => (
          <span className="whitespace-nowrap font-mono text-xs tabular-nums">
            {format.dateTime(new Date(timeOf(row.original)), {
              dateStyle: 'short',
              timeStyle: 'medium',
            })}
          </span>
        ),
      },
      ...config.columns,
      ...(requestOf ? [requestColumn(requestOf, t('viewRequest'))] : []),
    ],
    [t, format, zone, timeOf, config.columns, requestOf],
  );
  // The table fits its container: the columns that do not are hidden (all
  // the fields are in the detail)
  const frame = useRef<HTMLDivElement>(null);
  const width = useElementWidth(frame);
  const columnVisibility = useMemo(() => fitColumns(columns, width), [columns, width]);
  const table = useReactTable<T>({
    data: events,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getRowId: idOf,
    manualPagination: true,
    // unknown: the API has no total
    pageCount: -1,
    state: { pagination, columnVisibility },
    onPaginationChange: (updater) =>
      onPaginationChange(typeof updater === 'function' ? updater(pagination) : updater),
  });

  return (
    <div className="flex min-w-0 flex-col">
      <div ref={frame} className="overflow-x-auto rounded-md border">
        <FixedTable table={table}>
          <TableHeader>
            {table.getHeaderGroups().map((group) => (
              <TableRow key={group.id}>
                {group.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    style={
                      header.column.columnDef.size
                        ? { width: header.column.columnDef.size }
                        : undefined
                    }
                  >
                    {flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody aria-busy={loading} className={dimmed(loading && !initial)}>
            {initial ? (
              <TableSkeletonRows table={table} />
            ) : table.getRowModel().rows.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  tabIndex={0}
                  className="cursor-pointer"
                  onClick={() => onSelect(row.original)}
                  onKeyDown={(e) => e.key === 'Enter' && onSelect(row.original)}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id} className="py-1.5">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  {tc('noResults')}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </FixedTable>
      </div>

      <TablePagination table={table} hasNext={!loading && hasNext} pageSizes={PAGE_SIZES} />
    </div>
  );
}
