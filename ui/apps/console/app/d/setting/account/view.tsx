'use client';

import { ReactNode, useState } from 'react';

import { User } from '@sentinez/proto/sentinez/apps/iam/v1/model';
import { ControlPlane } from '@sentinez/proto/sentinez/types/v1/known';
import { Avatar, AvatarFallback, AvatarImage } from '@sentinez/ui/components/avatar';
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
import IsLoading from '@sentinez/ui/components/common/loading';
import { Input } from '@sentinez/ui/components/input';
import { cn } from '@sentinez/ui/lib/utils';
import { ShieldCheck } from 'lucide-react';
import { useFormatter, useTranslations } from 'next-intl';

import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useLocalUser } from '@/components/user-menu';
import { useApi } from '@/hooks/use-api';
import { GetUser, UpdateUser } from '@/lib/api/iam/user';
import { handleStatusError } from '@/lib/error-handler';
import { storeUser } from '@/lib/session';
import { toast } from '@/lib/toast';
import { UserDataLocal } from '@/lib/type/common';

import { AccountsCard } from './accounts';
import { mrzLinesOf, passportNumberOf } from './mrz';

/**
 * GetUser by id; a session stored before the id was kept falls back to the
 * email, and null skips the call when neither is known.
 */
function userQueryOf(local: UserDataLocal) {
  if (local.id) return { id: local.id };
  return local.email ? { email: local.email } : null;
}

function dateOrUndefined(iso: string) {
  return iso ? new Date(iso) : undefined;
}

/**
 * The full name with an Edit button that swaps it for an input. Saving renames
 * the signed-in user, the only field UpdateUser changes.
 */
function NameRow({
  userId,
  name,
  onSaved,
}: Readonly<{ userId: string; name: string; onSaved: (user?: User) => void }>) {
  const t = useTranslations('Setting');
  const tc = useTranslations('Common');
  const tv = useTranslations('Validation');
  const te = useTranslations('Errors');
  const [draft, setDraft] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const save = async () => {
    const fullName = (draft ?? '').trim();
    if (!fullName) return toast.error(tv('nameRequired'));
    if (fullName === name) return setDraft(null);

    setSaving(true);
    try {
      const res = await UpdateUser({ fullName });
      onSaved(res.user);
      toast.success(t('nameUpdated'));
      setDraft(null);
    } catch (err) {
      handleStatusError(err, te);
    } finally {
      setSaving(false);
    }
  };

  if (draft === null) {
    return (
      <div className="flex items-center justify-between gap-2">
        <span className="min-w-0 font-medium wrap-break-word">{name || tc('notAvailable')}</span>
        <Button size="sm" variant="outline" onClick={() => setDraft(name)}>
          {t('editName')}
        </Button>
      </div>
    );
  }

  return (
    <form
      className="flex flex-wrap items-center gap-2"
      onSubmit={(e) => {
        e.preventDefault();
        save();
      }}
    >
      <Input
        autoFocus
        aria-label={t('fullName')}
        className="min-w-0 flex-1"
        value={draft}
        disabled={saving}
        onChange={(e) => setDraft(e.target.value)}
      />
      <Button size="sm" type="submit" disabled={saving}>
        {saving ? tc('saving') : tc('save')}
      </Button>
      <Button
        size="sm"
        type="button"
        variant="outline"
        disabled={saving}
        onClick={() => setDraft(null)}
      >
        {tc('cancel')}
      </Button>
    </form>
  );
}

/** A caption over its value; `mono` is for ids */
function Field({
  label,
  mono,
  className,
  children,
}: Readonly<{ label: string; mono?: boolean; className?: string; children: ReactNode }>) {
  return (
    <div className={cn('grid min-w-0 content-start gap-1', className)}>
      <dt className="text-muted-foreground text-xs">{label}</dt>
      <dd className={cn('text-sm wrap-break-word', mono && 'font-mono break-all')}>{children}</dd>
    </div>
  );
}

export default function View() {
  const t = useTranslations('Setting');
  const tc = useTranslations('Common');
  const format = useFormatter();
  const local = useLocalUser();
  const { data, isLoading, refresh } = useApi(GetUser, userQueryOf(local));

  // the stored copy feeds the sidebar; the refetch brings the new update time
  const onNameSaved = (saved?: User) => {
    if (saved) storeUser(saved);
    refresh();
  };

  if (isLoading && !data) return <IsLoading />;

  // The copy stored at login still fills the page when the user cannot be fetched
  const user = data?.user;
  const name = user?.fullName || local.name;
  const email = user?.email || local.email;
  const id = user?.id || local.id;
  const createdAt = user?.metadata?.createdAt ?? dateOrUndefined(local.createdAt);
  const updatedAt = user?.metadata?.updatedAt ?? dateOrUndefined(local.updatedAt);
  const na = tc('notAvailable');

  const dateOf = (date?: Date) => (date ? format.dateTime(date, { dateStyle: 'medium' }) : na);

  const controlPlaneLabel = (controlPlane?: ControlPlane) => {
    switch (controlPlane) {
      case ControlPlane.CONTROL_PLANE_PORTAL:
        return t('controlPlanePortal');
      case ControlPlane.CONTROL_PLANE_ADMIN:
        return t('controlPlaneAdmin');
      default:
        return na;
    }
  };
  const controlPlane = controlPlaneLabel(user?.controlPlane ?? local.controlPlane);

  return (
    <PageLayout>
      <PageLayoutHeader title={t('accountTitle')} subtitle={t('accountSubtitle')} />
      <PageLayoutContent>
        <div className="mx-auto grid w-full max-w-3xl gap-6 py-4">
          <Card>
            <CardHeader className="border-b">
              <CardTitle className="flex items-center gap-2">
                <ShieldCheck className="text-primary size-5" aria-hidden />
                {t('passportTitle')}
              </CardTitle>
              <CardDescription>Sentinéz</CardDescription>
              {id && (
                <CardAction>
                  <Badge variant="outline" className="font-mono">
                    {passportNumberOf(id)}
                  </Badge>
                </CardAction>
              )}
            </CardHeader>
            <CardContent className="grid gap-6 sm:grid-cols-[8rem_minmax(0,1fr)]">
              <Avatar className="h-40 w-32 justify-self-center rounded-md border">
                <AvatarImage src={local.avatar || undefined} alt="" className="object-cover" />
                <AvatarFallback className="rounded-md text-4xl">
                  {(name || email).charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <dl className="grid content-start gap-4 sm:grid-cols-2">
                <Field label={t('fullName')} className="sm:col-span-2">
                  <NameRow userId={id} name={name} onSaved={onNameSaved} />
                </Field>
                <Field label={t('email')}>{email || na}</Field>
                <Field label={t('controlPlane')}>{controlPlane}</Field>
                <Field label={t('createdAt')}>{dateOf(createdAt)}</Field>
                <Field label={t('updatedAt')}>{dateOf(updatedAt)}</Field>
                <Field label={t('userId')} mono className="sm:col-span-2">
                  {id || na}
                </Field>
              </dl>
            </CardContent>
            <CardFooter className="bg-muted/50 overflow-x-auto border-t">
              <pre className="text-muted-foreground font-mono text-xs leading-relaxed tracking-widest">
                {mrzLinesOf({ name, id, issuedAt: createdAt, updatedAt }).join('\n')}
              </pre>
            </CardFooter>
          </Card>
          <AccountsCard userId={id} />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
