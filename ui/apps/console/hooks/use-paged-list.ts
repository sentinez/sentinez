'use client';

import { useCallback, useEffect, useState } from 'react';
import type { PaginationState } from '@tanstack/react-table';
import axios from 'axios';
import { toast } from '@/lib/toast';
import type { Pages } from '@/lib/api/pages';

export interface PagedResult<T> {
  items: T[];
  total: number;
}

export const PAGE_SIZE_OPTIONS = [10, 20, 50];

/**
 * Server-side pagination for list tables: loads one page at a time through
 * `fetchPage` (memoize it with useCallback) and refetches when the page,
 * page size or `fetchPage` changes. The table keeps a 0-based pageIndex; the
 * API gets a 1-based `page.index`.
 */
export function usePagedList<T>(
  fetchPage: (page: Pages, signal: AbortSignal) => Promise<PagedResult<T>>,
  errorMessage: string,
  initialPageSize = 10,
) {
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: initialPageSize,
  });
  const [items, setItems] = useState<T[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    const { pageIndex, pageSize } = pagination;
    setLoading(true);
    fetchPage({ index: pageIndex + 1, size: pageSize, total: true }, controller.signal)
      .then((res) => {
        // The last page emptied (e.g. after a delete): step back to the new last page
        const last = Math.max(0, Math.ceil(res.total / pageSize) - 1);
        if (res.items.length === 0 && pageIndex > last) {
          setPagination((p) => ({ ...p, pageIndex: last }));
          return;
        }
        setItems(res.items);
        setTotal(res.total);
      })
      .catch((err) => {
        if (axios.isCancel(err) || controller.signal.aborted) return;
        toast.error(errorMessage);
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [fetchPage, errorMessage, pagination, reloadKey]);

  const refresh = useCallback(() => setReloadKey((k) => k + 1), []);
  const pageCount = Math.max(1, Math.ceil(total / pagination.pageSize));

  return { items, total, loading, pagination, setPagination, pageCount, refresh };
}
