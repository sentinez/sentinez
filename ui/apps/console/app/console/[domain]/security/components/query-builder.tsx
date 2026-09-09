'use client';

import React, { useState } from 'react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { Plus, Trash2 } from 'lucide-react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';

import {
  AndCondition,
  Condition,
  Expression,
  Rule,
  FieldSource,
  Operator,
} from '@sentinez/proto/sentinez/secure/rule/v1/engine';
import { FIELD_SOURCE_OPTIONS, OPERATOR_OPTIONS } from '@/lib/type/security';
import { RuleBased } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';

// ─── Helpers ─────────────────────────────────────────────────────────────────

const generateId = () => Math.random().toString(36).substring(2, 9);

const createEmptyCondition = (): Condition => ({
  id: generateId(),
  source: FieldSource.FIELD_SOURCE_METHOD,
  key: '',
  operator: Operator.OPERATOR_EQ,
  value: '',
});

const createEmptyRule = (): Rule => ({
  id: generateId(),
  name: '',
  description: '',
  condition: createEmptyCondition(),
});

const createEmptyAndCondition = (): AndCondition => ({
  rules: [createEmptyRule()],
  orCondition: [],
});

const createEmptyExpression = (): Expression => ({
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

// ─── SelectMenu ──────────────────────────────────────────────────────────────

interface SelectMenuProps {
  value: number;
  onChange: (val: number) => void;
  options: { label: string; value: number }[];
  className?: string;
}

function SelectMenu({ value, onChange, options, className = '' }: SelectMenuProps) {
  return (
    <Select value={String(value)} onValueChange={(v) => onChange(Number(v))}>
      <SelectTrigger className={`h-9 ${className}`}>
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

// ─── QueryBuilder ─────────────────────────────────────────────────────────────

interface QueryBuilderProps {
  initialQuery?: RuleBased;
  onChange?: (query: RuleBased) => void;
  layout?: 'vertical' | 'horizontal';
}

export function QueryBuilder({ initialQuery, onChange, layout = 'vertical' }: QueryBuilderProps) {
  const [query, setQuery] = useState<RuleBased>(
    initialQuery ?? {
      enable: false,
      ingress: {
        id: generateId(),
        name: '',
        description: '',
        status: 0,
        priority: 0,
        expr: createEmptyExpression(),
      },
    },
  );

  React.useEffect(() => {
    if (initialQuery) setQuery(initialQuery);
  }, [initialQuery]);

  const notifyChange = (next: RuleBased) => {
    setQuery(next);
    onChange?.(next);
  };

  const getExpr = (): Expression => query.ingress?.expr ?? createEmptyExpression();

  const updateExpr = (updater: (e: Expression) => Expression) => {
    if (!query.ingress) return;
    notifyChange({ ...query, ingress: { ...query.ingress, expr: updater(getExpr()) } });
  };

  // ── OR-group actions ──────────────────────────────────────────────────────

  const addOrGroup = () =>
    updateExpr((e) => ({
      ...e,
      orCondition: [...e.orCondition, createEmptyAndCondition()],
    }));

  const removeOrGroup = (gIdx: number) =>
    updateExpr((e) => ({
      ...e,
      orCondition: e.orCondition.filter((_, i) => i !== gIdx),
    }));

  // ── Rule actions ──────────────────────────────────────────────────────────

  const addRule = (gIdx: number) =>
    updateExpr((e) => ({
      ...e,
      orCondition: e.orCondition.map((g, i) =>
        i !== gIdx ? g : { ...g, rules: [...g.rules, createEmptyRule()] },
      ),
    }));

  const removeRule = (gIdx: number, rIdx: number) =>
    updateExpr((e) => ({
      ...e,
      orCondition: e.orCondition.map((g, i) =>
        i !== gIdx ? g : { ...g, rules: g.rules.filter((_, ri) => ri !== rIdx) },
      ),
    }));

  const updateCondition = (gIdx: number, rIdx: number, updater: (c: Condition) => Condition) =>
    updateExpr((e) => ({
      ...e,
      orCondition: e.orCondition.map((g, i) =>
        i !== gIdx
          ? g
          : {
              ...g,
              rules: g.rules.map((rule, ri) =>
                ri !== rIdx
                  ? rule
                  : {
                      ...rule,
                      condition: updater(rule.condition ?? createEmptyCondition()),
                    },
              ),
            },
      ),
    }));

  // ── Render ────────────────────────────────────────────────────────────────

  const expr = getExpr();

  return (
    <div
      className={
        layout === 'horizontal' ? 'grid grid-cols-1 lg:grid-cols-5 gap-6' : 'flex flex-col gap-4'
      }
    >
      {/* ── Builder panel ── */}
      <div className={`flex flex-col gap-4 ${layout === 'horizontal' ? 'lg:col-span-3' : ''}`}>
        <Card className="w-full overflow-hidden shadow-none">
          <CardHeader>
            <CardTitle>Expression</CardTitle>
            <CardDescription>Visually assemble security rule expressions.</CardDescription>
          </CardHeader>
          <CardContent className="p-4 flex flex-col gap-4">
            {expr.orCondition.map((andCond, gIdx) => (
              <div key={gIdx}>
                {/* OR separator */}
                {gIdx > 0 && (
                  <div className="flex items-center gap-2 mb-4">
                    <div className="flex-1 h-px bg-border" />
                    <span className="text-xs font-bold px-2 py-0.5 rounded bg-muted text-muted-foreground">
                      OR
                    </span>
                    <div className="flex-1 h-px bg-border" />
                  </div>
                )}

                <Card className="border-dashed shadow-none">
                  <CardContent className="p-3 flex flex-col gap-3">
                    {/* AND group toolbar */}
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
                        AND Group
                      </span>
                      <div className="flex-1" />
                      <Button variant="outline" size="sm" onClick={() => addRule(gIdx)}>
                        <Plus className="w-4 h-4 mr-1" /> Rule
                      </Button>
                      {expr.orCondition.length > 1 && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="text-destructive"
                          onClick={() => removeOrGroup(gIdx)}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      )}
                    </div>

                    {andCond.rules.length === 0 && (
                      <p className="text-sm text-muted-foreground text-center py-2">
                        No conditions in this group.
                      </p>
                    )}

                    {/* Rules */}
                    {andCond.rules.map((rule, rIdx) => {
                      const cond = rule.condition ?? createEmptyCondition();
                      const needsKey = SOURCES_WITH_KEY.has(cond.source);

                      return (
                        <div key={rIdx} className="flex flex-col gap-1">
                          {rIdx > 0 && (
                            <span className="text-xs font-bold text-muted-foreground px-1">
                              AND
                            </span>
                          )}
                          <div className="flex flex-wrap md:flex-nowrap items-center gap-2 p-3 rounded-md bg-muted/40">
                            {/* Source */}
                            <SelectMenu
                              value={cond.source}
                              onChange={(val) =>
                                updateCondition(gIdx, rIdx, (c) => ({
                                  ...c,
                                  source: val as FieldSource,
                                  key: '',
                                }))
                              }
                              options={FIELD_SOURCE_OPTIONS}
                              className="w-full md:w-44"
                            />

                            {/* Key (header / query param / body field) */}
                            {needsKey && (
                              <Input
                                placeholder="Key (e.g. User-Agent)"
                                value={cond.key}
                                onChange={(e) =>
                                  updateCondition(gIdx, rIdx, (c) => ({
                                    ...c,
                                    key: e.target.value,
                                  }))
                                }
                                className="w-full md:w-40 font-mono text-xs"
                              />
                            )}

                            {/* Operator */}
                            <SelectMenu
                              value={cond.operator}
                              onChange={(val) =>
                                updateCondition(gIdx, rIdx, (c) => ({
                                  ...c,
                                  operator: val as Operator,
                                }))
                              }
                              options={OPERATOR_OPTIONS}
                              className="w-full md:w-32"
                            />

                            {/* Value */}
                            <Input
                              placeholder="Value"
                              value={String(cond.value ?? '')}
                              onChange={(e) =>
                                updateCondition(gIdx, rIdx, (c) => ({
                                  ...c,
                                  value: e.target.value,
                                }))
                              }
                              className="w-full flex-grow"
                            />

                            {/* Remove rule */}
                            <Button
                              variant="ghost"
                              size="icon"
                              className="text-destructive shrink-0"
                              onClick={() => removeRule(gIdx, rIdx)}
                            >
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          </div>
                        </div>
                      );
                    })}
                  </CardContent>
                </Card>
              </div>
            ))}

            {/* Add OR group */}
            <Button variant="outline" size="sm" className="self-start" onClick={addOrGroup}>
              <Plus className="w-4 h-4 mr-1" /> OR Group
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* ── JSON preview panel ── */}
      <Card
        className={`bg-muted/20 shadow-none h-fit sticky top-4 z-auto ${
          layout === 'horizontal' ? 'lg:col-span-2' : ''
        }`}
      >
        <CardHeader>
          <CardTitle>Generated Expression JSON</CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="text-xs text-muted-foreground overflow-auto max-h-[600px] scrollbar-thin">
            {JSON.stringify(query.ingress?.expr, null, 2)}
          </pre>
        </CardContent>
      </Card>
    </div>
  );
}
