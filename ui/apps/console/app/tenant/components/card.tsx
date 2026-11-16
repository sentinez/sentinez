'use client';

import * as React from 'react';

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@sentinez/ui/components/dialog';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { Skeleton } from '@sentinez/ui/components/skeleton';
import { Textarea } from '@sentinez/ui/components/textarea';
import { cn } from '@sentinez/ui/lib/utils';
import type { Tenant as TenantModel } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { Plan } from '@sentinez/proto/sentinez/types/v1/known';
import { MoveRight, Plus } from 'lucide-react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';

import { usePlanLabel } from '@/hooks/use-plan-label';
import { selectTenant } from '@/hooks/use-tenant-id';
import { createTenant } from '@/lib/api/tenant';
import { handleStatusError } from '@/lib/error-handler';
import { toast } from '@/lib/toast';

// Mirrors the buf.validate rules of CreateTenantRequest
const NAME_MAX_LENGTH = 64;
const DESCRIPTION_MAX_LENGTH = 255;

// Hover effect shared by the cards: lift, glow and a tinted border
const CARD_HOVER =
  'group relative overflow-hidden transition-[translate,box-shadow,border-color] duration-300 ease-out hover:border-primary/50 hover:shadow-[0_12px_32px_-12px] hover:shadow-primary/35 motion-safe:hover:-translate-y-1';
const ARROW_HOVER = 'transition-transform duration-300 group-hover:translate-x-1';

export function Tenant({ tenant }: { tenant: TenantModel }) {
  const t = useTranslations('Tenant');
  const planLabel = usePlanLabel();

  return (
    <Card className={cn(CARD_HOVER, 'w-full sm:w-[300px]')}>
      {/* Accent line that slides in on hover */}
      <span
        aria-hidden
        className="from-primary pointer-events-none absolute inset-x-0 top-0 h-0.5 origin-left scale-x-0 bg-linear-to-r to-sky-400 transition-transform duration-500 ease-out group-hover:scale-x-100"
      />
      <CardHeader className="grid-cols-[minmax(0,1fr)]">
        <CardTitle className="truncate" title={tenant.name || tenant.id}>
          {tenant.name || tenant.id}
        </CardTitle>
        <CardDescription className="truncate" title={tenant.description || undefined}>
          {tenant.description || t('member')}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Badge variant="secondary">{planLabel(tenant.plan)}</Badge>
      </CardContent>
      <CardFooter className="flex justify-between">
        <Link href="/d/console/domain" className="w-full" onClick={() => selectTenant(tenant.id)}>
          <Button className="w-full cursor-pointer">
            {t('go')} <MoveRight className={ARROW_HOVER} />
          </Button>
        </Link>
      </CardFooter>
    </Card>
  );
}

/** Placeholder with the layout of a Tenant card, shown while tenants load */
export function TenantSkeleton() {
  return (
    <Card className="w-full sm:w-[300px]" aria-hidden>
      <CardHeader className="grid-cols-[minmax(0,1fr)] gap-2">
        <Skeleton className="h-4 w-2/3" />
        <Skeleton className="h-4 w-1/2" />
      </CardHeader>
      <CardContent>
        <Skeleton className="h-5 w-16" />
      </CardContent>
      <CardFooter>
        <Skeleton className="h-9 w-full" />
      </CardFooter>
    </Card>
  );
}

export function TeamJoin({ onCreated }: { onCreated?: () => void }) {
  const t = useTranslations('Tenant');
  const [open, setOpen] = React.useState(false);

  return (
    <>
      <Card
        className={cn(
          CARD_HOVER,
          'bg-card/80 w-full max-w-sm border-dashed shadow-none hover:border-solid sm:w-[300px]',
        )}
      >
        <CardHeader>
          <CardTitle>{t('createOrJoin')}</CardTitle>
          <CardDescription>{t('createOrJoinDescription')}</CardDescription>
        </CardHeader>
        <CardContent></CardContent>
        <CardFooter className="flex justify-between">
          <Button variant="outline" className="cursor-pointer">
            {t('join')} <Plus className="transition-transform duration-300 group-hover:rotate-90" />
          </Button>
          <Button className="cursor-pointer" onClick={() => setOpen(true)}>
            {t('create')} <MoveRight className={ARROW_HOVER} />
          </Button>
        </CardFooter>
      </Card>
      <CreateTenantDialog open={open} onOpenChange={setOpen} onCreated={onCreated} />
    </>
  );
}

function CreateTenantDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated?: () => void;
}) {
  const t = useTranslations('Tenant');
  const tc = useTranslations('Common');
  const te = useTranslations('Errors');
  const [name, setName] = React.useState('');
  const [description, setDescription] = React.useState('');
  const [creating, setCreating] = React.useState(false);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();

    if (!name.trim()) {
      toast.error(t('nameRequired'));
      return;
    }

    try {
      setCreating(true);
      // a tenant starts free: its plan is changed by a super admin later
      await createTenant({
        name: name.trim(),
        description: description.trim(),
        plan: Plan.PLAN_FREE,
      });
      toast.success(t('created'));
      setName('');
      setDescription('');
      onOpenChange(false);
      onCreated?.();
    } catch (error: any) {
      handleStatusError(error, te);
    } finally {
      setCreating(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={handleCreate} className="grid gap-4">
          <DialogHeader>
            <DialogTitle>{t('createTitle')}</DialogTitle>
            <DialogDescription>{t('createDescription')}</DialogDescription>
          </DialogHeader>
          <div className="grid gap-2">
            <Label htmlFor="tenant-name">{t('name')}</Label>
            <Input
              id="tenant-name"
              value={name}
              maxLength={NAME_MAX_LENGTH}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="tenant-description">{t('description')}</Label>
            <Textarea
              id="tenant-description"
              value={description}
              maxLength={DESCRIPTION_MAX_LENGTH}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              className="cursor-pointer"
              onClick={() => onOpenChange(false)}
            >
              {tc('cancel')}
            </Button>
            <Button type="submit" className="cursor-pointer" disabled={creating}>
              {creating ? t('creating') : t('create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
