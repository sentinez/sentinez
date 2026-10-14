'use client';

import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { RateLimit } from '@sentinez/proto/sentinez/apps/security/v1/model';
import { ActionType } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { useTranslations } from 'next-intl';
import { useMemo } from 'react';
import { useSecurityOptions } from '@/hooks/use-security-options';
import {
  SecRuleFields,
  SecRuleFormValue,
  ValidationTranslator,
  createEmptySecRuleForm,
  secRuleFormOf,
  secRuleOf,
  validateSecRuleForm,
} from './sec-rule-fields';

export interface RateLimitFormValue extends SecRuleFormValue {
  /** Go duration string, e.g. "10s" */
  timeWindow: string;
  maxRequests: number;
  /** Go duration string; empty means no extra block time */
  timeout: string;
}

// Positive Go duration, matching the proto pattern without the sign
const DURATION_RE = /^(?:\d+(?:\.\d+)?(?:ns|us|µs|ms|s|m|h))+$/;

export function createEmptyRateLimitForm(): RateLimitFormValue {
  return { ...createEmptySecRuleForm(), timeWindow: '1m', maxRequests: 100, timeout: '' };
}

/** RateLimit (API) -> form value */
export function rateLimitFormOf(rl: RateLimit): RateLimitFormValue {
  return {
    ...secRuleFormOf(rl),
    timeWindow: rl.timeWindow,
    maxRequests: rl.maxRequests,
    timeout: rl.timeout,
  };
}

/** Form value -> RateLimit (API) */
export function rateLimitOf(f: RateLimitFormValue, id = ''): RateLimit {
  return {
    ...secRuleOf(f, id),
    timeWindow: f.timeWindow.trim(),
    maxRequests: f.maxRequests,
    timeout: f.timeout.trim(),
  };
}

/** Returns an error message, or null when the form is valid */
export function validateRateLimitForm(
  f: RateLimitFormValue,
  t: ValidationTranslator,
): string | null {
  const error = validateSecRuleForm(f, t);
  if (error) return error;
  if (!DURATION_RE.test(f.timeWindow.trim())) {
    return t('timeWindow');
  }
  if (!Number.isInteger(f.maxRequests) || f.maxRequests <= 0) {
    return t('maxRequests');
  }
  if (f.timeout.trim() && !DURATION_RE.test(f.timeout.trim())) {
    return t('blockTimeout');
  }
  return null;
}

interface Props {
  value: RateLimitFormValue;
  onChange: (patch: Partial<RateLimitFormValue>) => void;
  idPrefix?: string;
}

/** Form fields matching v1RateLimit in security.swagger.json */
export function RateLimitFields({ value, onChange, idPrefix = 'ratelimit' }: Props) {
  const t = useTranslations('RateLimitFields');
  const { actionTypeOptions } = useSecurityOptions();
  // The edge answers 429 once a limiter is exceeded, so only these make sense
  const actionOptions = useMemo(
    () =>
      actionTypeOptions
        .filter((o) => [ActionType.ACTION_TYPE_BLOCK, ActionType.ACTION_TYPE_LOG].includes(o.value))
        .map((o) =>
          o.value === ActionType.ACTION_TYPE_BLOCK
            ? { ...o, description: t('blockDescription') }
            : o,
        ),
    [actionTypeOptions, t],
  );

  return (
    <div className="flex flex-col gap-4">
      <SecRuleFields
        value={value}
        onChange={onChange}
        idPrefix={idPrefix}
        actionOptions={actionOptions}
      />

      <div className="grid gap-2 mt-2">
        <Label>{t('rateLimit')}</Label>
        <p className="text-sm text-muted-foreground">{t('rateLimitHint')}</p>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div className="grid gap-2 content-start">
            <Label htmlFor={`${idPrefix}-max-requests`}>{t('maxRequests')}</Label>
            <Input
              id={`${idPrefix}-max-requests`}
              type="number"
              inputMode="numeric"
              min={1}
              step={1}
              value={Number.isNaN(value.maxRequests) ? '' : value.maxRequests}
              onChange={(e) => onChange({ maxRequests: e.target.valueAsNumber })}
            />
          </div>
          <div className="grid gap-2 content-start">
            <Label htmlFor={`${idPrefix}-time-window`}>{t('timeWindow')}</Label>
            <Input
              id={`${idPrefix}-time-window`}
              placeholder="1m"
              spellCheck={false}
              autoComplete="off"
              value={value.timeWindow}
              onChange={(e) => onChange({ timeWindow: e.target.value })}
            />
            <p className="text-sm text-muted-foreground">{t('timeWindowHint')}</p>
          </div>
          <div className="grid gap-2 content-start">
            <Label htmlFor={`${idPrefix}-timeout`}>{t('blockTimeout')}</Label>
            <Input
              id={`${idPrefix}-timeout`}
              placeholder={t('blockTimeoutPlaceholder')}
              spellCheck={false}
              autoComplete="off"
              value={value.timeout}
              onChange={(e) => onChange({ timeout: e.target.value })}
            />
            <p className="text-sm text-muted-foreground">{t('blockTimeoutHint')}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
