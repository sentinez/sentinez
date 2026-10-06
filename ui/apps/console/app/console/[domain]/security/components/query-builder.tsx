'use client';

import * as React from 'react';
import { Plus, Trash2 } from 'lucide-react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Button } from '@sentinez/ui/components/button';
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
  Expression,
  Rule,
  FieldSource,
  Operator,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { FIELD_SOURCE_OPTIONS, OPERATOR_OPTIONS } from '@/lib/type/security';

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
  id: generateId(),
  expr: '',
  condition: createEmptyCondition(),
});

export const createEmptyAndCondition = (): AndCondition => ({
  rules: [createEmptyRule()],
  orCondition: [],
});

export const createEmptyExpression = (): Expression => ({
  orCondition: [createEmptyAndCondition()],
});

/** Sources that need a named key input (header name, query param, etc.) */
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
  orientation?: 'vertical' | 'horizontal';
};

export function QueryBuilder({
  value: valueProp,
  defaultValue,
  onValueChange,
  orientation = 'vertical',
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
    [expression, setExpression],
  );

  return (
    <QueryBuilderContext.Provider value={context}>
      <div
        data-slot="query-builder"
        data-orientation={orientation}
        className={cn(
          'group/query-builder flex gap-4 h-full',
          'data-[orientation=horizontal]:gap-6 lg:data-[orientation=horizontal]:grid-cols-5',
          className,
        )}
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

export type QueryBuilderEditorProps = React.ComponentProps<typeof Card> & {
  title?: React.ReactNode;
  description?: React.ReactNode;
};

export function QueryBuilderEditor({
  title = 'Expression',
  description = 'Visually assemble security rule expressions.',
  className,
  ...props
}: QueryBuilderEditorProps) {
  const { expression, addGroup } = useQueryBuilder();

  return (
    <Card
      data-slot="query-builder-editor"
      className={cn(
        'w-full h-full overflow-hidden shadow-none',
        'lg:group-data-[orientation=horizontal]/query-builder:col-span-3',
        className,
      )}
      {...props}
    >
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className="p-4 flex flex-col gap-4">
        {expression.orCondition.map((_, groupIndex) => (
          <React.Fragment key={groupIndex}>
            {groupIndex > 0 && <QueryBuilderSeparator>OR</QueryBuilderSeparator>}
            <QueryBuilderGroup groupIndex={groupIndex} />
          </React.Fragment>
        ))}

        <Button variant="outline" size="sm" className="self-start" onClick={addGroup}>
          <Plus className="w-4 h-4" aria-hidden /> OR Group
        </Button>
      </CardContent>
    </Card>
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
      <div className="flex-1 h-px bg-border" />
      <span className="text-xs font-bold px-2 py-0.5 rounded bg-muted text-muted-foreground">
        {children}
      </span>
      <div className="flex-1 h-px bg-border" />
    </div>
  );
}

// ─── QueryBuilderGroup ───────────────────────────────────────────────────────

export type QueryBuilderGroupProps = React.ComponentProps<typeof Card> & {
  groupIndex: number;
};

export function QueryBuilderGroup({ groupIndex, className, ...props }: QueryBuilderGroupProps) {
  const { expression, addRule, removeGroup } = useQueryBuilder();
  const group = expression.orCondition[groupIndex];
  if (!group) return null;

  const label = `AND group ${groupIndex + 1}`;
  const canRemove = expression.orCondition.length > 1;

  return (
    <Card
      role="group"
      aria-label={label}
      data-slot="query-builder-group"
      data-empty={group.rules.length === 0 || undefined}
      className={cn('border-dashed shadow-none', className)}
      {...props}
    >
      <CardContent className="p-3 flex flex-col gap-3">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
            AND Group
          </span>
          <div className="flex-1" />
          <Button variant="outline" size="sm" onClick={() => addRule(groupIndex)}>
            <Plus className="w-4 h-4" aria-hidden /> AND
          </Button>
          {canRemove && (
            <Button
              variant="ghost"
              size="icon"
              className="text-destructive"
              aria-label={`Remove ${label}`}
              onClick={() => removeGroup(groupIndex)}
            >
              <Trash2 className="w-4 h-4" aria-hidden />
            </Button>
          )}
        </div>

        {group.rules.length === 0 && (
          <p className="text-sm text-muted-foreground text-center py-2">
            No conditions in this group.
          </p>
        )}

        {group.rules.map((rule, ruleIndex) => (
          <div key={rule.id || ruleIndex} className="flex flex-col gap-1">
            {ruleIndex > 0 && (
              <span className="text-xs font-bold text-muted-foreground px-1" aria-hidden>
                AND
              </span>
            )}
            <QueryBuilderRule groupIndex={groupIndex} ruleIndex={ruleIndex} />
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

// ─── QueryBuilderRule ────────────────────────────────────────────────────────

export type QueryBuilderRuleProps = React.ComponentProps<'div'> & {
  groupIndex: number;
  ruleIndex: number;
};

export function QueryBuilderRule({
  groupIndex,
  ruleIndex,
  className,
  ...props
}: QueryBuilderRuleProps) {
  const { expression, updateCondition, removeRule } = useQueryBuilder();
  const rule = expression.orCondition[groupIndex]?.rules[ruleIndex];
  if (!rule) return null;

  const condition = rule.condition ?? createEmptyCondition();
  const needsKey = SOURCES_WITH_KEY.has(condition.source);
  const label = `Condition ${ruleIndex + 1}`;
  const update = (patch: Partial<Condition>) =>
    updateCondition(groupIndex, ruleIndex, (c) => ({ ...c, ...patch }));

  return (
    <div
      role="group"
      aria-label={label}
      data-slot="query-builder-rule"
      data-has-key={needsKey || undefined}
      className={cn(
        'flex flex-wrap md:flex-nowrap items-center gap-2 p-3 rounded-md bg-muted/40',
        className,
      )}
      {...props}
    >
      <EnumSelect
        aria-label="Source"
        value={condition.source}
        onValueChange={(source) => update({ source: source as FieldSource, key: '' })}
        options={FIELD_SOURCE_OPTIONS}
        className="w-full md:w-44"
      />

      {needsKey && (
        <Input
          aria-label="Key"
          placeholder="Key (e.g. User-Agent)"
          value={condition.key}
          onChange={(e) => update({ key: e.target.value })}
          className="w-full md:w-40 font-mono text-xs"
        />
      )}

      <EnumSelect
        aria-label="Operator"
        value={condition.operator}
        onValueChange={(operator) => update({ operator: operator as Operator })}
        options={OPERATOR_OPTIONS}
        className="w-full md:w-32"
      />

      <Input
        aria-label="Value"
        placeholder="Value"
        value={String(condition.value ?? '')}
        onChange={(e) => update({ value: e.target.value })}
        className="w-full grow"
      />

      <Button
        variant="ghost"
        size="icon"
        className="text-destructive shrink-0"
        aria-label={`Remove ${label.toLowerCase()}`}
        onClick={() => removeRule(groupIndex, ruleIndex)}
      >
        <Trash2 className="w-4 h-4" aria-hidden />
      </Button>
    </div>
  );
}

// ─── QueryBuilderPreview ─────────────────────────────────────────────────────

export type QueryBuilderPreviewProps = React.ComponentProps<typeof Card> & {
  title?: React.ReactNode;
};

export function QueryBuilderPreview({
  title = 'JSON',
  className,
  ...props
}: QueryBuilderPreviewProps) {
  const { expression } = useQueryBuilder();

  return (
    <Card
      data-slot="query-builder-preview"
      className={cn(
        'bg-muted/20 shadow-none h-full sticky top-4 z-auto',
        'lg:group-data-[orientation=horizontal]/query-builder:col-span-2',
        className,
      )}
      {...props}
    >
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <pre className="text-xs text-muted-foreground overflow-auto max-h-150 scrollbar-thin">
          {JSON.stringify(expression, null, 2)}
        </pre>
      </CardContent>
    </Card>
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
  return (
    <Select value={String(value)} onValueChange={(v) => onValueChange(Number(v))}>
      <SelectTrigger aria-label={ariaLabel} className={cn('h-9', className)}>
        <SelectValue placeholder="Select..." />
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
