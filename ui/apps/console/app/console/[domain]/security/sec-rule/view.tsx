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
import { useTranslations } from 'next-intl';
import { listSecRules, type Pages, deleteSecRule } from '@/lib/api/security';
import BadgeStatus from '@/components/badge-status';
import { Badge } from '@sentinez/ui/components/badge';
import { ActionValue, SecRule } from '@sentinez/proto/sentinez/apps/security/v1/model';
import {
  ActionType,
  Expression,
  actionTypeFromJSON,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { ACTION_TYPE_BADGE_VARIANT, statusLabel } from '@/lib/type/security';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { TablePagination } from '@/components/table-pagination';
import { usePagedList } from '@/hooks/use-paged-list';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useCallback, useMemo, useState } from 'react';

type SecRuleTranslator = ReturnType<typeof useTranslations<'SecRule'>>;
type CommonTranslator = ReturnType<typeof useTranslations<'Common'>>;

export const getColumns = (
  t: SecRuleTranslator,
  tc: CommonTranslator,
  actionTypeLabel: (type: ActionType) => string,
  onDelete: (rule: SecRule) => void,
): ColumnDef<SecRule>[] => [
  {
    accessorKey: 'name',
    size: 220,
    meta: { label: t('name') },
    header: ({ column }) => <div className="w-full">{t('name')}</div>,
    cell: ({ row }) => (
      <div className="w-full truncate">
        <Link
          className="text-blue-700 font-semibold underline"
          href={`./sec-rule/${row.original.id}`}
        >
          {row.original.name}
        </Link>
      </div>
    ),
  },
  {
    accessorKey: 'description',
    size: 0, // auto: takes the remaining width
    meta: { label: t('description') },
    header: () => <div>{t('description')}</div>,
    cell: ({ row }) => (
      <div className="truncate" title={row.original.description}>
        {row.original.description}
      </div>
    ),
  },
  {
    id: 'action',
    size: 140,
    meta: { label: t('action') },
    header: () => <div>{t('action')}</div>,
    cell: ({ row }) => {
      const { action, actionValue } = row.original;
      const type = action ? actionTypeFromJSON(action) : ActionType.ACTION_TYPE_UNSPECIFIED;
      const label = actionTypeLabel(type);
      const params = actionValue ? JSON.stringify(ActionValue.toJSON(actionValue)) : undefined;
      return (
        <div className="truncate" title={params}>
          <Badge variant={ACTION_TYPE_BADGE_VARIANT[type]}>{label}</Badge>
        </div>
      );
    },
  },
  {
    accessorKey: 'status',
    size: 120,
    meta: { label: t('status') },
    header: () => <div>{t('status')}</div>,
    cell: ({ row }) => {
      const s = statusLabel(row.original.status);
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
    meta: { label: t('priority') },
    header: () => <div className="w-full text-right">{t('priority')}</div>,
    cell: ({ row }) => {
      return <div className="w-full text-right">{row.original.priority}</div>;
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
            <DropdownMenuItem onClick={() => navigator.clipboard.writeText(row.original.id || '')}>
              {t('copyRuleId')}
            </DropdownMenuItem>
            <DropdownMenuItem variant="destructive" onClick={() => onDelete(row.original)}>
              {tc('delete')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    ),
  },
];

export default function View() {
  const t = useTranslations('SecRule');
  const tc = useTranslations('Common');
  const { actionTypeLabel } = useSecurityOptions();
  const [deleting, setDeleting] = useState<SecRule | null>(null);
  const [deletingBusy, setDeletingBusy] = useState(false);

  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});

  const fetchPage = useCallback(async (page: Pages, signal: AbortSignal) => {
    const res = await listSecRules({ page }, { signal });
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
  } = usePagedList(fetchPage, t('loadFailed'));

  const handleDelete = async () => {
    const id = deleting?.id;
    if (!id) return;
    setDeletingBusy(true);
    try {
      await deleteSecRule({ id });
      toast.success(t('deleted'));
      setDeleting(null);
      refresh();
    } catch (err: any) {
      toast.error(t('deleteFailed'));
    } finally {
      setDeletingBusy(false);
    }
  };

  const columns = useMemo(
    () => getColumns(t, tc, actionTypeLabel, setDeleting),
    [t, tc, actionTypeLabel],
  );

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
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')}>
        <Button size="sm" asChild>
          <Link href="./sec-rule/new">
            <PlusIcon className="w-4 h-4" />
            {tc('create')}
          </Link>
        </Button>
      </PageLayoutHeader>

      <PageLayoutContent>
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
                      {tc('loading')}
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
                      {tc('noResults')}
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
            <DialogTitle>{t('deleteTitle')}</DialogTitle>
            <DialogDescription>
              {tc('deleteConfirm', { name: deleting?.name ?? '' })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setDeleting(null)}>
              {tc('cancel')}
            </Button>
            <Button variant="destructive" disabled={deletingBusy} onClick={handleDelete}>
              {deletingBusy ? tc('deleting') : tc('delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  );
}
