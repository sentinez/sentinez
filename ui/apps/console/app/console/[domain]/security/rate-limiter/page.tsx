'use client';

import Title from '@/components/title';
import { Button } from '@sentinez/ui/components/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
  DialogClose,
} from '@sentinez/ui/components/dialog';
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
} from '@tanstack/react-table';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { MoreHorizontal, PlusIcon } from 'lucide-react';
import React from 'react';
import { toast } from '@/lib/toast';

import { listRuleBaseds, createRuleBased, RuleBased } from '@/lib/api/security';
import BadgeStatus from '@/components/badge-status';
import Link from 'next/link';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';
import { QueryBuilder, RuleGroup } from '../components';

export const columns: ColumnDef<RuleBased>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => <div className="w-full">Name</div>,
    cell: ({ row }) => (
      <div className="w-full truncate">
        <Link
          className="text-blue-700 font-semibold underline"
          href={`./rule-based/${row.original.id}`}
        >
          {row.getValue('name')}
        </Link>
      </div>
    ),
  },
  {
    accessorKey: 'description',
    header: () => <div>Description</div>,
    cell: ({ row }) => <div>{row.getValue('description')}</div>,
  },
  {
    accessorKey: 'status',
    header: () => <div>Status</div>,
    cell: ({ row }) => {
      let status = String(row.getValue('status') || 'disable').toLowerCase();
      if (status.includes('active')) status = 'active';
      return (
        <div className="capitalize">
          <BadgeStatus status={status as any} value={status} />
        </div>
      );
    },
  },
  {
    accessorKey: 'priority',
    header: () => <div className="w-full text-right">Priority</div>,
    cell: ({ row }) => {
      return <div className="w-full text-right">{row.getValue('priority')}</div>;
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
            <DropdownMenuItem onClick={() => navigator.clipboard.writeText(row.original.id || '')}>
              Copy Rule ID
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    ),
  },
];

export default function Page() {
  const [rules, setRules] = React.useState<RuleBased[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [open, setOpen] = React.useState(false);

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});
  // form state
  const [name, setName] = React.useState('');
  const [description, setDescription] = React.useState('');
  const [priority, setPriority] = React.useState('1');
  const [query, setQuery] = React.useState<RuleGroup | undefined>(undefined);

  const fetchRules = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await listRuleBaseds();
      setRules(data);
    } catch (err: any) {
      toast.error('Failed to load rule baseds');
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    fetchRules();
  }, [fetchRules]);

  const handleCreate = async () => {
    if (!name || !description) {
      toast.error('Fields name and description are required');
      return;
    }
    const priorityNumber = parseInt(priority, 10);
    if (isNaN(priorityNumber)) {
      toast.error('Priority must be a number');
      return;
    }

    try {
      await createRuleBased({
        name,
        description,
        status: 'STATUS_ACTIVE',
        priority: priorityNumber,
        node: query as any,
        action: {
          type: 'ACTION_TYPE_BLOCK',
        },
      });
      toast.success('Rule based created successfully');
      setOpen(false);
      setName('');
      setDescription('');
      setPriority('1');
      setQuery(undefined);
      fetchRules();
    } catch (err: any) {
      toast.error('Failed to create rule based');
    }
  };

  const table = useReactTable({
    data: rules,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  });

  return (
    <div className="">
      <Title title="Rate Limiter Rule" subtitle="Manage active security rules.">
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <PlusIcon className="w-4 h-4 mr-2" />
              Create
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-3xl max-h-[85vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Create Rule Based</DialogTitle>
              <DialogDescription>Define a new Web Application Firewall rule.</DialogDescription>
            </DialogHeader>
            <div className="py-6 flex flex-col gap-4">
              <div className="grid gap-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  placeholder="Rule Name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  placeholder="Description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="priority">Priority</Label>
                <Input
                  id="priority"
                  type="number"
                  placeholder="1"
                  value={priority}
                  onChange={(e) => setPriority(e.target.value)}
                />
              </div>
              <div className="grid gap-2 mt-2">
                <Label>Condition Logic</Label>
                <div className="-mx-1">
                  <QueryBuilder
                    key={open ? 'open' : 'closed'}
                    initialQuery={query}
                    onChange={setQuery}
                  />
                </div>
              </div>
            </div>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="secondary">Cancel</Button>
              </DialogClose>
              <Button onClick={handleCreate}>Save</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </Title>
    </div>
  );
}
