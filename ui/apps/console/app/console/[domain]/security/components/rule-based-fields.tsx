'use client';

import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { Textarea } from '@sentinez/ui/components/textarea';
import { Button } from '@sentinez/ui/components/button';
import { PlusIcon, Trash2Icon } from 'lucide-react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import { ActionType, Expression } from '@sentinez/proto/sentinez/secure/rule/v1/engine';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import {
  ACTIONS_WITHOUT_VALUE,
  ACTION_TYPE_OPTIONS,
  PRIORITY_OPTIONS,
  STATUS_OPTIONS,
  SelectOption,
} from '@/lib/type/security';
import { QueryBuilder } from './query-builder';

export interface ParamRow {
  key: string;
  value: string;
}

export interface RuleBasedFormValue {
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

/** Rows -> params (proto Action.params) */
export function actionParamsOf(f: RuleBasedFormValue): Record<string, string> | undefined {
  if (ACTIONS_WITHOUT_VALUE.includes(f.action)) return undefined;
  if (MAP_ACTIONS.includes(f.action)) {
    const entries = f.actionParams
      .filter((r) => r.key.trim() && r.value.trim())
      .map((r) => [r.key.trim(), r.value] as const);
    return entries.length ? Object.fromEntries(entries) : undefined;
  }
  const v = f.actionParams[0]?.value?.trim();
  return v ? { value: v } : undefined;
}

/** params (proto Action.params) -> rows */
export function paramRowsOf(action: ActionType, params?: Record<string, any>): ParamRow[] {
  const entries = Object.entries(params ?? {});
  if (MAP_ACTIONS.includes(action)) return entries.map(([key, v]) => ({ key, value: String(v) }));
  return entries.length ? [{ key: 'value', value: String(params?.value ?? entries[0]![1]) }] : [];
}

/** Returns an error message, or null when the form is valid */
export function validateRuleBasedForm(f: RuleBasedFormValue): string | null {
  if (!f.name || !f.description) return 'Fields name and description are required';
  if (ACTIONS_WITHOUT_VALUE.includes(f.action) || actionParamsOf(f)) return null;
  return MAP_ACTIONS.includes(f.action)
    ? 'At least one key and value is required'
    : 'Action value is required';
}

interface Props {
  value: RuleBasedFormValue;
  onChange: (patch: Partial<RuleBasedFormValue>) => void;
  idPrefix?: string;
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
  const list = rows.length ? rows : [{ key: '', value: '' }];
  const update = (i: number, patch: Partial<ParamRow>) =>
    onChange(list.map((r, idx) => (idx === i ? { ...r, ...patch } : r)));
  return (
    <div className="grid gap-2">
      <Label>Action value</Label>
      {list.map((r, i) => (
        <div key={i} className="grid grid-cols-[1fr_1fr_auto] gap-2 items-center">
          <Input
            aria-label={`${keyLabel} ${i + 1}`}
            id={`${idPrefix}-key-${i}`}
            placeholder={keyPlaceholder}
            value={r.key}
            onChange={(e) => update(i, { key: e.target.value })}
          />
          <Input
            aria-label={`Value ${i + 1}`}
            placeholder={valuePlaceholder}
            value={r.value}
            onChange={(e) => update(i, { value: e.target.value })}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label="Remove"
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
          Add
        </Button>
      </div>
    </div>
  );
}

/** Form fields matching v1RuleBased in security.swagger.json */
export function RuleBasedFields({ value, onChange, idPrefix = 'rule' }: Props) {
  // keep a rule's existing custom priority selectable
  const priorityOptions = PRIORITY_OPTIONS.some((o) => o.value === value.priority)
    ? PRIORITY_OPTIONS
    : [
        ...PRIORITY_OPTIONS,
        {
          label: `Custom (${value.priority})`,
          value: value.priority,
          description: 'Custom priority.',
        },
      ];

  const needsValue = !ACTIONS_WITHOUT_VALUE.includes(value.action);
  const isHeader = value.action === ActionType.ACTION_TYPE_MODIFY_HEADER;
  const isMap = MAP_ACTIONS.includes(value.action);

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-name`}>Name</Label>
        <Input
          id={`${idPrefix}-name`}
          placeholder="Rule Name"
          value={value.name}
          onChange={(e) => onChange({ name: e.target.value })}
        />
      </div>
      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-description`}>Description</Label>
        <Textarea
          id={`${idPrefix}-description`}
          placeholder="Description"
          value={value.description}
          onChange={(e) => onChange({ description: e.target.value })}
        />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <OptionSelect
          label="Status"
          value={value.status}
          options={STATUS_OPTIONS}
          onChange={(status) => onChange({ status })}
        />
        <OptionSelect
          label="Priority"
          value={value.priority}
          options={priorityOptions}
          onChange={(priority) => onChange({ priority })}
        />
      </div>

      <div className="grid gap-2 mt-2">
        <Label>Condition Logic</Label>
        <div className="-mx-1">
          <QueryBuilder value={value.expr} onValueChange={(expr) => onChange({ expr })} />
        </div>
      </div>

      <OptionSelect
        label="Action"
        value={value.action}
        options={ACTION_TYPE_OPTIONS}
        onChange={(action) => onChange({ action, actionParams: [] })}
      />

      {needsValue &&
        (isMap ? (
          <KeyValueRows
            idPrefix={idPrefix}
            rows={value.actionParams}
            onChange={(actionParams) => onChange({ actionParams })}
            keyLabel={isHeader ? 'Header name' : 'Tag key'}
            keyPlaceholder={isHeader ? 'X-Custom-Header' : 'tag-key'}
            valuePlaceholder={isHeader ? 'header value' : 'tag value'}
          />
        ) : (
          <div className="grid gap-2">
            <Label htmlFor={`${idPrefix}-action-value`}>Action value</Label>
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
