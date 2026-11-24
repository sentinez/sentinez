'use client';

import * as React from 'react';
import { BracesIcon, ChevronDownIcon, PlusIcon, Trash2Icon } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { Button } from '@sentinez/ui/components/button';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@sentinez/ui/components/collapsible';
import { Input } from '@sentinez/ui/components/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import { useControllableState } from '@sentinez/ui/hooks/use-controllable-state';
import { cn } from '@sentinez/ui/lib/utils';

import {
  AndCondition,
  Condition,
  ConditionSchema,
  Expression,
  Rule,
  FieldSource,
  Operator,
  OperatorSpec,
  ValueKind,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { OPERATOR_OPTIONS } from '@/lib/type/security';
import {
  acceptsKey,
  conditionSpec,
  conformCondition,
  keyIsOptional,
  listToText,
  textToList,
} from '@/lib/type/rule-schema';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { CodeBlock } from '@/components/code-block';

// ─── Factories ───────────────────────────────────────────────────────────────

const generateId = () => Math.random().toString(36).substring(2, 9);

export const createEmptyCondition = (): Condition => ({
  id: generateId(),
  source: FieldSource.FIELD_SOURCE_METHOD,
  key: '',
  operator: Operator.OPERATOR_EQ,
  value: '',
});

export const createEmptyRule = (): Rule => ({
  condition: createEmptyCondition(),
});

export const createEmptyAndCondition = (): AndCondition => ({
  rules: [createEmptyRule()],
  orCondition: [],
  andCondition: [],
});

export const createEmptyExpression = (): Expression => ({
  orCondition: [createEmptyAndCondition()],
});

/**
 * Sources that need a named key input (header name, query param, etc.),
 * while the condition schema of the server has not loaded.
 */
const SOURCES_WITH_KEY = new Set<FieldSource>([
  FieldSource.FIELD_SOURCE_HEADER,
  FieldSource.FIELD_SOURCE_QUERY,
  FieldSource.FIELD_SOURCE_BODY,
  FieldSource.FIELD_SOURCE_JA4,
  FieldSource.FIELD_SOURCE_TLS,
]);

// ─── Context ─────────────────────────────────────────────────────────────────

type QueryBuilderContextValue = {
  expression: Expression;
  schema?: ConditionSchema;
  addGroup: () => void;
  removeGroup: (groupIndex: number) => void;
  addRule: (groupIndex: number) => void;
  removeRule: (groupIndex: number, ruleIndex: number) => void;
  updateCondition: (
    groupIndex: number,
    ruleIndex: number,
    updater: (condition: Condition) => Condition,
  ) => void;
};

const QueryBuilderContext = React.createContext<QueryBuilderContextValue | null>(null);

export function useQueryBuilder() {
  const context = React.useContext(QueryBuilderContext);
  if (!context) {
    throw new Error('useQueryBuilder must be used within <QueryBuilder>');
  }
  return context;
}

const mapGroup = (
  expression: Expression,
  groupIndex: number,
  updater: (group: AndCondition) => AndCondition,
): Expression => ({
  ...expression,
  orCondition: expression.orCondition.map((g, i) => (i === groupIndex ? updater(g) : g)),
});

// ─── QueryBuilder (root) ─────────────────────────────────────────────────────

export type QueryBuilderProps = Omit<React.ComponentProps<'div'>, 'defaultValue'> & {
  /** Controlled expression. */
  value?: Expression;
  /** Initial expression in uncontrolled mode. */
  defaultValue?: Expression;
  onValueChange?: (value: Expression) => void;
  /**
   * The conditions the rule accepts, see useConditionSchema: the editor only
   * offers what it lists. Without it, every source and operator is offered.
   */
  schema?: ConditionSchema;
};

/**
 * Editor of a rule expression: OR of AND groups of conditions. It takes the
 * width of the form it sits in (a container query lays a condition out on one
 * row only when there is room), so it fits the narrow create / update pages.
 * Without children it shows the editor and a folded JSON preview.
 */
export function QueryBuilder({
  value: valueProp,
  defaultValue,
  onValueChange,
  schema,
  className,
  children,
  ...props
}: QueryBuilderProps) {
  const [expression, setExpression] = useControllableState<Expression>({
    prop: valueProp,
    defaultProp: defaultValue ?? createEmptyExpression(),
    onChange: onValueChange,
  });

  const context = React.useMemo<QueryBuilderContextValue>(
    () => ({
      expression,
      schema,
      addGroup: () =>
        setExpression((e) => ({
          ...e,
          orCondition: [...e.orCondition, createEmptyAndCondition()],
        })),
      removeGroup: (groupIndex) =>
        setExpression((e) => ({
          ...e,
          orCondition: e.orCondition.filter((_, i) => i !== groupIndex),
        })),
      addRule: (groupIndex) =>
        setExpression((e) =>
          mapGroup(e, groupIndex, (g) => ({ ...g, rules: [...g.rules, createEmptyRule()] })),
        ),
      removeRule: (groupIndex, ruleIndex) =>
        setExpression((e) =>
          mapGroup(e, groupIndex, (g) => ({
            ...g,
            rules: g.rules.filter((_, i) => i !== ruleIndex),
          })),
        ),
      updateCondition: (groupIndex, ruleIndex, updater) =>
        setExpression((e) =>
          mapGroup(e, groupIndex, (g) => ({
            ...g,
            rules: g.rules.map((rule, i) =>
              i === ruleIndex
                ? { ...rule, condition: updater(rule.condition ?? createEmptyCondition()) }
                : rule,
            ),
          })),
        ),
    }),
    [expression, schema, setExpression],
  );

  return (
    <QueryBuilderContext.Provider value={context}>
      <div
        data-slot="query-builder"
        className={cn('@container/query-builder flex min-w-0 flex-col gap-3', className)}
        {...props}
      >
        {children ?? (
          <>
            <QueryBuilderEditor />
            <QueryBuilderPreview />
          </>
        )}
      </div>
    </QueryBuilderContext.Provider>
  );
}

// ─── QueryBuilderEditor ──────────────────────────────────────────────────────

export function QueryBuilderEditor({ className, ...props }: React.ComponentProps<'div'>) {
  const t = useTranslations('QueryBuilder');
  const { expression, addGroup } = useQueryBuilder();

  return (
    <div
      data-slot="query-builder-editor"
      className={cn('flex min-w-0 flex-col gap-3', className)}
      {...props}
    >
      {expression.orCondition.map((_, groupIndex) => (
        <React.Fragment key={groupIndex}>
          {groupIndex > 0 && <QueryBuilderSeparator>{t('or')}</QueryBuilderSeparator>}
          <QueryBuilderGroup groupIndex={groupIndex} />
        </React.Fragment>
      ))}

      <Button
        type="button"
        variant="outline"
        size="sm"
        className="max-w-full self-start"
        onClick={addGroup}
      >
        <PlusIcon className="size-4" aria-hidden />
        <span className="truncate">{t('orGroup')}</span>
      </Button>
    </div>
  );
}

// ─── QueryBuilderSeparator ───────────────────────────────────────────────────

export function QueryBuilderSeparator({
  className,
  children,
  ...props
}: React.ComponentProps<'div'>) {
  return (
    <div
      role="separator"
      data-slot="query-builder-separator"
      className={cn('flex items-center gap-2', className)}
      {...props}
    >
      <div className="h-px flex-1 bg-border" />
      <span className="shrink-0 whitespace-nowrap rounded-full border px-2 py-0.5 text-xs font-semibold text-muted-foreground">
        {children}
      </span>
      <div className="h-px flex-1 bg-border" />
    </div>
  );
}

// ─── QueryBuilderGroup ───────────────────────────────────────────────────────

export type QueryBuilderGroupProps = React.ComponentProps<'div'> & {
  groupIndex: number;
};

export function QueryBuilderGroup({ groupIndex, className, ...props }: QueryBuilderGroupProps) {
  const t = useTranslations('QueryBuilder');
  const { expression, addRule, removeGroup } = useQueryBuilder();
  const group = expression.orCondition[groupIndex];
  if (!group) return null;

  const label = t('andGroupN', { n: groupIndex + 1 });
  const canRemove = expression.orCondition.length > 1;

  return (
    <div
      role="group"
      aria-label={label}
      data-slot="query-builder-group"
      data-empty={group.rules.length === 0 || undefined}
      className={cn('grid min-w-0 gap-3 rounded-md border p-3', className)}
      {...props}
    >
      <div className="flex min-h-8 items-center justify-between gap-2">
        <span className="min-w-0 truncate text-sm font-medium">{label}</span>
        {canRemove && (
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="size-8 shrink-0"
            aria-label={t('removeGroup', { n: groupIndex + 1 })}
            onClick={() => removeGroup(groupIndex)}
          >
            <Trash2Icon className="size-4" aria-hidden />
          </Button>
        )}
      </div>

      {group.rules.length === 0 && (
        <p className="py-2 text-center text-sm text-balance text-muted-foreground">
          {t('noConditions')}
        </p>
      )}

      {(group.orCondition.length > 0 || group.andCondition.length > 0) && (
        <p className="text-sm text-balance text-muted-foreground">{t('nestedGroups')}</p>
      )}

      {group.rules.map((rule, ruleIndex) => (
        <React.Fragment key={rule.condition?.id || ruleIndex}>
          {ruleIndex > 0 && (
            <span className="text-xs font-semibold text-muted-foreground" aria-hidden>
              {t('and')}
            </span>
          )}
          <QueryBuilderRule groupIndex={groupIndex} ruleIndex={ruleIndex} />
        </React.Fragment>
      ))}

      <Button
        type="button"
        variant="ghost"
        size="sm"
        className="max-w-full justify-self-start"
        onClick={() => addRule(groupIndex)}
      >
        <PlusIcon className="size-4" aria-hidden />
        <span className="truncate">{t('addCondition')}</span>
      </Button>
    </div>
  );
}

// ─── QueryBuilderRule ────────────────────────────────────────────────────────

export type QueryBuilderRuleProps = React.ComponentProps<'div'> & {
  groupIndex: number;
  ruleIndex: number;
};

/**
 * One condition: source, key (header, query... name), operator and value.
 * Two per row in a narrow form, all on one row when the builder is wide.
 * The condition schema of the builder, when it has one, decides which
 * sources, keys and operators it offers and how the value is entered; an
 * edit keeps the condition inside it.
 */
export function QueryBuilderRule({
  groupIndex,
  ruleIndex,
  className,
  ...props
}: QueryBuilderRuleProps) {
  const t = useTranslations('QueryBuilder');
  const { fieldSourceOptions } = useSecurityOptions();
  const { expression, schema, updateCondition, removeRule } = useQueryBuilder();
  const rule = expression.orCondition[groupIndex]?.rules[ruleIndex];
  if (!rule) return null;

  const condition = rule.condition ?? createEmptyCondition();
  const spec = conditionSpec(schema, condition);
  const needsKey = schema ? acceptsKey(spec.source) : SOURCES_WITH_KEY.has(condition.source);
  const sourceOptions = schema
    ? fieldSourceOptions.filter((o) => schema.sources.some((s) => s.source === o.value))
    : fieldSourceOptions;
  const operatorOptions = spec.variant
    ? OPERATOR_OPTIONS.filter((o) => spec.variant?.operators.some((s) => s.operator === o.value))
    : OPERATOR_OPTIONS;
  const update = (patch: Partial<Condition>) =>
    updateCondition(groupIndex, ruleIndex, (c) => conformCondition(schema, { ...c, ...patch }));

  return (
    <div
      role="group"
      aria-label={t('conditionN', { n: ruleIndex + 1 })}
      data-slot="query-builder-rule"
      data-has-key={needsKey || undefined}
      className={cn('flex min-w-0 items-start gap-1 rounded-md bg-muted/50 p-2', className)}
      {...props}
    >
      <div className="grid min-w-0 flex-1 grid-cols-1 gap-2 @xs/query-builder:grid-cols-2 @2xl/query-builder:grid-cols-4">
        <EnumSelect
          aria-label={t('source')}
          value={condition.source}
          onValueChange={(source) => update({ source: source as FieldSource, key: '' })}
          options={sourceOptions}
        />

        {needsKey && (
          <Input
            aria-label={t('key')}
            placeholder={
              keyIsOptional(spec.source) ? t('keyOptionalPlaceholder') : t('keyPlaceholder')
            }
            spellCheck={false}
            autoComplete="off"
            value={condition.key}
            onChange={(e) => update({ key: e.target.value })}
            className="min-w-0 bg-background font-mono text-xs text-ellipsis"
          />
        )}

        <EnumSelect
          aria-label={t('operator')}
          value={condition.operator}
          onValueChange={(operator) => update({ operator: operator as Operator })}
          options={operatorOptions}
        />

        <ConditionValue
          spec={spec.operator}
          value={condition.value}
          onValueChange={(value) => update({ value })}
          // without a key, the value takes the rest of the row
          className={cn(!needsKey && '@xs/query-builder:col-span-2')}
        />
      </div>

      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="size-9 shrink-0 text-muted-foreground hover:text-destructive"
        aria-label={t('removeCondition', { n: ruleIndex + 1 })}
        onClick={() => removeRule(groupIndex, ruleIndex)}
      >
        <Trash2Icon className="size-4" aria-hidden />
      </Button>
    </div>
  );
}

// ─── ConditionValue ──────────────────────────────────────────────────────────

type ConditionValueProps = {
  /** The operator of the condition in the schema, undefined without a schema */
  spec?: OperatorSpec;
  value: unknown;
  onValueChange: (value: string | string[]) => void;
  className?: string;
};

/**
 * The value input the operator takes: a select of the allowed values, a
 * comma separated list, or a text with a hint of the expected kind.
 */
function ConditionValue({ spec, value, onValueChange, className }: ConditionValueProps) {
  const t = useTranslations('QueryBuilder');
  const inputClass = cn('min-w-0 bg-background text-ellipsis', className);

  if (spec?.valueKind === ValueKind.VALUE_KIND_ENUM && !spec.list) {
    return (
      <StringSelect
        aria-label={t('value')}
        value={typeof value === 'string' ? value : ''}
        onValueChange={onValueChange}
        options={spec.enumValues}
        className={className}
      />
    );
  }

  if (spec?.list) {
    const placeholder =
      spec.valueKind === ValueKind.VALUE_KIND_ENUM
        ? t('valuePlaceholder.enumList', { values: spec.enumValues.join(', ') })
        : t('valuePlaceholder.list');
    return (
      <ListInput
        aria-label={t('value')}
        placeholder={placeholder}
        value={value}
        onValueChange={onValueChange}
        className={inputClass}
      />
    );
  }

  return (
    <Input
      aria-label={t('value')}
      placeholder={t(`valuePlaceholder.${valueHint(spec?.valueKind)}`)}
      spellCheck={false}
      autoComplete="off"
      inputMode={spec?.valueKind === ValueKind.VALUE_KIND_NUMBER ? 'decimal' : undefined}
      value={listToText(value)}
      onChange={(e) => onValueChange(e.target.value)}
      className={inputClass}
    />
  );
}

/** The placeholder key of a single value of the kind */
function valueHint(kind?: ValueKind): 'string' | 'number' | 'regex' | 'ip' | 'name' {
  switch (kind) {
    case ValueKind.VALUE_KIND_NUMBER:
      return 'number';
    case ValueKind.VALUE_KIND_REGEX:
      return 'regex';
    case ValueKind.VALUE_KIND_IP:
      return 'ip';
    case ValueKind.VALUE_KIND_NAME:
      return 'name';
    default:
      return 'string';
  }
}

type ListInputProps = Omit<React.ComponentProps<typeof Input>, 'value' | 'onChange'> & {
  value: unknown;
  onValueChange: (value: string[]) => void;
};

/**
 * A list value typed as comma separated text. The text is kept as typed
 * (a trailing comma, spaces) while the list it makes is what changes.
 */
function ListInput({ value, onValueChange, ...props }: ListInputProps) {
  const text = listToText(value);
  const [draft, setDraft] = React.useState(text);
  const [synced, setSynced] = React.useState(text);

  // the value changed from outside (another operator, a reset): show it
  if (text !== synced) {
    setSynced(text);
    if (textToList(draft).join(', ') !== text) setDraft(text);
  }

  return (
    <Input
      spellCheck={false}
      autoComplete="off"
      value={draft}
      onChange={(e) => {
        setDraft(e.target.value);
        onValueChange(textToList(e.target.value));
      }}
      {...props}
    />
  );
}

// ─── QueryBuilderPreview ─────────────────────────────────────────────────────

export type QueryBuilderPreviewProps = React.ComponentProps<typeof Collapsible>;

/** The expression as JSON, folded by default. */
export function QueryBuilderPreview({ className, ...props }: QueryBuilderPreviewProps) {
  const t = useTranslations('QueryBuilder');
  const { expression } = useQueryBuilder();

  return (
    <Collapsible
      data-slot="query-builder-preview"
      className={cn('group/preview flex min-w-0 flex-col gap-2', className)}
      {...props}
    >
      <CollapsibleTrigger asChild>
        <Button type="button" variant="ghost" size="sm" className="max-w-full self-start">
          <BracesIcon className="size-4" aria-hidden />
          <span className="truncate">{t('json')}</span>
          <ChevronDownIcon
            className="size-4 transition-transform group-data-[state=open]/preview:rotate-180"
            aria-hidden
          />
        </Button>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <CodeBlock value={expression} />
      </CollapsibleContent>
    </Collapsible>
  );
}

// ─── StringSelect ────────────────────────────────────────────────────────────

type StringSelectProps = Omit<EnumSelectProps, 'value' | 'onValueChange' | 'options'> & {
  value: string;
  onValueChange: (value: string) => void;
  options: string[];
};

/** A select of literal values, shown as they are sent (GET, true...) */
function StringSelect({ value, onValueChange, options, ...props }: StringSelectProps) {
  return (
    <EnumSelect
      {...props}
      value={options.indexOf(value)}
      onValueChange={(i) => onValueChange(options[i] ?? '')}
      options={options.map((label, value) => ({ label, value }))}
    />
  );
}

// ─── EnumSelect ──────────────────────────────────────────────────────────────

type EnumSelectProps = {
  value: number;
  onValueChange: (value: number) => void;
  options: { label: string; value: number }[];
  className?: string;
  'aria-label'?: string;
};

function EnumSelect({
  value,
  onValueChange,
  options,
  className,
  'aria-label': ariaLabel,
}: EnumSelectProps) {
  const tc = useTranslations('Common');

  return (
    <Select value={String(value)} onValueChange={(v) => onValueChange(Number(v))}>
      <SelectTrigger
        aria-label={ariaLabel}
        className={cn(
          'w-full min-w-0 bg-background',
          // the value truncates instead of widening the trigger
          '*:data-[slot=select-value]:block *:data-[slot=select-value]:min-w-0',
          '*:data-[slot=select-value]:truncate',
          className,
        )}
      >
        <SelectValue placeholder={tc('select')} />
      </SelectTrigger>
      <SelectContent>
        {options.map((opt) => (
          <SelectItem key={opt.value} value={String(opt.value)}>
            {opt.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
