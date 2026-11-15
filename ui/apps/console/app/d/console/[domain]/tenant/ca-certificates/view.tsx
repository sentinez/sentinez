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
import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { toast } from '@/lib/toast';
import { ChevronDown, KeyRound, MoreHorizontal, PlusIcon } from 'lucide-react';
import { useFormatter, useTranslations } from 'next-intl';
import { CaCertificate } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { deleteCaCertificate, listCaCertificates, type ApiOptions } from '@/lib/api/tenant';
import type { Pages } from '@/lib/api/pages';
import { TablePagination } from '@/components/table-pagination';
import { TableSkeletonRows } from '@/components/table-skeleton';
import { RefreshButton } from '@/components/refresh-button';
import { usePagedList } from '@/hooks/use-paged-list';
import { useResourceId } from '@/hooks/use-resource-id';
import {
  PageLayout,
  PageLayoutContent,
  PageLayoutHeader,
  PageLayoutToolbar,
} from '@/components/page-layout';
import { FixedTable } from '@/components/fixed-table';
import { useCallback, useMemo, useState } from 'react';
import { CaExpiryBadge, caUsageLabel } from './components';

type CaTranslator = ReturnType<typeof useTranslations<'CaCertificate'>>;
type CommonTranslator = ReturnType<typeof useTranslations<'Common'>>;
type Formatter = ReturnType<typeof useFormatter>;

export const getColumns = (
  t: CaTranslator,
  tc: CommonTranslator,
  format: Formatter,
  onDelete: (cert: CaCertificate) => void,
): ColumnDef<CaCertificate>[] => [
  {
    id: 'name',
    accessorFn: (c) => c.name,
    size: 0, // auto: takes the remaining width
    meta: { label: t('name') },
    header: () => <div className="w-full">{t('name')}</div>,
    cell: ({ row }) => (
      <div className="w-full truncate">
        <Link
          className="text-blue-700 font-semibold underline"
          href={`./ca-certificates/${row.original.id}`}
        >
          {row.original.name}
        </Link>
      </div>
    ),
  },
  {
    accessorKey: 'provider',
    size: 200,
    meta: { label: t('provider') },
    header: () => <div>{t('provider')}</div>,
    cell: ({ row }) => (
      <div className="truncate" title={row.original.subject}>
        {row.original.provider}
      </div>
    ),
  },
  {
    accessorKey: 'usage',
    size: 130,
    meta: { label: t('usage') },
    header: () => <div>{t('usage')}</div>,
    cell: ({ row }) => {
      const label = caUsageLabel(t, row.original.usage);
      // a long label is cut inside its cell, it never spills over the next one
      return (
        <Badge variant="outline" className="max-w-full" title={label}>
          {label}
        </Badge>
      );
    },
  },
  {
    id: 'expires',
    accessorFn: (c) => c.notAfter?.getTime() ?? 0,
    size: 220,
    meta: { label: t('expires') },
    header: () => <div>{t('expires')}</div>,
    cell: ({ row }) => (
      <div className="flex items-center gap-2">
        <span className="truncate tabular-nums">
          {row.original.notAfter
            ? format.dateTime(row.original.notAfter, { dateStyle: 'medium' })
            : '—'}
        </span>
        <CaExpiryBadge notAfter={row.original.notAfter} />
      </div>
    ),
  },
  {
    accessorKey: 'hasPrivateKey',
    size: 120,
    meta: { label: t('privateKey') },
    header: () => <div>{t('privateKey')}</div>,
    cell: ({ row }) =>
      row.original.hasPrivateKey ? (
        <div className="flex items-center gap-1">
          <KeyRound className="w-4 h-4" />
          {t('stored')}
        </div>
      ) : (
        <span className="text-muted-foreground">{t('none')}</span>
      ),
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
              onClick={async () => {
                await navigator.clipboard.writeText(row.original.fingerprintSha256);
                toast.success(t('copied'));
              }}
            >
              {t('copyFingerprint')}
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
  const t = useTranslations('CaCertificate');
  const tc = useTranslations('Common');
  const format = useFormatter();
  const [deleting, setDeleting] = useState<CaCertificate | null>(null);
  const [deletingBusy, setDeletingBusy] = useState(false);

  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});

  // Only the certificates of the resource of this domain
  const { resourceId, isLoading: resourceLoading } = useResourceId();
  const fetchPage = useCallback(
    async (page: Pages, signal: AbortSignal) => {
      if (!resourceId) return { items: [], total: 0 };
      const options: ApiOptions = { signal };
      const res = await listCaCertificates({ page, resourceIds: [resourceId] }, options);
      return { items: res.caCertificates, total: res.total };
    },
    [resourceId],
  );
  const {
    items: certs,
    total,
    loading,
    pagination,
    setPagination,
    pageCount,
    refresh,
  } = usePagedList(fetchPage, t('loadFailed'), 10, !resourceLoading);

  const handleDelete = async () => {
    const id = deleting?.id;
    if (!id) return;
    setDeletingBusy(true);
    try {
      await deleteCaCertificate({ id });
      toast.success(t('deleted'));
      setDeleting(null);
      refresh();
    } catch {
      toast.error(t('deleteFailed'));
    } finally {
      setDeletingBusy(false);
    }
  };

  const columns = useMemo(() => getColumns(t, tc, format, setDeleting), [t, tc, format]);

  const table = useReactTable<CaCertificate>({
    data: certs,
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
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      pagination,
    },
  });

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')}>
        <Button size="sm" asChild>
          <Link href="./ca-certificates/new">
            <PlusIcon className="w-4 h-4" />
            {tc('create')}
          </Link>
        </Button>
      </PageLayoutHeader>

      <PageLayoutContent>
        <div className="w-full">
          <PageLayoutToolbar>
            <Input
              placeholder={tc('filterNames')}
              value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
              onChange={(event) => table.getColumn('name')?.setFilterValue(event.target.value)}
              className="max-w-sm"
            />
            <RefreshButton loading={loading} onRefresh={refresh} />
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
          </PageLayoutToolbar>
          <div className="overflow-hidden rounded-md">
            <FixedTable table={table}>
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
              <TableBody aria-busy={loading}>
                {loading ? (
                  <TableSkeletonRows table={table} />
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
            </FixedTable>
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
