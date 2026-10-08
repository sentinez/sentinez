'use client';

import { useRouter } from 'next/navigation';
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
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<ResourceFormValue>(createEmptyResourceForm);
  const patchForm = (patch: Partial<ResourceFormValue>) => setForm((f) => ({ ...f, ...patch }));

  const handleCreate = async () => {
    const error = validateResourceForm(form);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      await createResource(resourceOf(form));
      toast.success('Resource created successfully');
      router.back();
    } catch {
      toast.error('Failed to create resource');
    } finally {
      setSaving(false);
    }
  };

  return (
    <PageLayout>
      <PageLayoutHeader
        title="Create Resource"
        subtitle="Add a domain for the edge to protect and serve."
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
          <ResourceFields value={form} onChange={patchForm} idPrefix="create" />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
