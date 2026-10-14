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
import { Textarea } from '@sentinez/ui/components/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
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

const PLANS = [Plan.PLAN_FREE, Plan.PLAN_STANDARD, Plan.PLAN_PRO];

export function Tenant({ tenant }: { tenant: TenantModel }) {
  const t = useTranslations('Tenant');
  const planLabel = usePlanLabel();

  return (
    <Card className="w-full transition-colors duration-300 hover:border-primary sm:w-[300px]">
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
            {t('go')} <MoveRight />
          </Button>
        </Link>
      </CardFooter>
    </Card>
  );
}

export function TeamJoin({ onCreated }: { onCreated?: () => void }) {
  const t = useTranslations('Tenant');
  const [open, setOpen] = React.useState(false);

  return (
    <>
      <Card className="w-full max-w-sm transition-transform duration-300 sm:w-[300px] sm:hover:scale-105">
        <CardHeader>
          <CardTitle>{t('createOrJoin')}</CardTitle>
          <CardDescription>{t('createOrJoinDescription')}</CardDescription>
        </CardHeader>
        <CardContent></CardContent>
        <CardFooter className="flex justify-between">
          <Button variant="outline" className=" cursor-pointer">
            {t('join')} <Plus />
          </Button>
          <Button className=" cursor-pointer" onClick={() => setOpen(true)}>
            {t('create')} <MoveRight />
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
  const planLabel = usePlanLabel();
  const [name, setName] = React.useState('');
  const [description, setDescription] = React.useState('');
  const [plan, setPlan] = React.useState<Plan>(Plan.PLAN_FREE);
  const [creating, setCreating] = React.useState(false);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();

    if (!name.trim()) {
      toast.error(t('nameRequired'));
      return;
    }

    try {
      setCreating(true);
      await createTenant({ name: name.trim(), description: description.trim(), plan });
      toast.success(t('created'));
      setName('');
      setDescription('');
      setPlan(Plan.PLAN_FREE);
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
          <div className="grid gap-2">
            <Label htmlFor="tenant-plan">{t('plan')}</Label>
            <Select value={String(plan)} onValueChange={(v) => setPlan(Number(v) as Plan)}>
              <SelectTrigger id="tenant-plan" className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {PLANS.map((p) => (
                  <SelectItem key={p} value={String(p)}>
                    {planLabel(p)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
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
