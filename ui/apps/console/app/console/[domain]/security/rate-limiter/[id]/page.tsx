'use client';

import { useRouter } from 'next/navigation';
import { Button } from '@sentinez/ui/components/button';
import { toast } from '@/lib/toast';
import IsLoading from '@sentinez/ui/components/common/loading';
import {
  RateLimitFields,
  RateLimitFormValue,
  createEmptyRateLimitForm,
  rateLimitFormOf,
  rateLimitOf,
  validateRateLimitForm,
} from '../../components';
import { getRateLimit, updateRateLimit } from '@/lib/api/security';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useEffect, useState } from 'react';

const UPDATE_MASK = 'name,description,expr,action,status,priority,time_window,max_requests,timeout';

export default function EditRateLimitPage({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [id, setId] = useState<string | null>(null);
  const [form, setForm] = useState<RateLimitFormValue>(createEmptyRateLimitForm);
  const patchForm = (patch: Partial<RateLimitFormValue>) => setForm((f) => ({ ...f, ...patch }));

  useEffect(() => {
    async function load() {
      try {
        const { id } = await params;
        setId(id);
        setForm(rateLimitFormOf(await getRateLimit(id)));
      } catch {
        toast.error('Failed to load rate limit details');
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [params]);

  const handleSave = async () => {
    if (!id) return;
    const error = validateRateLimitForm(form);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      await updateRateLimit(id, rateLimitOf(form, id), UPDATE_MASK);
      toast.success('Rate limit rule updated successfully');
      router.back();
    } catch {
      toast.error('Failed to save rate limit rule');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <IsLoading />;
  }

  return (
    <PageLayout>
      <PageLayoutHeader
        title="Edit Rate Limit"
        subtitle="Modify the rate limit condition, threshold and block time."
      >
        <div className="flex justify-start gap-2">
          <Button disabled={saving} onClick={handleSave}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
          <Button variant="secondary" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </PageLayoutHeader>
      <PageLayoutContent>
        <div className="max-w-3xl mx-auto py-4">
          <RateLimitFields value={form} onChange={patchForm} idPrefix="edit" />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
