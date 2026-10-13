'use client';

import { useState } from 'react';

import { User } from '@sentinez/proto/sentinez/apps/iam/v1/model';
import { ControlPlane } from '@sentinez/proto/sentinez/types/v1/known';
import { Avatar, AvatarFallback, AvatarImage } from '@sentinez/ui/components/avatar';
import { Button } from '@sentinez/ui/components/button';
import IsLoading from '@sentinez/ui/components/common/loading';
import { Input } from '@sentinez/ui/components/input';
import { useFormatter, useTranslations } from 'next-intl';

import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useLocalUser } from '@/components/user-menu';
import { useApi } from '@/hooks/use-api';
import { GetUser, UpdateUser } from '@/lib/api/iam/user';
import { handleStatusError } from '@/lib/error-handler';
import { storeUser } from '@/lib/session';
import { toast } from '@/lib/toast';
import { UserDataLocal } from '@/lib/type/common';

import { InfoCard, InfoRow } from '../components';

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
function NameRow({ name, onSaved }: Readonly<{ name: string; onSaved: (user?: User) => void }>) {
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
        <span className="min-w-0 break-words">{name || tc('notAvailable')}</span>
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
  const createdAt = user?.metadata?.createdAt ?? dateOrUndefined(local.createdAt);
  const updatedAt = user?.metadata?.updatedAt ?? dateOrUndefined(local.updatedAt);
  const na = tc('notAvailable');

  const dateOf = (date?: Date) =>
    date ? format.dateTime(date, { dateStyle: 'medium', timeStyle: 'short' }) : na;

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

  return (
    <PageLayout>
      <PageLayoutHeader title={t('accountTitle')} subtitle={t('accountSubtitle')} />
      <PageLayoutContent>
        <div className="mx-auto w-full max-w-3xl py-4">
          <InfoCard
            title={
              <div className="flex items-center gap-3">
                <Avatar className="size-12 rounded-lg">
                  <AvatarImage src={local.avatar || undefined} alt="" />
                  <AvatarFallback className="rounded-lg">
                    {(name || email).charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="grid min-w-0">
                  <span className="truncate">{name || na}</span>
                  <span className="text-muted-foreground truncate text-sm font-normal">
                    {email}
                  </span>
                </div>
              </div>
            }
          >
            <InfoRow label={t('fullName')}>
              <NameRow name={name} onSaved={onNameSaved} />
            </InfoRow>
            <InfoRow label={t('email')}>{email || na}</InfoRow>
            <InfoRow label={t('userId')} mono>
              {user?.id || local.id || na}
            </InfoRow>
            <InfoRow label={t('controlPlane')}>
              {controlPlaneLabel(user?.controlPlane ?? local.controlPlane)}
            </InfoRow>
            <InfoRow label={t('createdAt')}>{dateOf(createdAt)}</InfoRow>
            <InfoRow label={t('updatedAt')}>{dateOf(updatedAt)}</InfoRow>
          </InfoCard>
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
