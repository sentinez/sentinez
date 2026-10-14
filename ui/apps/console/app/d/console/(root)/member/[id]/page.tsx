'use client';

import { useParams, useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Button } from '@sentinez/ui/components/button';
import IsLoading from '@sentinez/ui/components/common/loading';
import { toast } from '@/lib/toast';
import {
  MemberAvatar,
  MemberFields,
  MemberFormValue,
  memberFormOf,
  validateMemberRoles,
} from '../components';
import { getMember, updateMember } from '@/lib/api/tenant';
import { GetUser } from '@/lib/api/iam/user';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useEffect, useState } from 'react';

/** The member's user as shown above the form */
interface Profile {
  name: string;
  /** Avatar seed */
  seed: string;
  email: string;
}

export default function EditMemberPage() {
  const t = useTranslations('Members');
  const tc = useTranslations('Common');
  const tv = useTranslations('Validation');
  const router = useRouter();
  const { id } = useParams<{ id: string }>();
  const [form, setForm] = useState<MemberFormValue | null>(null);
  const [profile, setProfile] = useState<Profile>({ name: '', seed: '', email: '' });
  const [saving, setSaving] = useState(false);
  const patchForm = (patch: Partial<MemberFormValue>) =>
    setForm((f) => (f ? { ...f, ...patch } : f));

  useEffect(() => {
    const controller = new AbortController();
    const { signal } = controller;
    (async () => {
      try {
        const { member } = await getMember({ id }, { signal });
        if (!member) throw new Error('member not found');
        // The email is only shown; a failed lookup still lets the roles be edited
        const user = await GetUser({ id: member.userId }, { signal })
          .then((r) => r.user)
          .catch(() => undefined);
        setProfile({
          name: user?.fullName || user?.email || member.userId,
          // same seed as the member list and the user menu
          seed: user?.fullName || member.userId,
          email: user?.email ?? '',
        });
        setForm(memberFormOf(member, user?.email ?? ''));
      } catch {
        if (signal.aborted) return;
        toast.error(t('loadOneFailed'));
        router.back();
      }
    })();
    return () => controller.abort();
  }, [id, router, t]);

  const handleSave = async () => {
    if (!form) return;
    // The email cannot change here, so only the roles are checked
    const error = validateMemberRoles(form.roles, tv);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      await updateMember({ id, roles: form.roles });
      toast.success(t('updated'));
      router.back();
    } catch {
      toast.error(t('updateFailed'));
    } finally {
      setSaving(false);
    }
  };

  if (!form) return <IsLoading />;

  return (
    <PageLayout>
      <PageLayoutHeader title={t('editTitle', { name: profile.name })} subtitle={t('editSubtitle')}>
        <div className="flex justify-start gap-2">
          <Button disabled={saving} onClick={handleSave}>
            {saving ? tc('saving') : tc('save')}
          </Button>
          <Button variant="secondary" onClick={() => router.back()}>
            {tc('cancel')}
          </Button>
        </div>
      </PageLayoutHeader>
      <PageLayoutContent>
        <div className="max-w-3xl mx-auto py-4">
          <div className="flex items-center gap-4 mb-6">
            <MemberAvatar name={profile.name} seed={profile.seed} className="h-16 w-16" />
            <div className="grid min-w-0">
              <span className="truncate text-lg font-semibold">{profile.name}</span>
              {profile.email && (
                <span className="truncate text-sm text-muted-foreground">{profile.email}</span>
              )}
            </div>
          </div>
          <MemberFields value={form} onChange={patchForm} idPrefix="edit" emailReadOnly />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
