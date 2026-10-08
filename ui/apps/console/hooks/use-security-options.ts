'use client';

import { useMemo } from 'react';
import { useTranslations } from 'next-intl';
import { ActionType, FieldSource } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import {
  ACTION_TYPES,
  ACTION_TYPE_KEY,
  FIELD_SOURCES,
  PRIORITIES,
  STATUSES,
  type SelectOption,
} from '@/lib/type/security';

/** Translated labels and select options for the rule enums in lib/type/security */
export function useSecurityOptions() {
  const t = useTranslations('Enum');

  return useMemo(() => {
    const actionTypeLabel = (type: ActionType) => t(`actionType.${ACTION_TYPE_KEY[type]}`);

    const fieldSourceOptions: { label: string; value: FieldSource }[] = FIELD_SOURCES.map(
      ({ key, value }) => ({ label: t(`fieldSource.${key}`), value }),
    );
    const statusOptions = STATUSES.map(({ key, value }) => ({
      label: t(`status.${key}`),
      value,
      description: t(`statusDescription.${key}`),
    }));
    const priorityOptions: SelectOption<number>[] = PRIORITIES.map(({ key, value }) => ({
      label: t(`priority.${key}`),
      value,
      description: t(`priorityDescription.${key}`),
    }));
    const customPriorityOption = (value: number): SelectOption<number> => ({
      label: t('priority.custom', { value }),
      value,
      description: t('priorityDescription.custom'),
    });
    const actionTypeOptions: SelectOption<ActionType>[] = ACTION_TYPES.map((value) => ({
      label: actionTypeLabel(value),
      value,
      description: t(`actionTypeDescription.${ACTION_TYPE_KEY[value]}`),
    }));

    return {
      actionTypeLabel,
      fieldSourceOptions,
      statusOptions,
      priorityOptions,
      customPriorityOption,
      actionTypeOptions,
    };
  }, [t]);
}
