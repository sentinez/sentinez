'use client';

import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Button } from '@sentinez/ui/components/button';
import { toast } from '@/lib/toast';
import {
  ResourceFields,
  ResourceFormValue,
  createEmptyResourceForm,
  resourceOf,
  validateResourceForm,
} from '../components';
import { createResource } from '@/lib/api/tenant';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useState } from 'react';

export default function CreateResourcePage() {
  const t = useTranslations('Domain');
  const tc = useTranslations('Common');
  const tv = useTranslations('Validation');
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<ResourceFormValue>(createEmptyResourceForm);
  const patchForm = (patch: Partial<ResourceFormValue>) => setForm((f) => ({ ...f, ...patch }));

  const handleCreate = async () => {
    const error = validateResourceForm(form, tv);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      await createResource(resourceOf(form));
      toast.success(t('created'));
      router.back();
    } catch {
      toast.error(t('createFailed'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <PageLayout>
      <PageLayoutHeader title={t('createTitle')} subtitle={t('createSubtitle')}>
        <div className="flex justify-start gap-2">
          <Button disabled={saving} onClick={handleCreate}>
            {saving ? tc('saving') : tc('save')}
          </Button>
          <Button variant="secondary" onClick={() => router.back()}>
            {tc('cancel')}
          </Button>
        </div>
      </PageLayoutHeader>
      <PageLayoutContent>
        <div className="max-w-3xl mx-auto py-4">
          <ResourceFields value={form} onChange={patchForm} idPrefix="create" />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
