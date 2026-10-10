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
import { useTranslations } from 'next-intl';

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
import type { Resource } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { planToJSON, statusToJSON } from '@sentinez/proto/sentinez/types/v1/known';
import { listResources } from '@/lib/api/tenant';

// "PLAN_STANDARD" -> "standard", "STATUS_ACTIVE" -> "active"
const enumLabel = (v: string, prefix: string) => v.replace(prefix, '').toLowerCase();

function rowOf(r: Resource): Payment {
  return {
    id: r.id,
    plan: enumLabel(planToJSON(r.plan), 'PLAN_'),
    status: enumLabel(statusToJSON(r.status), 'STATUS_') as Payment['status'],
    name: r.resourceDomain,
  };
}

export type Payment = {
  id: string;
  plan: string;
  status: 'active' | 'disable' | 'success' | 'failed';
  name: string;
};

type DomainTranslator = ReturnType<typeof useTranslations<'Domain'>>;

export const getColumns = (t: DomainTranslator): ColumnDef<Payment>[] => [
  {
    accessorKey: 'name',
    meta: { label: t('name') },
    header: ({ column }) => <div className="w-full">{t('name')}</div>,
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
    meta: { label: t('status') },
    header: () => <div>{t('status')}</div>,
    cell: ({ row }) => (
      <div className="capitalize">
        <BadgeStatus status={row.getValue('status')} value={row.getValue('status')} />
      </div>
    ),
  },
  {
    accessorKey: 'unique visitor',
    meta: { label: t('uniqueVisitor') },
    header: () => <div className="w-full">{t('uniqueVisitor')}</div>,
    cell: () => (
      <div className="w-full max-h-16 overflow-hidden flex items-center">
        <UniqueVisitorChart />
      </div>
    ),
  },
  {
    accessorKey: 'plan',
    meta: { label: t('plan') },
    header: () => <div className="w-full text-right">{t('plan')}</div>,
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
              {t('copyResourceId')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    ),
  },
];

export function ResourceTable() {
  const t = useTranslations('Domain');
  const tc = useTranslations('Common');
  const columns = React.useMemo(() => getColumns(t), [t]);
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});
  const fetchPage = React.useCallback(async (page: Pages, signal: AbortSignal) => {
    const list = await listResources({ page }, { signal });
    if (!list) throw new Error('list resources failed');
    return { items: list.resources, total: list.total };
  }, []);
  const { items, total, pagination, setPagination, pageCount } = usePagedList(
    fetchPage,
    t('loadFailed'),
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
          placeholder={tc('filterNames')}
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onChange={(event) => table.getColumn('name')?.setFilterValue(event.target.value)}
          className="max-w-sm"
        />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto">
              {tc('columns')} <ChevronDown />
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
                    {column.columnDef.meta?.label ?? column.id}
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
                  {tc('noResults')}
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
