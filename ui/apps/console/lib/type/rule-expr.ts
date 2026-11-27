import {
  AndCondition,
  Condition,
  Expression,
  FieldSource,
  Operator,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';

// The editor works on the Expression of the API itself, kept flat: an OR of
// its orCondition, each an AND of its rules. Shown as one line, AND binding
// tighter: `a AND b OR c` is [[a, b], [c]]. Each group and condition has
// an id, the key of its block or row, so an edit keeps what is typed in. Between two rules of a group the
// operator is AND, between two groups OR, so setting one splits or merges
// the groups. An expression of nested groups (a Lua rule script, an older
// rule) is expanded into a flat one when loaded.

export type Join = 'and' | 'or';

/** Most conditions of an expression: scripting.MaxConditions */
export const MAX_CONDITIONS = 256;

const newId = () => Math.random().toString(36).substring(2, 9);

/** An empty condition, with the fields of patch */
export const createCondition = (patch: Partial<Omit<Condition, 'id'>> = {}): Condition => ({
  id: newId(),
  source: FieldSource.FIELD_SOURCE_METHOD,
  key: '',
  operator: Operator.OPERATOR_EQ,
  value: '',
  ...patch,
});

/** A group of the conditions, an AND of them: id names its block in the editor */
const groupOf = (conditions: Condition[], id = newId()): AndCondition => ({
  id,
  rules: conditions.map((condition) => ({ condition })),
  orCondition: [],
  andCondition: [],
});

const conditionsOf = (g: AndCondition): Condition[] =>
  g.rules.map((r) => r.condition ?? createCondition());

/** An expression of a single empty condition */
export const createEmptyExpression = (): Expression => ({
  orCondition: [groupOf([createCondition()])],
});

export function countConditions(expr: Expression | undefined): number {
  const count = (g: AndCondition): number =>
    g.rules.length +
    g.orCondition.reduce((n, a) => n + count(a), 0) +
    g.andCondition.reduce((n, a) => n + count(a), 0);
  return (expr?.orCondition ?? []).reduce((n, g) => n + count(g), 0);
}

// ─── Flatten ─────────────────────────────────────────────────────────────────

/**
 * The expression flat, its nested groups expanded: `a AND (b OR c)` is
 * `a AND b OR a AND c`. Every condition has an id of its own, empty groups
 * are dropped. Undefined when it makes more than MAX_CONDITIONS conditions.
 */
export function flatten(expr: Expression | undefined): Expression | undefined {
  // the first run of a group keeps its id
  const runs: { id: string; conditions: Condition[] }[] = [];
  for (const g of expr?.orCondition ?? []) {
    const expanded = runsOf(g);
    if (!expanded) return undefined;
    expanded.forEach((conditions, i) => runs.push({ id: i === 0 ? g.id : '', conditions }));
  }
  if (size(runs.map((r) => r.conditions)) > MAX_CONDITIONS) return undefined;

  // an expanded condition is in several groups: each copy gets its own id
  const seen = new Set<string>();
  const unique = (id: string) => {
    const next = id && !seen.has(id) ? id : newId();
    seen.add(next);
    return next;
  };
  return {
    orCondition: runs
      .filter((run) => run.conditions.length > 0)
      .map((run) =>
        groupOf(
          run.conditions.map((c) => ({ ...c, id: unique(c.id) })),
          unique(run.id),
        ),
      ),
  };
}

/**
 * The AND runs a group matches as: its rules, AND one of its orCondition,
 * AND each of its andCondition. Undefined past MAX_CONDITIONS.
 */
function runsOf(g: AndCondition): Condition[][] | undefined {
  let runs = [conditionsOf(g)];
  const factors = g.andCondition.map((a) => [a]);
  if (g.orCondition.length > 0) factors.unshift(g.orCondition);

  for (const alternatives of factors) {
    const options: Condition[][] = [];
    for (const alt of alternatives) {
      const expanded = runsOf(alt);
      if (!expanded) return undefined;
      options.push(...expanded);
    }
    runs = runs.flatMap((run) => options.map((o) => [...run, ...o]));
    if (size(runs) > MAX_CONDITIONS) return undefined;
  }
  return runs;
}

const size = (runs: Condition[][]) => runs.reduce((n, run) => n + run.length, 0);

// ─── Edits of a flat expression ──────────────────────────────────────────────

type Group = { id: string; conditions: Condition[] };

/** The groups as lists of conditions, and back */
const groupsOf = (expr: Expression): Group[] =>
  expr.orCondition.map((g) => ({ id: g.id, conditions: conditionsOf(g) }));
const exprOf = (groups: Group[]): Expression => ({
  orCondition: groups
    .filter((g) => g.conditions.length > 0)
    .map((g) => groupOf(g.conditions, g.id)),
});

/** Appends an empty condition to the group g (AND), or as a new group past the last (OR) */
export function addCondition(expr: Expression, g: number): Expression {
  const groups = groupsOf(expr);
  const group = groups[g];
  if (group) group.conditions.push(createCondition());
  else groups.push({ id: newId(), conditions: [createCondition()] });
  return exprOf(groups);
}

/**
 * Sets the operator before the condition r of the group g: OR splits the
 * group there (the first part keeps its id), AND before the first condition
 * of a group merges it into the one before.
 */
export function setJoin(expr: Expression, g: number, r: number, join: Join): Expression {
  const groups = groupsOf(expr);
  const group = groups[g];
  const before = groups[g - 1];
  if (group && r > 0 && join === 'or') {
    groups.splice(
      g,
      1,
      { id: group.id, conditions: group.conditions.slice(0, r) },
      { id: newId(), conditions: group.conditions.slice(r) },
    );
  } else if (group && before && r === 0 && join === 'and') {
    groups.splice(g - 1, 2, {
      id: before.id,
      conditions: [...before.conditions, ...group.conditions],
    });
  } else return expr;
  return exprOf(groups);
}

/** The expression without the condition; a group left empty is dropped */
export const removeCondition = (expr: Expression, id: string): Expression =>
  exprOf(
    groupsOf(expr).map((g) => ({ ...g, conditions: g.conditions.filter((c) => c.id !== id) })),
  );

export const updateCondition = (
  expr: Expression,
  id: string,
  updater: (condition: Condition) => Condition,
): Expression =>
  exprOf(
    groupsOf(expr).map((g) => ({
      ...g,
      conditions: g.conditions.map((c) => (c.id === id ? updater(c) : c)),
    })),
  );

/**
 * The expression with the condition moved to the index of the group toGroup
 * (an id): to another index of its own group, or into another one. A group
 * it leaves empty is dropped.
 */
export function moveCondition(
  expr: Expression,
  id: string,
  toGroup: string,
  toIndex: number,
): Expression {
  const groups = groupsOf(expr);
  const target = groups.find((g) => g.id === toGroup);
  const condition = target && take(groups, id);
  if (!target || !condition) return expr;

  target.conditions.splice(toIndex, 0, condition);
  return exprOf(groups);
}

/** The expression with the condition moved out of its group, into a new one last */
export function moveToNewGroup(expr: Expression, id: string): Expression {
  const groups = groupsOf(expr);
  const condition = take(groups, id);
  if (!condition) return expr;

  groups.push({ id: newId(), conditions: [condition] });
  return exprOf(groups);
}

/** Removes the condition from the groups it is in, and returns it */
function take(groups: Group[], id: string): Condition | undefined {
  for (const g of groups) {
    const i = g.conditions.findIndex((c) => c.id === id);
    if (i >= 0) return g.conditions.splice(i, 1)[0];
  }
  return undefined;
}

/** The expression with the group (an id) moved to the index */
export function moveGroup(expr: Expression, id: string, toIndex: number): Expression {
  const groups = [...expr.orCondition];
  const from = groups.findIndex((g) => g.id === id);
  if (from < 0 || from === toIndex) return expr;
  const [group] = groups.splice(from, 1);
  groups.splice(toIndex, 0, group!);
  return { orCondition: groups };
}
