'use client';

import Link from 'next/link';
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@sentinez/ui/components/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@sentinez/ui/components/dialog';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { toast } from '@/lib/toast';
import { ChevronDown, MoreHorizontal, PlusIcon } from 'lucide-react';
import { listSecRulesWithTotal, type Pages, deleteSecRule } from '@/lib/api/security';
import BadgeStatus from '@/components/badge-status';
import { Badge } from '@sentinez/ui/components/badge';
import { SecRule } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { ActionType, Expression } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { ACTION_TYPE_BADGE_VARIANT, ACTION_TYPE_LABEL, statusLabel } from '@/lib/type/security';
import { TablePagination } from '@/components/table-pagination';
import { usePagedList } from '@/hooks/use-paged-list';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useCallback, useMemo, useState } from 'react';

export const getColumns = (onDelete: (rule: SecRule) => void): ColumnDef<SecRule>[] => [
  {
    accessorKey: 'name',
    size: 220,
    header: ({ column }) => <div className="w-full">Name</div>,
    cell: ({ row }) => (
      <div className="w-full truncate">
        <Link
          className="text-blue-700 font-semibold underline"
          href={`./sec-rule/${row.original.ingressRuntime?.id}`}
        >
          {row.original.ingressRuntime?.name}
        </Link>
      </div>
    ),
  },
  {
    accessorKey: 'description',
    size: 0, // auto: takes the remaining width
    header: () => <div>Description</div>,
    cell: ({ row }) => (
      <div className="truncate" title={row.original.ingressRuntime?.description}>
        {row.original.ingressRuntime?.description}
      </div>
    ),
  },
  {
    id: 'action',
    size: 140,
    header: () => <div>Action</div>,
    cell: ({ row }) => {
      const action = row.original.ingressRuntime?.action;
      const type = action?.type ?? ActionType.ACTION_TYPE_UNSPECIFIED;
      const label = ACTION_TYPE_LABEL[type];
      const params = action?.params ? JSON.stringify(action.params) : undefined;
      return (
        <div className="truncate" title={params}>
          <Badge variant={ACTION_TYPE_BADGE_VARIANT[type]}>{label ?? 'Unknown'}</Badge>
        </div>
      );
    },
  },
  {
    accessorKey: 'status',
    size: 120,
    header: () => <div>Status</div>,
    cell: ({ row }) => {
      const s = statusLabel(row.original.ingressRuntime?.status);
      return (
        <div className="capitalize">
          <BadgeStatus status={s as any} value={s} />
        </div>
      );
    },
  },
  {
    accessorKey: 'priority',
    size: 100,
    header: () => <div className="w-full text-right">Priority</div>,
    cell: ({ row }) => {
      return <div className="w-full text-right">{row.original.ingressRuntime?.priority}</div>;
    },
  },
  {
    id: 'actions',
    size: 64,
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
            <DropdownMenuItem
              onClick={() => navigator.clipboard.writeText(row.original.ingressRuntime?.id || '')}
            >
              Copy Rule ID
            </DropdownMenuItem>
            <DropdownMenuItem variant="destructive" onClick={() => onDelete(row.original)}>
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    ),
  },
];

export default function View() {
  const [deleting, setDeleting] = useState<SecRule | null>(null);
  const [deletingBusy, setDeletingBusy] = useState(false);

  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});

  const fetchPage = useCallback(async (page: Pages, signal: AbortSignal) => {
    const res = await listSecRulesWithTotal({ page }, { signal });
    return { items: res.secRules, total: res.total };
  }, []);
  const {
    items: rules,
    total,
    loading,
    pagination,
    setPagination,
    pageCount,
    refresh,
  } = usePagedList(fetchPage, 'Failed to load SecRules');

  const handleDelete = async () => {
    const id = deleting?.ingressRuntime?.id;
    if (!id) return;
    setDeletingBusy(true);
    try {
      await deleteSecRule(id);
      toast.success('SecRule deleted successfully');
      setDeleting(null);
      refresh();
    } catch (err: any) {
      toast.error('Failed to delete SecRule');
    } finally {
      setDeletingBusy(false);
    }
  };

  const columns = useMemo(() => getColumns(setDeleting), []);

  const table = useReactTable<SecRule>({
    data: rules,
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
    <PageLayout>
      <PageLayoutHeader title="Security Rule" subtitle="Manage active security rules.">
        <Button size="sm" asChild>
          <Link href="./sec-rule/new">
            <PlusIcon className="w-4 h-4" />
            Create
          </Link>
        </Button>
      </PageLayoutHeader>

      <PageLayoutContent>
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
                        <TableHead
                          key={header.id}
                          style={
                            header.column.columnDef.size
                              ? { width: header.column.columnDef.size }
                              : undefined
                          }
                        >
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
                {loading ? (
                  <TableRow className="max-h-fit">
                    <TableCell colSpan={columns.length} className="h-24 text-center">
                      Loading...
                    </TableCell>
                  </TableRow>
                ) : table.getRowModel().rows?.length ? (
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
      </PageLayoutContent>

      <Dialog open={!!deleting} onOpenChange={(o) => !o && setDeleting(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete SecRule</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete &quot;{deleting?.ingressRuntime?.name}&quot;? This
              action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setDeleting(null)}>
              Cancel
            </Button>
            <Button variant="destructive" disabled={deletingBusy} onClick={handleDelete}>
              {deletingBusy ? 'Deleting...' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  );
}
