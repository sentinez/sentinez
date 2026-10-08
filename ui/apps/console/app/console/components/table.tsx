'use client';

import * as React from 'react';
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
} from '@tanstack/react-table';
import { ArrowUpDown, ChevronDown, MoreHorizontal } from 'lucide-react';

import { Button } from '@sentinez/ui/components/button';
import { Checkbox } from '@sentinez/ui/components/checkbox';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';
import { Input } from '@sentinez/ui/components/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@sentinez/ui/components/table';
import { UniqueVisitorChart } from './chart';
import BadgeStatus from '@/components/badge-status';
import { usePagedList } from '@/hooks/use-paged-list';
import { TablePagination } from '@/components/table-pagination';
import type { Pages } from '@/lib/api/pages';
import { listResources, type TenantResource } from '@/lib/api/tenant';

// "PLAN_STANDARD" -> "standard", "STATUS_ACTIVE" -> "active"
const enumLabel = (v: string | undefined, prefix: string) =>
  (v ?? '').replace(prefix, '').toLowerCase();

function rowOf(r: TenantResource): Payment {
  return {
    id: r.id ?? '',
    plan: enumLabel(r.plan, 'PLAN_'),
    status: enumLabel(r.status, 'STATUS_') as Payment['status'],
    name: r.resourceDomain ?? '',
  };
}

export type Payment = {
  id: string;
  plan: string;
  status: 'active' | 'disable' | 'success' | 'failed';
  name: string;
};

export const columns: ColumnDef<Payment>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => <div className="w-full">Name</div>,
    cell: ({ row }) => (
      <div className="w-full lowercase truncate">
        <a
          className=" text-blue-700 font-semibold underline"
          href={`/console/${row.original.name}/tenant/resource`}
        >
          {row.getValue('name')}
        </a>
      </div>
    ),
  },
  {
    accessorKey: 'status',
    header: () => <div>Status</div>,
    cell: ({ row }) => (
      <div className="capitalize">
        <BadgeStatus status={row.getValue('status')} value={row.getValue('status')} />
      </div>
    ),
  },
  {
    accessorKey: 'unique visitor',
    header: () => <div className="w-full">Unique Visitor</div>,
    cell: () => (
      <div className="w-full max-h-16 overflow-hidden flex items-center">
        <UniqueVisitorChart />
      </div>
    ),
  },
  {
    accessorKey: 'plan',
    header: () => <div className="w-full text-right">Plan</div>,
    cell: ({ row }) => {
      return <div className="w-full text-right font-medium capitalize">{row.getValue('plan')}</div>;
    },
  },
  {
    id: 'actions',
    header: () => <div className="w-full" />,
    cell: ({ row }) => (
      <div className="w-full flex justify-center">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="h-8 w-8 p-0">
              <MoreHorizontal />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigator.clipboard.writeText(row.original.id)}>
              Copy resource ID
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    ),
  },
];

export function ResourceTable() {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});
  const fetchPage = React.useCallback(async (page: Pages, signal: AbortSignal) => {
    const list = await listResources({ page }, { signal });
    if (!list) throw new Error('list resources failed');
    return { items: list.resources ?? [], total: Number(list.total ?? 0) };
  }, []);
  const { items, total, pagination, setPagination, pageCount } = usePagedList(
    fetchPage,
    'Failed to load resources',
  );
  const data = React.useMemo(() => items.map(rowOf), [items]);

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount,
    onPaginationChange: setPagination,
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
      pagination,
    },
  });

  return (
    <div className="w-full">
      <div className="flex items-center py-4">
        <Input
          placeholder="Filter names..."
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onChange={(event) => table.getColumn('name')?.setFilterValue(event.target.value)}
          className="max-w-sm"
        />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto">
              Columns <ChevronDown />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {table
              .getAllColumns()
              .filter((column) => column.getCanHide())
              .map((column) => {
                return (
                  <DropdownMenuCheckboxItem
                    key={column.id}
                    className="capitalize"
                    checked={column.getIsVisible()}
                    onCheckedChange={(value) => column.toggleVisibility(!!value)}
                  >
                    {column.id}
                  </DropdownMenuCheckboxItem>
                );
              })}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <div className="overflow-hidden rounded-md">
        <Table className="table-fixed w-full">
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id} className="max-h-fit">
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} className="max-h-fit">
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow className="max-h-fit">
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <TablePagination table={table} total={total} />
    </div>
  );
}
