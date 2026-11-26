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
import { cn } from '@sentinez/ui/lib/utils';

import {
  Condition,
  ConditionSchema,
  Expression,
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
import {
  CondNode,
  GroupNode,
  GroupOp,
  MAX_CONDITIONS,
  MAX_GROUP_DEPTH,
  countConditions,
  createCondNode,
  createGroupNode,
  exprOf,
  isEmptyGroup,
  mapNode,
  removeNode,
  treeOf,
} from '@/lib/type/rule-tree';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { CodeBlock } from '@/components/code-block';

// ─── Factories ───────────────────────────────────────────────────────────────

/** An expression of a single empty condition */
export const createEmptyExpression = (): Expression =>
  exprOf(createGroupNode('all', [createCondNode()]));

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
  root: GroupNode;
  /** The expression the tree makes: what the form receives */
  expression: Expression;
  schema?: ConditionSchema;
  conditionCount: number;
  addCondition: (groupId: string) => void;
  addGroup: (groupId: string) => void;
  setOp: (groupId: string, op: GroupOp) => void;
  remove: (nodeId: string) => void;
  updateCondition: (condId: string, updater: (condition: Condition) => Condition) => void;
};

const QueryBuilderContext = React.createContext<QueryBuilderContextValue | null>(null);

export function useQueryBuilder() {
  const context = React.useContext(QueryBuilderContext);
  if (!context) {
    throw new Error('useQueryBuilder must be used within <QueryBuilder>');
  }
  return context;
}

const mapGroup = (root: GroupNode, id: string, fn: (g: GroupNode) => GroupNode): GroupNode =>
  mapNode(root, id, (n) => (n.kind === 'group' ? fn(n) : n));

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
 * Editor of a rule expression: a tree of groups, each matching when all or
 * any of its conditions and groups do. It takes the width of the form it
 * sits in (a container query lays a condition out on one row only when there
 * is room), so it fits the narrow create / update pages. Without children it
 * shows the editor and a folded JSON preview.
 *
 * The tree is the state, so the order and the ids of the nodes survive an
 * edit; it is only rebuilt when `value` changes to something the builder did
 * not emit itself.
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
  const [root, setRoot] = React.useState(() =>
    treeOf(valueProp ?? defaultValue ?? createEmptyExpression()),
  );
  // the last expression from or to the parent: it is the same object while
  // nothing changed, so only another value is serialised to compare
  const [emitted, setEmitted] = React.useState(valueProp);
  const onValueChangeRef = React.useRef(onValueChange);
  React.useEffect(() => {
    onValueChangeRef.current = onValueChange;
  });

  // the value changed from outside (a reset, another record): rebuild
  if (
    valueProp !== undefined &&
    valueProp !== emitted &&
    JSON.stringify(valueProp) !== JSON.stringify(emitted)
  ) {
    setRoot(treeOf(valueProp));
    setEmitted(valueProp);
  }

  const context = React.useMemo<QueryBuilderContextValue>(() => {
    const commit = (next: GroupNode) => {
      const expression = exprOf(next);
      setRoot(next);
      setEmitted(expression);
      onValueChangeRef.current?.(expression);
    };

    return {
      root,
      expression: exprOf(root),
      schema,
      conditionCount: countConditions(root),
      addCondition: (groupId) =>
        commit(
          mapGroup(root, groupId, (g) => ({ ...g, children: [...g.children, createCondNode()] })),
        ),
      addGroup: (groupId) =>
        commit(
          mapGroup(root, groupId, (g) => ({
            ...g,
            // a group opens with a condition, and the other operator: the
            // same one would only be flattened into its parent
            children: [
              ...g.children,
              createGroupNode(g.op === 'all' ? 'any' : 'all', [createCondNode()]),
            ],
          })),
        ),
      setOp: (groupId, op) => commit(mapGroup(root, groupId, (g) => ({ ...g, op }))),
      remove: (nodeId) => commit(removeNode(root, nodeId)),
      updateCondition: (condId, updater) =>
        commit(
          mapNode(root, condId, (n) =>
            n.kind === 'cond' ? { ...n, condition: updater(n.condition) } : n,
          ),
        ),
    };
  }, [root, schema]);

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
  const { root } = useQueryBuilder();

  return (
    <div
      data-slot="query-builder-editor"
      className={cn('flex min-w-0 flex-col gap-3', className)}
      {...props}
    >
      <QueryBuilderGroup node={root} depth={0} />
    </div>
  );
}

// ─── QueryBuilderGroup ───────────────────────────────────────────────────────

export type QueryBuilderGroupProps = Omit<React.ComponentProps<'div'>, 'children'> & {
  node: GroupNode;
  /** 0 for the root group */
  depth: number;
};

