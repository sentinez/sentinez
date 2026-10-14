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
import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { toast } from '@/lib/toast';
import { ChevronDown, MoreHorizontal, PlusIcon } from 'lucide-react';
import { useFormatter, useTranslations } from 'next-intl';
import type { Member } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import type { User } from '@sentinez/proto/sentinez/apps/iam/v1/model';
import { deleteMember, listMembers } from '@/lib/api/tenant';
import { ListUsers } from '@/lib/api/iam/user';
import type { Pages } from '@/lib/api/pages';
import { TablePagination } from '@/components/table-pagination';
import { TableSkeletonRows } from '@/components/table-skeleton';
import { RefreshButton } from '@/components/refresh-button';
import { usePagedList } from '@/hooks/use-paged-list';
import { useTenantId } from '@/hooks/use-tenant-id';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useCallback, useMemo, useState } from 'react';
import { MemberAvatar, roleLabel } from './components';

/** A member and its IAM user (missing when the user lookup failed) */
export interface MemberRow {
  member: Member;
  user?: User;
}

type MembersTranslator = ReturnType<typeof useTranslations<'Members'>>;
type CommonTranslator = ReturnType<typeof useTranslations<'Common'>>;
type EnumTranslator = ReturnType<typeof useTranslations<'Enum'>>;
type Formatter = ReturnType<typeof useFormatter>;

const displayName = (r: MemberRow) => r.user?.fullName || r.user?.email || r.member.userId;

export const getColumns = (
  t: MembersTranslator,
  tc: CommonTranslator,
  te: EnumTranslator,
  format: Formatter,
  onDelete: (row: MemberRow) => void,
): ColumnDef<MemberRow>[] => [
  {
    // Name and email both match the name filter
    id: 'name',
    accessorFn: (r) => `${displayName(r)} ${r.user?.email ?? ''}`,
    size: 0, // auto: takes the remaining width
    meta: { label: t('member') },
    header: () => <div className="w-full">{t('member')}</div>,
    cell: ({ row }) => {
      const r = row.original;
      const name = displayName(r);
      return (
        <div className="flex items-center gap-3 min-w-0">
          <MemberAvatar name={name} seed={r.user?.fullName || r.member.userId} />
          <div className="grid min-w-0 leading-tight">
            <Link
              className="truncate text-blue-700 font-semibold underline"
              href={`./member/${r.member.id}`}
            >
              {name}
            </Link>
            {r.user?.email && (
              <span className="truncate text-xs text-muted-foreground">{r.user.email}</span>
            )}
          </div>
        </div>
      );
    },
  },
  {
    id: 'roles',
    size: 220,
    meta: { label: t('roles') },
    header: () => <div>{t('roles')}</div>,
    cell: ({ row }) => (
      <div className="flex flex-wrap gap-1">
        {row.original.member.roles.map((role) => (
          <Badge key={role} variant="secondary">
            {roleLabel(te, role)}
          </Badge>
        ))}
      </div>
    ),
  },
  {
    id: 'joinedAt',
    size: 160,
    meta: { label: t('joinedAt') },
    header: () => <div>{t('joinedAt')}</div>,
    cell: ({ row }) => {
      const at = row.original.member.metadata?.createdAt;
      return (
        <div className="truncate tabular-nums">
          {at ? format.dateTime(at, { dateStyle: 'medium' }) : tc('notAvailable')}
        </div>
      );
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
            <DropdownMenuItem asChild>
              <Link href={`./member/${row.original.member.id}`}>{t('editRoles')}</Link>
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => navigator.clipboard.writeText(row.original.member.userId)}
            >
              {t('copyUserId')}
            </DropdownMenuItem>
            <DropdownMenuItem variant="destructive" onClick={() => onDelete(row.original)}>
              {t('remove')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    ),
  },
];

export default function View() {
  const t = useTranslations('Members');
  const tc = useTranslations('Common');
  const te = useTranslations('Enum');
  const format = useFormatter();
  const [deleting, setDeleting] = useState<MemberRow | null>(null);
  const [deletingBusy, setDeletingBusy] = useState(false);

  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});

  // Only the members of the selected tenant
  const [tenantId] = useTenantId();
  const fetchPage = useCallback(
    async (page: Pages, signal: AbortSignal) => {
      const res = await listMembers({ page, tenantIds: [tenantId] }, { signal });
      const ids = res.members.map((m) => m.userId);
      // Names and emails live in IAM; without them the rows still show user ids
      const users =
        ids.length === 0
          ? []
          : await ListUsers({ ids }, { signal })
              .then((u) => u.users)
              .catch(() => [] as User[]);
      const byId = new Map(users.map((u) => [u.id, u]));
      const items = res.members.map((member) => ({ member, user: byId.get(member.userId) }));
      return { items, total: res.total };
    },
    [tenantId],
  );
  const { items, total, loading, pagination, setPagination, pageCount, refresh } = usePagedList(
    fetchPage,
    t('loadFailed'),
    10,
    // Wait for the selected tenant: it is unknown while rendering on the server
    tenantId !== '',
  );

  const handleDelete = async () => {
    const id = deleting?.member.id;
    if (!id) return;
    setDeletingBusy(true);
    try {
      await deleteMember({ id });
      toast.success(t('removed'));
      setDeleting(null);
      refresh();
    } catch {
      toast.error(t('removeFailed'));
    } finally {
      setDeletingBusy(false);
    }
  };

  const columns = useMemo(() => getColumns(t, tc, te, format, setDeleting), [t, tc, te, format]);

  const table = useReactTable<MemberRow>({
    data: items,
    columns,
    getRowId: (r) => r.member.id,
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
          <Link href="./member/new">
            <PlusIcon className="w-4 h-4" />
            {t('invite')}
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
            <RefreshButton className="ml-2" loading={loading} onRefresh={refresh} />
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
            </Table>
          </div>
          <TablePagination table={table} total={total} />
        </div>
      </PageLayoutContent>

      <Dialog open={!!deleting} onOpenChange={(o) => !o && setDeleting(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('removeTitle')}</DialogTitle>
            <DialogDescription>
              {t('removeConfirm', { name: deleting ? displayName(deleting) : '' })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setDeleting(null)}>
              {tc('cancel')}
            </Button>
            <Button variant="destructive" disabled={deletingBusy} onClick={handleDelete}>
              {deletingBusy ? t('removing') : t('remove')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  );
}
