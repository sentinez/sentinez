'use client';

import type { Table } from '@tanstack/react-table';
import { Skeleton } from '@sentinez/ui/components/skeleton';
import { TableCell, TableRow } from '@sentinez/ui/components/table';

const DEFAULT_ROWS = 5;

interface Props<T> {
  table: Table<T>;
}

/**
 * Placeholder rows for a table body while its page is loading. Renders as
 * many rows as the table currently shows (capped by the page size), so the
 * table keeps its height when the rows come back.
 */
export function TableSkeletonRows<T>({ table }: Props<T>) {
  const shown = table.getRowModel().rows.length || DEFAULT_ROWS;
  const rows = Math.min(shown, table.getState().pagination.pageSize);

  return Array.from({ length: rows }, (_, i) => (
    <TableRow key={i} className="max-h-fit" aria-hidden>
      {table.getVisibleLeafColumns().map((column) => (
        <TableCell key={column.id}>
          <Skeleton className="h-5 w-full" />
        </TableCell>
      ))}
    </TableRow>
  ));
}
