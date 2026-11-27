'use client';

import { ReactNode, useState } from 'react';

import type { Tenant } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { Plan } from '@sentinez/proto/sentinez/types/v1/known';
import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
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
import { GlobeIcon, UsersIcon } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useFormatter, useTranslations } from 'next-intl';

import { CopyButton } from '@/components/copy-button';
import { InfoField } from '@/components/info-card';
import { useApi } from '@/hooks/use-api';
import { usePlanLabel } from '@/hooks/use-plan-label';
import { selectTenant } from '@/hooks/use-tenant-id';
import { deleteTenant, listMembers, listResources, updateTenant } from '@/lib/api/tenant';
import { handleStatusError } from '@/lib/error-handler';
import { toast } from '@/lib/toast';

// The limits of apps.tenant.v1.Tenant (buf.validate)
const NAME_MAX = 64;
const DESCRIPTION_MAX = 255;
// A list call that only reads its total
const COUNT_PAGE = { index: 1, size: 1 };

// The hint of each plan, a key of Setting
const PLAN_HINT: Partial<Record<Plan, 'planFreeHint' | 'planStandardHint' | 'planProHint'>> = {
  [Plan.PLAN_FREE]: 'planFreeHint',
  [Plan.PLAN_STANDARD]: 'planStandardHint',
  [Plan.PLAN_PRO]: 'planProHint',
};

