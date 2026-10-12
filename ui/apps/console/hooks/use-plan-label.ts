'use client';

import { Plan } from '@sentinez/proto/sentinez/types/v1/known';
import { useTranslations } from 'next-intl';

/** Translated label of a tenant plan */
export function usePlanLabel() {
  const t = useTranslations('Tenant');

  return (plan: Plan) => {
    switch (plan) {
      case Plan.PLAN_FREE:
        return t('planFree');
      case Plan.PLAN_STANDARD:
        return t('planStandard');
      case Plan.PLAN_PRO:
        return t('planPro');
      default:
        return t('planUnknown');
    }
  };
}
