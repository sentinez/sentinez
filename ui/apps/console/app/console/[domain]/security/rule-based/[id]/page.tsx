'use client';

import { useRouter } from 'next/navigation';
import { Button } from '@sentinez/ui/components/button';
import { toast } from '@/lib/toast';
import IsLoading from '@sentinez/ui/components/common/loading';
import {
  DEFAULT_PRIORITY,
  RuleBasedFields,
  RuleBasedFormValue,
  actionParamsOf,
  createEmptyExpression,
  paramRowsOf,
  validateRuleBasedForm,
} from '../../components';
import { ActionType } from '@sentinez/proto/sentinez/secure/rule/v1/engine';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import { getRuleBased, updateRuleBased } from '@/lib/api/security';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useEffect, useState } from 'react';

const UPDATE_MASK = 'name,description,expr,action,status,priority';

export default function EditRuleBasedPage({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [id, setId] = useState<string | null>(null);

  const [form, setForm] = useState<RuleBasedFormValue>({
    name: '',
    description: '',
    priority: DEFAULT_PRIORITY,
    actionParams: [],
    status: Status.STATUS_ACTIVE,
    action: ActionType.ACTION_TYPE_BLOCK,
    expr: createEmptyExpression(),
  });
  const patchForm = (patch: Partial<RuleBasedFormValue>) => setForm((f) => ({ ...f, ...patch }));

  useEffect(() => {
    async function load() {
      try {
        const { id } = await params;
        setId(id);
        const r = (await getRuleBased(id)).ingressRuntime;
        setForm({
          name: r?.name || '',
          description: r?.description || '',
          priority: r?.priority || DEFAULT_PRIORITY,
          status: r?.status ?? Status.STATUS_ACTIVE,
          action: r?.action?.type ?? ActionType.ACTION_TYPE_BLOCK,
          actionParams: paramRowsOf(
            r?.action?.type ?? ActionType.ACTION_TYPE_BLOCK,
            r?.action?.params,
          ),
          expr: r?.expr ?? createEmptyExpression(),
        });
      } catch (err: any) {
        toast.error('Failed to load rule details');
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [params]);

  const handleSave = async () => {
    if (!id) return;
    const error = validateRuleBasedForm(form);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      await updateRuleBased(
        id,
        {
          ingressRuntime: {
            id,
            name: form.name,
            description: form.description,
            priority: form.priority,
            status: form.status,
            expr: form.expr,
            action: { type: form.action, params: actionParamsOf(form) },
          },
        },
        UPDATE_MASK,
      );
      toast.success('Rule updated successfully');
      router.back();
    } catch (err: any) {
      toast.error('Failed to save rule');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <IsLoading />;
  }

  return (
    <PageLayout>
      <PageLayoutHeader title="Edit Rule" subtitle="Modify security rule configuration and logic.">
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
          <RuleBasedFields value={form} onChange={patchForm} idPrefix="edit" />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