/** The name and description of the tenant, saved together */
export function ProfileCard({ tenant, onSaved }: { tenant: Tenant; onSaved: () => void }) {
  const t = useTranslations('Setting');
  const tt = useTranslations('Tenant');
  const tc = useTranslations('Common');
  const te = useTranslations('Errors');
  const planLabel = usePlanLabel();
  const [name, setName] = useState(tenant.name);
  const [description, setDescription] = useState(tenant.description);
  const [saving, setSaving] = useState(false);
  const changed = name !== tenant.name || description !== tenant.description;

  const save = async () => {
    if (!name.trim()) return toast.error(tt('nameRequired'));
    setSaving(true);
    try {
      await updateTenant({
        id: tenant.id,
        tenant: { ...tenant, name: name.trim(), description: description.trim() },
        updateMask: ['name', 'description'],
      });
      toast.success(t('tenantUpdated'));
      onSaved();
    } catch (err) {
      handleStatusError(err, te);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card>
      <CardHeader className="border-b">
        <CardTitle>{t('profileTitle')}</CardTitle>
        <CardDescription>{t('profileDescription')}</CardDescription>
        <CardAction>
          <Badge variant="secondary">{planLabel(tenant.plan)}</Badge>
        </CardAction>
      </CardHeader>
      <form
        id="tenant-profile"
        className="contents"
        onSubmit={(e) => {
          e.preventDefault();
          save();
        }}
      >
        <CardContent className="grid gap-4">
          <div className="grid gap-2">
            <Label htmlFor="tenant-name">{tt('name')}</Label>
            <Input
              id="tenant-name"
              maxLength={NAME_MAX}
              autoComplete="off"
              value={name}
              disabled={saving}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="tenant-description">{tt('description')}</Label>
            <Textarea
              id="tenant-description"
              maxLength={DESCRIPTION_MAX}
              rows={3}
              placeholder={t('descriptionPlaceholder')}
              value={description}
              disabled={saving}
              onChange={(e) => setDescription(e.target.value)}
            />
            <p className="text-right text-xs text-muted-foreground tabular-nums">
              {description.length}/{DESCRIPTION_MAX}
            </p>
          </div>
        </CardContent>
        <CardFooter className="justify-between gap-2 border-t">
          <p className="text-sm text-muted-foreground">{t('profileHint')}</p>
          <div className="flex shrink-0 gap-2">
            {changed && (
              <Button
                type="button"
                variant="outline"
                disabled={saving}
                onClick={() => {
                  setName(tenant.name);
                  setDescription(tenant.description);
                }}
              >
                {tc('cancel')}
              </Button>
            )}
            <Button type="submit" disabled={!changed || saving}>
              {saving ? tc('saving') : tc('save')}
            </Button>
          </div>
        </CardFooter>
      </form>
    </Card>
  );
}

/** A count of the tenant, linking to its page */
function Stat({
  href,
  icon,
  label,
  count,
}: {
  href: string;
  icon: ReactNode;
  label: string;
  count?: number;
}) {
  const format = useFormatter();
  return (
    <Link
      href={href}
      className="flex items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-muted/50"
    >
      <span className="flex size-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary [&_svg]:size-4">
        {icon}
      </span>
      <span className="grid gap-0.5">
        <span className="text-xs text-muted-foreground">{label}</span>
        {count === undefined ? (
          <Skeleton className="h-5 w-8" />
        ) : (
          <span className="text-lg leading-5 font-semibold tabular-nums">
            {format.number(count)}
          </span>
        )}
      </span>
    </Link>
  );
}

/** What the tenant holds (domains, members), its plan, id and dates */
export function OverviewCard({ tenant }: { tenant: Tenant }) {
  const t = useTranslations('Setting');
  const tt = useTranslations('Tenant');
  const tc = useTranslations('Common');
  const format = useFormatter();
  const planLabel = usePlanLabel();
  const domains = useApi(listResources, { page: COUNT_PAGE });
  const members = useApi(listMembers, { page: COUNT_PAGE, tenantIds: [tenant.id] });
  const na = tc('notAvailable');
  const dateOf = (date?: Date) =>
    date ? format.dateTime(date, { dateStyle: 'medium', timeStyle: 'short' }) : na;
  const planHintKey = PLAN_HINT[tenant.plan];

  return (
    <Card>
      <CardHeader className="border-b">
        <CardTitle>{t('overviewTitle')}</CardTitle>
        <CardDescription>{t('overviewDescription')}</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-6">
        <div className="grid gap-3 sm:grid-cols-2">
          <Stat
            href="/d/console/domain"
            icon={<GlobeIcon />}
            label={t('domains')}
            count={domains.isLoading ? undefined : (domains.data?.total ?? 0)}
          />
          <Stat
            href="/d/console/member"
            icon={<UsersIcon />}
            label={t('members')}
            count={members.isLoading ? undefined : (members.data?.total ?? 0)}
          />
        </div>
        <dl className="grid gap-4 sm:grid-cols-2">
          <InfoField label={tt('plan')} className="sm:col-span-2">
            <span className="font-medium">{planLabel(tenant.plan)}</span>
            {planHintKey && <span className="text-muted-foreground"> · {t(planHintKey)}</span>}
          </InfoField>
          <InfoField label={t('tenantId')} mono className="sm:col-span-2">
            <span className="flex items-center gap-1">
              <span className="min-w-0 break-all">{tenant.id}</span>
              <CopyButton text={tenant.id} label={t('copyTenantId')} />
            </span>
          </InfoField>
          <InfoField label={t('createdAt')}>{dateOf(tenant.metadata?.createdAt)}</InfoField>
          <InfoField label={t('updatedAt')}>{dateOf(tenant.metadata?.updatedAt)}</InfoField>
        </dl>
      </CardContent>
    </Card>
  );
}

/**
 * Deletes the tenant once its name is typed back; the console then leaves
 * it for the tenant picker.
 */
export function DangerCard({ tenant }: { tenant: Tenant }) {
  const t = useTranslations('Setting');
  const tc = useTranslations('Common');
  const te = useTranslations('Errors');
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [typed, setTyped] = useState('');
  const [deleting, setDeleting] = useState(false);
  const name = tenant.name || tenant.id;

  const remove = async () => {
    setDeleting(true);
    try {
      await deleteTenant({ id: tenant.id });
      toast.success(t('tenantDeleted', { name }));
      selectTenant('');
      router.replace('/tenant');
    } catch (err) {
      handleStatusError(err, te);
      setDeleting(false);
    }
  };

  return (
    <Card className="ring-destructive/30">
      <CardHeader>
        <CardTitle className="text-destructive">{t('dangerTitle')}</CardTitle>
        <CardDescription>{t('dangerDescription')}</CardDescription>
        <CardAction>
          <Button
            variant="destructive"
            onClick={() => {
              setTyped('');
              setOpen(true);
            }}
          >
            {t('deleteTenant')}
          </Button>
        </CardAction>
      </CardHeader>

      <Dialog open={open} onOpenChange={(o) => !deleting && setOpen(o)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('deleteTenantTitle')}</DialogTitle>
            <DialogDescription>{t('deleteTenantConfirm', { name })}</DialogDescription>
          </DialogHeader>
          <form
            className="grid gap-4"
            onSubmit={(e) => {
              e.preventDefault();
              if (typed === name) remove();
            }}
          >
            <div className="grid gap-2">
              <Label htmlFor="tenant-delete-name">{t('deleteTenantType', { name })}</Label>
              <Input
                id="tenant-delete-name"
                autoComplete="off"
                autoFocus
                value={typed}
                disabled={deleting}
                onChange={(e) => setTyped(e.target.value)}
              />
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="secondary"
                disabled={deleting}
                onClick={() => setOpen(false)}
              >
                {tc('cancel')}
              </Button>
              <Button type="submit" variant="destructive" disabled={typed !== name || deleting}>
                {deleting ? tc('deleting') : t('deleteTenant')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
