'use client';

import axios from 'axios';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Button } from '@sentinez/ui/components/button';
import { toast } from '@/lib/toast';
import {
  MemberFields,
  MemberFormValue,
  createEmptyMemberForm,
  validateMemberForm,
} from '../components';
import { createMember } from '@/lib/api/tenant';
import { ListUsers } from '@/lib/api/iam/user';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useState } from 'react';
import { useTenantId } from '@/hooks/use-tenant-id';

export default function InviteMemberPage() {
  const t = useTranslations('Members');
  const tc = useTranslations('Common');
  const tv = useTranslations('Validation');
  const router = useRouter();
  const [tenantId] = useTenantId();
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<MemberFormValue>(createEmptyMemberForm);
  const patchForm = (patch: Partial<MemberFormValue>) => setForm((f) => ({ ...f, ...patch }));

  const handleInvite = async () => {
    const error = validateMemberForm(form, tv);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      // A member is added by user id, so find the account behind the email first
      const { users } = await ListUsers({ emails: [form.email.trim()] });
      const user = users[0];
      if (!user) {
        toast.error(t('userNotFound'));
        return;
      }
      await createMember({ userId: user.id, tenantId, roles: form.roles });
      toast.success(t('invited'));
      router.back();
    } catch (err) {
      // CreateMember answers AlreadyExists when the user is in the tenant
      const exists = axios.isAxiosError(err) && err.response?.status === 409;
      toast.error(exists ? t('alreadyMember') : t('inviteFailed'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <PageLayout>
      <PageLayoutHeader title={t('inviteTitle')} subtitle={t('inviteSubtitle')}>
        <div className="flex justify-start gap-2">
          <Button disabled={saving || !tenantId} onClick={handleInvite}>
            {saving ? t('inviting') : t('invite')}
          </Button>
          <Button variant="secondary" onClick={() => router.back()}>
            {tc('cancel')}
          </Button>
        </div>
      </PageLayoutHeader>
      <PageLayoutContent>
        <div className="max-w-3xl mx-auto py-4">
          <MemberFields value={form} onChange={patchForm} idPrefix="invite" />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
