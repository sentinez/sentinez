'use client';

import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Button } from '@sentinez/ui/components/button';
import { toast } from '@/lib/toast';
import { handleStatusError } from '@/lib/error-handler';
import {
  CaCertificateFields,
  CaCertificateFormValue,
  caCertificateRequestOf,
  createEmptyCaCertificateForm,
  validateCaCertificateForm,
} from '../components';
import { createCaCertificate } from '@/lib/api/tenant';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useResourceId } from '@/hooks/use-resource-id';
import { useState } from 'react';

export default function CreateCaCertificatePage() {
  const t = useTranslations('CaCertificate');
  const tc = useTranslations('Common');
  const tv = useTranslations('Validation');
  const te = useTranslations('Errors');
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  // The certificates belong to the resource of this domain
  const { resourceId } = useResourceId();
  const [form, setForm] = useState<CaCertificateFormValue>(createEmptyCaCertificateForm);
  const patchForm = (patch: Partial<CaCertificateFormValue>) =>
    setForm((f) => ({ ...f, ...patch }));

  const handleCreate = async () => {
    const error = validateCaCertificateForm(form, tv);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      const res = await createCaCertificate(caCertificateRequestOf(form, resourceId));
      toast.success(t('created', { count: res.caCertificates.length }));
      router.back();
    } catch (err) {
      // the server tells why: expired, not a CA, already exists, key mismatch...
      handleStatusError(err, te);
    } finally {
      setSaving(false);
    }
  };

  return (
    <PageLayout>
      <PageLayoutHeader title={t('createTitle')} subtitle={t('createSubtitle')}>
        <div className="flex justify-start gap-2">
          <Button disabled={saving || !resourceId} onClick={handleCreate}>
            {saving ? tc('saving') : tc('save')}
          </Button>
          <Button variant="secondary" onClick={() => router.back()}>
            {tc('cancel')}
          </Button>
        </div>
      </PageLayoutHeader>
      <PageLayoutContent>
        <div className="max-w-3xl mx-auto py-4">
          <CaCertificateFields value={form} onChange={patchForm} />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
