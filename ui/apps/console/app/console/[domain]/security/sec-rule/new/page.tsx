'use client';

import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Button } from '@sentinez/ui/components/button';
import { toast } from '@/lib/toast';
import {
  DEFAULT_PRIORITY,
  SecRuleFields,
  SecRuleFormValue,
  actionParamsOf,
  createEmptyExpression,
  validateSecRuleForm,
} from '../../components';
import { ActionType } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import { createSecRule } from '@/lib/api/security';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useState } from 'react';

export default function CreateSecRulePage() {
  const t = useTranslations('SecRule');
  const tc = useTranslations('Common');
  const tv = useTranslations('Validation');
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<SecRuleFormValue>({
    name: '',
    description: '',
    priority: DEFAULT_PRIORITY,
    status: Status.STATUS_ACTIVE,
    action: ActionType.ACTION_TYPE_BLOCK,
    actionParams: [],
    expr: createEmptyExpression(),
  });
  const patchForm = (patch: Partial<SecRuleFormValue>) => setForm((f) => ({ ...f, ...patch }));

  const handleCreate = async () => {
    const error = validateSecRuleForm(form, tv);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      await createSecRule({
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
      toast.success(t('created'));
      router.back();
    } catch (err: any) {
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
          <SecRuleFields value={form} onChange={patchForm} idPrefix="create" />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
