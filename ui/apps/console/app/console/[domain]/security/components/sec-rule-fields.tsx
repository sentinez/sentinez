'use client';

import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { Textarea } from '@sentinez/ui/components/textarea';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon, Trash2Icon } from 'lucide-react';
import { useTranslations } from 'next-intl';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import { ActionValue, SecRule } from '@sentinez/proto/sentinez/apps/security/v1/model';
import {
  ActionType,
  Expression,
  actionTypeFromJSON,
  actionTypeToJSON,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import { ACTIONS_WITHOUT_VALUE, SelectOption } from '@/lib/type/security';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { QueryBuilder, createEmptyExpression } from './query-builder';

export interface ParamRow {
  key: string;
  value: string;
}

export interface SecRuleFormValue {
  name: string;
  description: string;
  priority: number;
  status: Status;
  action: ActionType;
  /** Action value rows: single-value actions use `rows[0].value`; Set Tag / Modify Header use key/value rows */
  actionParams: ParamRow[];
  expr: Expression;
}

export const DEFAULT_PRIORITY = 50;

const MAP_ACTIONS = [ActionType.ACTION_TYPE_SET_TAG, ActionType.ACTION_TYPE_MODIFY_HEADER];

export type ValidationTranslator = ReturnType<typeof useTranslations<'Validation'>>;

/** Rows -> ActionValue: Set Tag / Modify Header use mapValue, other actions strValue */
export function actionValueOf(f: SecRuleFormValue): ActionValue | undefined {
  if (ACTIONS_WITHOUT_VALUE.includes(f.action)) return undefined;
  if (MAP_ACTIONS.includes(f.action)) {
    const entries = f.actionParams
      .filter((r) => r.key.trim() && r.value.trim())
      .map((r) => [r.key.trim(), r.value] as const);
    return entries.length ? { strValue: '', mapValue: Object.fromEntries(entries) } : undefined;
  }
  const v = f.actionParams[0]?.value?.trim();
  return v ? { strValue: v, mapValue: {} } : undefined;
}

/** ActionValue -> rows */
export function paramRowsOf(action: ActionType, v?: ActionValue): ParamRow[] {
  const entries = Object.entries(v?.mapValue ?? {});
  if (MAP_ACTIONS.includes(action)) return entries.map(([key, value]) => ({ key, value }));
  const value = v?.strValue || entries[0]?.[1];
  return value ? [{ key: 'value', value }] : [];
}

/** SecRule.action (ActionType name) -> ActionType, defaulting to Block */
export function actionTypeOf(action?: string): ActionType {
  return action ? actionTypeFromJSON(action) : ActionType.ACTION_TYPE_BLOCK;
}

export function createEmptySecRuleForm(): SecRuleFormValue {
  return {
    name: '',
    description: '',
    priority: DEFAULT_PRIORITY,
    status: Status.STATUS_ACTIVE,
    action: ActionType.ACTION_TYPE_BLOCK,
    actionParams: [],
    expr: createEmptyExpression(),
  };
}

/** SecRule (API) -> form value */
export function secRuleFormOf(r: SecRule): SecRuleFormValue {
  const action = actionTypeOf(r.action);
  return {
    name: r.name,
    description: r.description,
    priority: r.priority || DEFAULT_PRIORITY,
    status: r.status || Status.STATUS_ACTIVE,
    action,
    actionParams: paramRowsOf(action, r.actionValue),
    expr: r.expr ?? createEmptyExpression(),
  };
}

/** Form value -> SecRule (API) */
export function secRuleOf(f: SecRuleFormValue, id = ''): SecRule {
  return {
    id,
    name: f.name,
    description: f.description,
    priority: f.priority,
    status: f.status,
    expr: f.expr,
    action: actionTypeToJSON(f.action),
    actionValue: actionValueOf(f),
  };
}

/** Returns an error message, or null when the form is valid */
export function validateSecRuleForm(f: SecRuleFormValue, t: ValidationTranslator): string | null {
  if (!f.name || !f.description) return t('nameDescriptionRequired');
  if (ACTIONS_WITHOUT_VALUE.includes(f.action) || actionValueOf(f)) return null;
  return MAP_ACTIONS.includes(f.action) ? t('keyValueRequired') : t('actionValueRequired');
}

interface Props {
  value: SecRuleFormValue;
  onChange: (patch: Partial<SecRuleFormValue>) => void;
  idPrefix?: string;
  /** Restrict the selectable actions (defaults to every action type) */
  actionOptions?: SelectOption<ActionType>[];
}

function OptionSelect<T extends number>({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: T;
  options: SelectOption<T>[];
  onChange: (v: T) => void;
}) {
  const selected = options.find((o) => o.value === value);
  return (
    <div className="grid gap-2 content-start">
      <Label>{label}</Label>
      <Select value={String(value)} onValueChange={(v) => onChange(Number(v) as T)}>
        <SelectTrigger className="w-full">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {options.map((o) => (
            <SelectItem key={o.value} value={String(o.value)}>
              {o.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {selected?.description && (
        <p className="text-sm text-muted-foreground">{selected.description}</p>
      )}
    </div>
  );
}

function KeyValueRows({
  rows,
  onChange,
  keyLabel,
  keyPlaceholder,
  valuePlaceholder,
  idPrefix,
}: {
  rows: ParamRow[];
  onChange: (rows: ParamRow[]) => void;
  keyLabel: string;
  keyPlaceholder: string;
  valuePlaceholder: string;
  idPrefix: string;
}) {
  const t = useTranslations('SecRuleFields');
  const tc = useTranslations('Common');
  const list = rows.length ? rows : [{ key: '', value: '' }];
  const update = (i: number, patch: Partial<ParamRow>) =>
    onChange(list.map((r, idx) => (idx === i ? { ...r, ...patch } : r)));
  return (
    <div className="grid gap-2">
      <Label>{t('actionValue')}</Label>
      {list.map((r, i) => (
        <div key={i} className="grid grid-cols-[1fr_1fr_auto] gap-2 items-center">
          <Input
            aria-label={t('keyN', { key: keyLabel, n: i + 1 })}
            id={`${idPrefix}-key-${i}`}
            placeholder={keyPlaceholder}
            value={r.key}
            onChange={(e) => update(i, { key: e.target.value })}
          />
          <Input
            aria-label={t('valueN', { n: i + 1 })}
            placeholder={valuePlaceholder}
            value={r.value}
            onChange={(e) => update(i, { value: e.target.value })}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={tc('remove')}
            disabled={list.length === 1 && !r.key && !r.value}
            onClick={() => onChange(list.filter((_, idx) => idx !== i))}
          >
            <Trash2Icon className="w-4 h-4" />
          </Button>
        </div>
      ))}
      <div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => onChange([...list, { key: '', value: '' }])}
        >
          <PlusIcon className="w-4 h-4" />
          {tc('add')}
        </Button>
      </div>
    </div>
  );
}

/** Form fields matching v1SecRule in security.swagger.json */
export function SecRuleFields({ value, onChange, idPrefix = 'rule', actionOptions }: Props) {
  const t = useTranslations('SecRuleFields');
  const options = useSecurityOptions();
  // keep a rule's existing custom priority selectable
  const priorityOptions = options.priorityOptions.some((o) => o.value === value.priority)
    ? options.priorityOptions
    : [...options.priorityOptions, options.customPriorityOption(value.priority)];

  const needsValue = !ACTIONS_WITHOUT_VALUE.includes(value.action);
  const isHeader = value.action === ActionType.ACTION_TYPE_MODIFY_HEADER;
  const isMap = MAP_ACTIONS.includes(value.action);

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-name`}>{t('name')}</Label>
        <Input
          id={`${idPrefix}-name`}
          placeholder={t('namePlaceholder')}
          value={value.name}
          onChange={(e) => onChange({ name: e.target.value })}
        />
      </div>
      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-description`}>{t('description')}</Label>
        <Textarea
          id={`${idPrefix}-description`}
          placeholder={t('description')}
          value={value.description}
          onChange={(e) => onChange({ description: e.target.value })}
        />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <OptionSelect
          label={t('status')}
          value={value.status}
          options={options.statusOptions}
          onChange={(status) => onChange({ status })}
        />
        <OptionSelect
          label={t('priority')}
          value={value.priority}
          options={priorityOptions}
          onChange={(priority) => onChange({ priority })}
        />
      </div>

      <div className="grid gap-2 mt-2">
        <Label>{t('conditionLogic')}</Label>
        <div className="-mx-1">
          <QueryBuilder value={value.expr} onValueChange={(expr) => onChange({ expr })} />
        </div>
      </div>

      <OptionSelect
        label={t('action')}
        value={value.action}
        options={actionOptions ?? options.actionTypeOptions}
        onChange={(action) => onChange({ action, actionParams: [] })}
      />

      {needsValue &&
        (isMap ? (
          <KeyValueRows
            idPrefix={idPrefix}
            rows={value.actionParams}
            onChange={(actionParams) => onChange({ actionParams })}
            keyLabel={isHeader ? t('headerName') : t('tagKey')}
            keyPlaceholder={isHeader ? 'X-Custom-Header' : 'tag-key'}
            valuePlaceholder={isHeader ? t('headerValuePlaceholder') : t('tagValuePlaceholder')}
          />
        ) : (
          <div className="grid gap-2">
            <Label htmlFor={`${idPrefix}-action-value`}>{t('actionValue')}</Label>
            <Input
              id={`${idPrefix}-action-value`}
              placeholder={
                value.action === ActionType.ACTION_TYPE_REDIRECT
                  ? 'https://example.com/blocked'
                  : 'upstream-name'
              }
              value={value.actionParams[0]?.value ?? ''}
              onChange={(e) =>
                onChange({ actionParams: [{ key: 'value', value: e.target.value }] })
              }
            />
          </div>
        ))}
    </div>
  );
}
