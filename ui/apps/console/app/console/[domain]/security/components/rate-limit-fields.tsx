'use client';

import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { RateLimit } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { ActionType } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import { ACTION_TYPE_OPTIONS } from '@/lib/type/security';
import {
  DEFAULT_PRIORITY,
  SecRuleFields,
  SecRuleFormValue,
  actionParamsOf,
  paramRowsOf,
  validateSecRuleForm,
} from './sec-rule-fields';
import { createEmptyExpression } from './query-builder';

export interface RateLimitFormValue extends SecRuleFormValue {
  /** Go duration string, e.g. "10s" */
  timeWindow: string;
  maxRequests: number;
  /** Go duration string; empty means no extra block time */
  timeout: string;
}

// The edge answers 429 once a limiter is exceeded, so only these make sense
const RATE_LIMIT_ACTION_OPTIONS = ACTION_TYPE_OPTIONS.filter((o) =>
  [ActionType.ACTION_TYPE_BLOCK, ActionType.ACTION_TYPE_LOG].includes(o.value),
).map((o) =>
  o.value === ActionType.ACTION_TYPE_BLOCK
    ? { ...o, description: 'Reject requests over the limit with 429 Too Many Requests.' }
    : o,
);

// Positive Go duration, matching the proto pattern without the sign
const DURATION_RE = /^(?:\d+(?:\.\d+)?(?:ns|us|µs|ms|s|m|h))+$/;

export function createEmptyRateLimitForm(): RateLimitFormValue {
  return {
    name: '',
    description: '',
    priority: DEFAULT_PRIORITY,
    status: Status.STATUS_ACTIVE,
    action: ActionType.ACTION_TYPE_BLOCK,
    actionParams: [],
    expr: createEmptyExpression(),
    timeWindow: '1m',
    maxRequests: 100,
    timeout: '',
  };
}

/** RateLimit (API) -> form value */
export function rateLimitFormOf(rl: RateLimit): RateLimitFormValue {
  const r = rl.ingressRuntime;
  const action = r?.action?.type ?? ActionType.ACTION_TYPE_BLOCK;
  return {
    name: r?.name || '',
    description: r?.description || '',
    priority: r?.priority || DEFAULT_PRIORITY,
    status: r?.status ?? Status.STATUS_ACTIVE,
    action,
    actionParams: paramRowsOf(action, r?.action?.params),
    expr: r?.expr ?? createEmptyExpression(),
    timeWindow: rl.timeWindow,
    maxRequests: rl.maxRequests,
    timeout: rl.timeout,
  };
}

/** Form value -> RateLimit (API) */
export function rateLimitOf(f: RateLimitFormValue, id = ''): RateLimit {
  return {
    ingressRuntime: {
      id,
      name: f.name,
      description: f.description,
      status: f.status,
      priority: f.priority,
      expr: f.expr,
      action: { type: f.action, params: actionParamsOf(f) },
    },
    timeWindow: f.timeWindow.trim(),
    maxRequests: f.maxRequests,
    timeout: f.timeout.trim(),
  } as RateLimit;
}

/** Returns an error message, or null when the form is valid */
export function validateRateLimitForm(f: RateLimitFormValue): string | null {
  const error = validateSecRuleForm(f);
  if (error) return error;
  if (!DURATION_RE.test(f.timeWindow.trim())) {
    return 'Time window must be a positive duration, e.g. 10s, 1m';
  }
  if (!Number.isInteger(f.maxRequests) || f.maxRequests <= 0) {
    return 'Max requests must be a whole number greater than 0';
  }
  if (f.timeout.trim() && !DURATION_RE.test(f.timeout.trim())) {
    return 'Block timeout must be a positive duration, e.g. 30s, 10m';
  }
  return null;
}

/** Human-readable limit, e.g. "100 req / 1m" */
export function formatRateLimit(rl: Pick<RateLimit, 'maxRequests' | 'timeWindow'>): string {
  return `${rl.maxRequests} req / ${rl.timeWindow || '?'}`;
}

interface Props {
  value: RateLimitFormValue;
  onChange: (patch: Partial<RateLimitFormValue>) => void;
  idPrefix?: string;
}

/** Form fields matching v1RateLimit in security.swagger.json */
export function RateLimitFields({ value, onChange, idPrefix = 'ratelimit' }: Props) {
  return (
    <div className="flex flex-col gap-4">
      <SecRuleFields
        value={value}
        onChange={onChange}
        idPrefix={idPrefix}
        actionOptions={RATE_LIMIT_ACTION_OPTIONS}
      />

      <div className="grid gap-2 mt-2">
        <Label>Rate limit</Label>
        <p className="text-sm text-muted-foreground">
          Requests matching the condition are counted per client IP over a sliding window.
        </p>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div className="grid gap-2 content-start">
            <Label htmlFor={`${idPrefix}-max-requests`}>Max requests</Label>
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
            <Label htmlFor={`${idPrefix}-time-window`}>Time window</Label>
            <Input
              id={`${idPrefix}-time-window`}
              placeholder="1m"
              spellCheck={false}
              autoComplete="off"
              value={value.timeWindow}
              onChange={(e) => onChange({ timeWindow: e.target.value })}
            />
            <p className="text-sm text-muted-foreground">e.g. 10s, 1m, 1h</p>
          </div>
          <div className="grid gap-2 content-start">
            <Label htmlFor={`${idPrefix}-timeout`}>Block timeout</Label>
            <Input
              id={`${idPrefix}-timeout`}
              placeholder="Optional, e.g. 10m"
              spellCheck={false}
              autoComplete="off"
              value={value.timeout}
              onChange={(e) => onChange({ timeout: e.target.value })}
            />
            <p className="text-sm text-muted-foreground">
              How long a client stays blocked after exceeding the limit.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