/** A group: the operator it joins its children with, and the children */
export function QueryBuilderGroup({ node, depth, className, ...props }: QueryBuilderGroupProps) {
  const t = useTranslations('QueryBuilder');
  const { conditionCount, addCondition, addGroup, setOp, remove } = useQueryBuilder();
  const isRoot = depth === 0;
  const canAddCondition = conditionCount < MAX_CONDITIONS;
  const canAddGroup = canAddCondition && depth < MAX_GROUP_DEPTH;
  const join = node.op === 'all' ? t('and') : t('or');

  return (
    <div
      role="group"
      aria-label={t('groupN', { n: depth + 1 })}
      data-slot="query-builder-group"
      data-op={node.op}
      data-empty={node.children.length === 0 || undefined}
      className={cn(
        'grid min-w-0 gap-3 rounded-md border p-3',
        !isRoot && 'border-l-2 border-l-primary/40',
        className,
      )}
      {...props}
    >
      <div className="flex min-h-8 items-center justify-between gap-2">
        <Select value={node.op} onValueChange={(op) => setOp(node.id, op as GroupOp)}>
          <SelectTrigger aria-label={t('operator')} className="h-8 w-auto min-w-0 bg-background">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('matchAll')}</SelectItem>
            <SelectItem value="any">{t('matchAny')}</SelectItem>
          </SelectContent>
        </Select>
        {!isRoot && (
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="size-8 shrink-0"
            aria-label={t('removeGroup')}
            onClick={() => remove(node.id)}
          >
            <Trash2Icon className="size-4" aria-hidden />
          </Button>
        )}
      </div>

      {isEmptyGroup(node) && (
        <p className="py-2 text-center text-sm text-balance text-muted-foreground">
          {isRoot ? t('noConditions') : t('emptyGroup')}
        </p>
      )}

      {node.children.map((child, i) => (
        <React.Fragment key={child.id}>
          {i > 0 && <QueryBuilderSeparator>{join}</QueryBuilderSeparator>}
          {child.kind === 'group' ? (
            <QueryBuilderGroup node={child} depth={depth + 1} />
          ) : (
            <QueryBuilderRule node={child} index={i} />
          )}
        </React.Fragment>
      ))}

      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="max-w-full"
          disabled={!canAddCondition}
          onClick={() => addCondition(node.id)}
        >
          <PlusIcon className="size-4" aria-hidden />
          <span className="truncate">{t('addCondition')}</span>
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="max-w-full"
          disabled={!canAddGroup}
          onClick={() => addGroup(node.id)}
        >
          <PlusIcon className="size-4" aria-hidden />
          <span className="truncate">{t('addGroup')}</span>
        </Button>
      </div>
      {!canAddCondition && (
        <p className="text-xs text-muted-foreground">{t('maxConditions', { n: MAX_CONDITIONS })}</p>
      )}
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

// ─── QueryBuilderRule ────────────────────────────────────────────────────────

export type QueryBuilderRuleProps = Omit<React.ComponentProps<'div'>, 'children'> & {
  node: CondNode;
  /** Position in its group, for the labels */
  index: number;
};

/**
 * One condition: source, key (header, query... name), operator and value.
 * Two per row in a narrow form, all on one row when the builder is wide.
 * The condition schema of the builder, when it has one, decides which
 * sources, keys and operators it offers and how the value is entered; an
 * edit keeps the condition inside it.
 */
export function QueryBuilderRule({ node, index, className, ...props }: QueryBuilderRuleProps) {
  const t = useTranslations('QueryBuilder');
  const { fieldSourceOptions } = useSecurityOptions();
  const { schema, updateCondition, remove } = useQueryBuilder();

  const condition = node.condition;
  const spec = conditionSpec(schema, condition);
  const needsKey = schema ? acceptsKey(spec.source) : SOURCES_WITH_KEY.has(condition.source);
  const sourceOptions = schema
    ? fieldSourceOptions.filter((o) => schema.sources.some((s) => s.source === o.value))
    : fieldSourceOptions;
  const operatorOptions = spec.variant
    ? OPERATOR_OPTIONS.filter((o) => spec.variant?.operators.some((s) => s.operator === o.value))
    : OPERATOR_OPTIONS;
  const update = (patch: Partial<Condition>) =>
    updateCondition(node.id, (c) => conformCondition(schema, { ...c, ...patch }));

  return (
    <div
      role="group"
      aria-label={t('conditionN', { n: index + 1 })}
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
          // a row of its own until the four fit on one line
          className="@xs/query-builder:col-span-2 @2xl/query-builder:col-span-1"
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
          // with a key, the value has a row of its own on two columns; without
          // one, it takes the rest of the row on four
          className={
            needsKey
              ? '@xs/query-builder:col-span-2 @2xl/query-builder:col-span-1'
              : '@2xl/query-builder:col-span-2'
          }
        />
      </div>

      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="size-9 shrink-0 text-muted-foreground hover:text-destructive"
        aria-label={t('removeCondition', { n: index + 1 })}
        onClick={() => remove(node.id)}
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
