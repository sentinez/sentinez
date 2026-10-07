'use client';

import { useRouter } from 'next/navigation';
import { Button } from '@sentinez/ui/components/button';
import { toast } from '@/lib/toast';
import {
  RateLimitFields,
  RateLimitFormValue,
  createEmptyRateLimitForm,
  rateLimitOf,
  validateRateLimitForm,
} from '../../components';
import { createRateLimit } from '@/lib/api/security';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useState } from 'react';

export default function CreateRateLimitPage() {
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<RateLimitFormValue>(createEmptyRateLimitForm);
  const patchForm = (patch: Partial<RateLimitFormValue>) => setForm((f) => ({ ...f, ...patch }));

  const handleCreate = async () => {
    const error = validateRateLimitForm(form);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      await createRateLimit(rateLimitOf(form));
      toast.success('Rate limit rule created successfully');
      router.back();
    } catch {
      toast.error('Failed to create rate limit rule');
    } finally {
      setSaving(false);
    }
  };

  return (
    <PageLayout>
      <PageLayoutHeader
        title="Create Rate Limit"
        subtitle="Limit how many matching requests a client can send in a time window."
      >
        <div className="flex justify-start gap-2">
          <Button disabled={saving} onClick={handleCreate}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
          <Button variant="secondary" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </PageLayoutHeader>
      <PageLayoutContent>
        <div className="max-w-3xl mx-auto py-4">
          <RateLimitFields value={form} onChange={patchForm} idPrefix="create" />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
