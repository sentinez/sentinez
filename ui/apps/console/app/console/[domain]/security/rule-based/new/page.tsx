'use client';

import { useRouter } from 'next/navigation';
import { Button } from '@sentinez/ui/components/button';
import { toast } from '@/lib/toast';
import {
  DEFAULT_PRIORITY,
  RuleBasedFields,
  RuleBasedFormValue,
  actionParamsOf,
  createEmptyExpression,
  validateRuleBasedForm,
} from '../../components';
import { ActionType } from '@sentinez/proto/sentinez/security/rule/v1/engine';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import { createRuleBased } from '@/lib/api/security';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useState } from 'react';

export default function CreateRuleBasedPage() {
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<RuleBasedFormValue>({
    name: '',
    description: '',
    priority: DEFAULT_PRIORITY,
    status: Status.STATUS_ACTIVE,
    action: ActionType.ACTION_TYPE_BLOCK,
    actionParams: [],
    expr: createEmptyExpression(),
  });
  const patchForm = (patch: Partial<RuleBasedFormValue>) => setForm((f) => ({ ...f, ...patch }));

  const handleCreate = async () => {
    const error = validateRuleBasedForm(form);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      await createRuleBased({
        ingressRuntime: {
          id: '',
          name: form.name,
          description: form.description,
          status: form.status,
          priority: form.priority,
          expr: form.expr,
          action: { type: form.action, params: actionParamsOf(form) },
        },
      });
      toast.success('Rule based created successfully');
      router.back();
    } catch (err: any) {
      toast.error('Failed to create rule based');
    } finally {
      setSaving(false);
    }
  };

  return (
    <PageLayout>
      <PageLayoutHeader title="Create Rule" subtitle="Define a new Web Application Firewall rule.">
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
          <RuleBasedFields value={form} onChange={patchForm} idPrefix="create" />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
