'use client';

import Link from 'next/link';
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { toast } from '@/lib/toast';
import { ChevronDown, MoreHorizontal, PlusIcon } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { listCoreRulesets, type Pages, deleteCoreRuleset } from '@/lib/api/security';
import BadgeStatus from '@/components/badge-status';
import { CoreRuleset } from '@sentinez/proto/sentinez/apps/security/v1/model';
import { EngineMode } from '@sentinez/proto/sentinez/types/coreruleset/v1/setting';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import { ENGINE_MODE_BADGE_VARIANT, statusLabel } from '@/lib/type/security';
import { TablePagination } from '@/components/table-pagination';
import { useDebounced } from '@/hooks/use-debounced';
import { usePagedList } from '@/hooks/use-paged-list';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useCallback, useEffect, useMemo, useState } from 'react';

type RulesetsTranslator = ReturnType<typeof useTranslations<'Rulesets'>>;
type CommonTranslator = ReturnType<typeof useTranslations<'Common'>>;
type SecurityOptions = ReturnType<typeof useSecurityOptions>;

export const getColumns = (
  t: RulesetsTranslator,
  tc: CommonTranslator,
  options: SecurityOptions,
  onDelete: (ruleset: CoreRuleset) => void,
): ColumnDef<CoreRuleset>[] => [
  {
    id: 'name',
    accessorFn: (r) => r.name,
    size: 220,
    meta: { label: t('name') },
    header: () => <div className="w-full">{t('name')}</div>,
    cell: ({ row }) => (
      <div className="w-full truncate">
        <Link
          className="text-blue-700 font-semibold underline"
          href={`./rulesets/${row.original.id}`}
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
    id: 'mode',
    accessorFn: (r) => r.content?.mode,
    size: 150,
    meta: { label: t('mode') },
    header: () => <div>{t('mode')}</div>,
    cell: ({ row }) => {
      const mode = row.original.content?.mode ?? EngineMode.ENGINE_MODE_UNSPECIFIED;
      return (
        <Badge variant={ENGINE_MODE_BADGE_VARIANT[mode]}>{options.engineModeLabel(mode)}</Badge>
      );
    },
  },
  {
    id: 'version',
    accessorFn: (r) => r.content?.version,
    size: 150,
    meta: { label: t('version') },
    header: () => <div>{t('version')}</div>,
    cell: ({ row }) => (
      <div className="truncate tabular-nums">
        {options.crsVersionLabel(row.original.content?.version)}
      </div>
    ),
  },
  {
    id: 'paranoiaLevel',
    accessorFn: (r) => r.content?.paranoiaLevel,
    size: 110,
    meta: { label: t('paranoiaLevel') },
    header: () => <div>{t('paranoiaLevel')}</div>,
    cell: ({ row }) => (
      <div className="truncate tabular-nums">
        {row.original.content?.paranoiaLevel || t('paranoiaDefault')}
      </div>
    ),
  },
  {
    id: 'ruleGroups',
    accessorFn: (r) => r.content?.categories.length,
    size: 120,
    meta: { label: t('ruleGroups') },
    header: () => <div>{t('ruleGroups')}</div>,
    cell: ({ row }) => {
      const count = row.original.content?.categories.length ?? 0;
      return (
        <div className="truncate tabular-nums">
          {count ? t('groupsCount', { count }) : t('allGroups')}
        </div>
      );
    },
  },
  {
    id: 'status',
    accessorFn: (r) => r.content?.status,
    size: 120,
    meta: { label: t('status') },
    header: () => <div>{t('status')}</div>,
    cell: ({ row }) => {
      const s = statusLabel(row.original.content?.status);
      return (
        <div className="capitalize">
          <BadgeStatus status={s} value={s} />
        </div>
      );
    },
  },
  {
    id: 'actions',
    size: 64,
    enableHiding: false,
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
              {t('copyId')}
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
  const t = useTranslations('Rulesets');
  const tc = useTranslations('Common');
  const options = useSecurityOptions();
  const [deleting, setDeleting] = useState<CoreRuleset | null>(null);
  const [deletingBusy, setDeletingBusy] = useState(false);

  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});

  // Filters are applied by the API, not on the loaded page
  const [name, setName] = useState('');
  const nameQuery = useDebounced(name);
  const [mode, setMode] = useState(EngineMode.ENGINE_MODE_UNSPECIFIED);
  const [status, setStatus] = useState(Status.STATUS_UNSPECIFIED);

  const fetchPage = useCallback(
    async (page: Pages, signal: AbortSignal) => {
      const res = await listCoreRulesets(
        {
          page,
          name: nameQuery.trim(),
          modes: mode ? [mode] : [],
          statuses: status ? [status] : [],
        },
        { signal },
      );
      return { items: res.coreRulesets, total: res.total };
    },
    [nameQuery, mode, status],
  );
  const {
    items: rulesets,
    total,
    loading,
    pagination,
    setPagination,
    pageCount,
    refresh,
  } = usePagedList(fetchPage, t('loadFailed'));

  // A new filter starts again from the first page
  useEffect(() => {
    setPagination((p) => (p.pageIndex === 0 ? p : { ...p, pageIndex: 0 }));
  }, [nameQuery, mode, status, setPagination]);

  const handleDelete = async () => {
    const id = deleting?.id;
    if (!id) return;
    setDeletingBusy(true);
    try {
      await deleteCoreRuleset({ id });
      toast.success(t('deleted'));
      setDeleting(null);
      refresh();
    } catch {
      toast.error(t('deleteFailed'));
    } finally {
      setDeletingBusy(false);
    }
  };

  const columns = useMemo(() => getColumns(t, tc, options, setDeleting), [t, tc, options]);

  const table = useReactTable<CoreRuleset>({
    data: rulesets,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount,
    onPaginationChange: setPagination,
    onColumnVisibilityChange: setColumnVisibility,
    state: {
      columnVisibility,
      pagination,
    },
  });

  return (
    <PageLayout>
      <PageLayoutHeader title={t('title')} subtitle={t('subtitle')}>
        <Button size="sm" asChild>
          <Link href="./rulesets/new">
            <PlusIcon className="w-4 h-4" />
            {tc('create')}
          </Link>
        </Button>
      </PageLayoutHeader>

      <PageLayoutContent>
        <div className="w-full">
          <div className="flex flex-wrap items-center gap-2 py-4">
            <Input
              aria-label={tc('filterNames')}
              placeholder={tc('filterNames')}
              value={name}
              onChange={(event) => setName(event.target.value)}
              className="max-w-sm"
            />
            <Select value={String(mode)} onValueChange={(v) => setMode(Number(v) as EngineMode)}>
              <SelectTrigger className="w-44" aria-label={t('mode')}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={String(EngineMode.ENGINE_MODE_UNSPECIFIED)}>
                  {t('allModes')}
                </SelectItem>
                {options.engineModeOptions.map((o) => (
                  <SelectItem key={o.value} value={String(o.value)}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={String(status)} onValueChange={(v) => setStatus(Number(v) as Status)}>
              <SelectTrigger className="w-44" aria-label={t('status')}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={String(Status.STATUS_UNSPECIFIED)}>
                  {t('allStatuses')}
                </SelectItem>
                {options.statusOptions.map((o) => (
                  <SelectItem key={o.value} value={String(o.value)}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
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
