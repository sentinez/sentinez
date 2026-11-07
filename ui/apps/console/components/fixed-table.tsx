'use client';

import type { ReactNode } from 'react';
import type { Table as TableInstance } from '@tanstack/react-table';
import { Table } from '@sentinez/ui/components/table';

/** Narrowest an auto-width column (`size: 0`) gets before the table scrolls */
const AUTO_COLUMN_MIN_WIDTH = 200;

interface Props<T> {
  table: TableInstance<T>;
  children: ReactNode;
}

/**
 * `table-fixed` list table whose columns take their `columnDef.size`, the
 * auto-width ones (`size: 0`) the remaining width. Its min width is the sum
 * of the visible columns (AUTO_COLUMN_MIN_WIDTH for an auto one), so a
 * narrow screen scrolls the table instead of squeezing the auto columns to
 * nothing and spilling their text over the next column.
 */
export function FixedTable<T>({ table, children }: Readonly<Props<T>>) {
  const minWidth = table
    .getVisibleLeafColumns()
    .reduce((width, column) => width + (column.columnDef.size || AUTO_COLUMN_MIN_WIDTH), 0);

  return (
    <Table className="table-fixed w-full" style={{ minWidth }}>
      {children}
    </Table>
  );
}
