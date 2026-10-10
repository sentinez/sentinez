'use client';

import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Button } from '@sentinez/ui/components/button';
import { toast } from '@/lib/toast';
import IsLoading from '@sentinez/ui/components/common/loading';
import {
  SecRuleFields,
  SecRuleFormValue,
  createEmptySecRuleForm,
  secRuleFormOf,
  secRuleOf,
  validateSecRuleForm,
} from '../../components';
import { getSecRule, updateSecRule } from '@/lib/api/security';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useEffect, useState } from 'react';

const UPDATE_MASK = ['name', 'description', 'expr', 'action', 'status', 'priority'];

export default function EditSecRulePage({ params }: { params: Promise<{ id: string }> }) {
  const t = useTranslations('SecRule');
  const tc = useTranslations('Common');
  const tv = useTranslations('Validation');
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [id, setId] = useState<string | null>(null);

  const [form, setForm] = useState<SecRuleFormValue>(createEmptySecRuleForm);
  const patchForm = (patch: Partial<SecRuleFormValue>) => setForm((f) => ({ ...f, ...patch }));

  useEffect(() => {
    async function load() {
      try {
        const { id } = await params;
        setId(id);
        const { secRule } = await getSecRule({ id });
        if (secRule) setForm(secRuleFormOf(secRule));
      } catch (err: any) {
        toast.error(t('loadDetailFailed'));
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [params]);

  const handleSave = async () => {
    if (!id) return;
    const error = validateSecRuleForm(form, tv);
    if (error) {
      toast.error(error);
      return;
    }

    setSaving(true);
    try {
      await updateSecRule({ id, secRule: secRuleOf(form, id), updateMask: UPDATE_MASK });
      toast.success(t('updated'));
      router.back();
    } catch (err: any) {
      toast.error(t('saveFailed'));
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <IsLoading />;
  }

  return (
    <PageLayout>
      <PageLayoutHeader title={t('editTitle')} subtitle={t('editSubtitle')}>
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
          <SecRuleFields value={form} onChange={patchForm} idPrefix="edit" />
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
