'use client';

import { useMemo } from 'react';
import { useTranslations } from 'next-intl';
import {
  Category,
  EngineMode,
  RuleState,
  Version,
} from '@sentinez/proto/sentinez/types/coreruleset/v1/setting';
import { ActionType, FieldSource } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import {
  ACTION_TYPES,
  ACTION_TYPE_KEY,
  CRS_CATEGORIES,
  CRS_RESPONSE_CODE,
  CRS_VERSIONS,
  ENGINE_MODES,
  FIELD_SOURCES,
  PRIORITIES,
  RULE_STATES,
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

    // ── OWASP Core Rule Set ──
    const engineModeOptions: SelectOption<EngineMode>[] = ENGINE_MODES.map(({ key, value }) => ({
      label: t(`engineMode.${key}`),
      value,
      description: t(`engineModeDescription.${key}`),
    }));
    // unspecified is treated as off by the edge
    const engineModeLabel = (mode?: EngineMode) =>
      (engineModeOptions.find((o) => o.value === mode) ?? engineModeOptions.at(-1))?.label ?? '';
    const crsVersionOptions: SelectOption<Version>[] = [
      {
        label: t('crsVersion.latest'),
        value: Version.VERSION_UNSPECIFIED,
        description: t('crsVersion.latestDescription'),
      },
      ...CRS_VERSIONS.map(({ label, value }) => ({ label, value, description: '' })),
    ];
    const crsVersionLabel = (version?: Version) =>
      (crsVersionOptions.find((o) => o.value === version) ?? crsVersionOptions[0])?.label ?? '';
    const crsCategoryOptions: {
      label: string;
      value: Category;
      code: number;
      response: boolean;
    }[] = CRS_CATEGORIES.map(({ key, code, value }) => ({
      label: t(`crsCategory.${key}`),
      value,
      code,
      response: code >= CRS_RESPONSE_CODE,
    }));
    const ruleStateOptions: SelectOption<RuleState>[] = RULE_STATES.map(({ key, value }) => ({
      label: t(`ruleState.${key}`),
      value,
      description: t(`ruleStateDescription.${key}`),
    }));

    return {
      engineModeOptions,
      engineModeLabel,
      crsVersionOptions,
      crsVersionLabel,
      crsCategoryOptions,
      ruleStateOptions,
      actionTypeLabel,
      fieldSourceOptions,
      statusOptions,
      priorityOptions,
      customPriorityOption,
      actionTypeOptions,
    };
  }, [t]);
}
