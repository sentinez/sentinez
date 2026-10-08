'use client';

import type { Table } from '@tanstack/react-table';
import { Button } from '@sentinez/ui/components/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { PAGE_SIZE_OPTIONS } from '@/hooks/use-paged-list';

interface Props<T> {
  table: Table<T>;
  /** Total rows on the server */
  total: number;
}

/** Footer for a table with manualPagination (see usePagedList) */
export function TablePagination<T>({ table, total }: Props<T>) {
  const t = useTranslations('Pagination');
  const { pageIndex, pageSize } = table.getState().pagination;
  const from = total === 0 ? 0 : pageIndex * pageSize + 1;
  const to = Math.min(total, (pageIndex + 1) * pageSize);

  return (
    <div className="flex flex-wrap items-center justify-between gap-4 py-4">
      <div className="text-sm text-muted-foreground tabular-nums">
        {t('range', { from, to, total })}
      </div>
      <div className="flex flex-wrap items-center gap-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">{t('rowsPerPage')}</span>
          <Select
            value={String(pageSize)}
            onValueChange={(v) => {
              table.setPageSize(Number(v));
              table.setPageIndex(0);
            }}
          >
            <SelectTrigger size="sm" className="w-20">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {PAGE_SIZE_OPTIONS.map((size) => (
                <SelectItem key={size} value={String(size)}>
                  {size}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="text-sm tabular-nums">
          {t('page', { page: pageIndex + 1, count: table.getPageCount() })}
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            aria-label={t('first')}
            onClick={() => table.firstPage()}
            disabled={!table.getCanPreviousPage()}
          >
            <ChevronsLeft />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            aria-label={t('previous')}
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            <ChevronLeft />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            aria-label={t('next')}
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            <ChevronRight />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            aria-label={t('last')}
            onClick={() => table.lastPage()}
            disabled={!table.getCanNextPage()}
          >
            <ChevronsRight />
          </Button>
        </div>
      </div>
    </div>
  );
}
