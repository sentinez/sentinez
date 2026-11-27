import {
  AndCondition,
  Condition,
  Expression,
  FieldSource,
  Operator,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';

// The editor works on a tree of groups, as scripting.Node does on the edge:
// a group matches when all (or any) of its children do. It converts to and
// from the Expression of the API the way core/szrule/rulescript does:
//
//   AndCondition = rules AND (one of orCondition) AND (every one of andCondition)

export type GroupOp = 'all' | 'any';

export type CondNode = { kind: 'cond'; id: string; condition: Condition };
export type GroupNode = {
  kind: 'group';
  id: string;
  op: GroupOp;
  children: RuleNode[];
};
export type RuleNode = CondNode | GroupNode;

/** Most conditions of an expression: scripting.MaxConditions */
export const MAX_CONDITIONS = 256;
/**
 * Deepest group a user can add. The edge accepts 8 nested levels of and / or
 * groups; a group of the tree costs at most one, plus the root or.
 */
export const MAX_GROUP_DEPTH = 2;

const newId = () => Math.random().toString(36).substring(2, 9);

const createEmptyCondition = (): Condition => ({
  id: newId(),
  source: FieldSource.FIELD_SOURCE_METHOD,
  key: '',
  operator: Operator.OPERATOR_EQ,
  value: '',
});

/** A condition node: an empty condition, with the fields of patch */
export const createCondNode = (patch: Partial<Omit<Condition, 'id'>> = {}): CondNode => {
  const condition = { ...createEmptyCondition(), ...patch };
  return { kind: 'cond', id: condition.id, condition };
};

export const createGroupNode = (op: GroupOp, children: RuleNode[] = []): GroupNode => ({
  kind: 'group',
  id: newId(),
  op,
  children,
});

// ─── Expression -> tree ──────────────────────────────────────────────────────

/**
 * The tree of an expression. One or branch is the root itself; several make
 * an "any" root. A group of a single child is the child.
 */
export function treeOf(expr: Expression | undefined): GroupNode {
  const ors = expr?.orCondition ?? [];
  if (ors.length === 0) return createGroupNode('all');
  if (ors.length === 1) {
    const n = fromAnd(ors[0]!);
    return n.kind === 'group' ? n : createGroupNode('all', [n]);
  }
  return createGroupNode('any', ors.map(fromAnd));
}

function fromAnd(and: AndCondition): RuleNode {
  const children: RuleNode[] = [];
  for (const rule of and.rules) {
    const condition = rule.condition ?? createEmptyCondition();
    const id = condition.id || newId();
    children.push({ kind: 'cond', id, condition: { ...condition, id } });
  }
  if (and.orCondition.length > 0) {
    children.push(collapse(createGroupNode('any', and.orCondition.map(fromAnd))));
  }
  for (const sub of and.andCondition) children.push(fromAnd(sub));

  return collapse(createGroupNode('all', children));
}

/** A group of one child is that child, so a lone condition is not boxed */
const collapse = (g: GroupNode): RuleNode => (g.children.length === 1 ? g.children[0]! : g);

// ─── tree -> Expression ──────────────────────────────────────────────────────

/** The expression of the tree. Empty groups are dropped, as the edge refuses them. */
export function exprOf(root: RuleNode): Expression {
  const n = prune(root);
  if (!n) return { orCondition: [] };
  if (n.kind === 'group' && n.op === 'any') {
    return { orCondition: n.children.map(toAnd) };
  }
  return { orCondition: [toAnd(n)] };
}

function prune(n: RuleNode): RuleNode | undefined {
  if (n.kind === 'cond') return n;
  const children = n.children.map(prune).filter((c): c is RuleNode => !!c);
  return children.length > 0 ? { ...n, children } : undefined;
}

/** Same shape as andOf of core/szrule/rulescript */
function toAnd(n: RuleNode): AndCondition {
  if (n.kind === 'cond')
    return { rules: [{ condition: n.condition }], orCondition: [], andCondition: [] };
  if (n.op === 'any') {
    return { rules: [], orCondition: n.children.map(toAnd), andCondition: [] };
  }

  const and: AndCondition = { rules: [], orCondition: [], andCondition: [] };
  for (const c of flatAll(n.children)) {
    if (c.kind === 'cond') and.rules.push({ condition: c.condition });
    else if (c.op === 'any' && and.orCondition.length === 0) {
      and.orCondition = c.children.map(toAnd);
    } else and.andCondition.push(toAnd(c));
  }
  return and;
}

/** The children of an "all", the ones of a nested "all" spliced in */
const flatAll = (children: RuleNode[]): RuleNode[] =>
  children.flatMap((c) => (c.kind === 'group' && c.op === 'all' ? flatAll(c.children) : [c]));

// ─── Edits ───────────────────────────────────────────────────────────────────

export function mapNode<T extends RuleNode>(n: T, id: string, fn: (n: RuleNode) => RuleNode): T {
  if (n.id === id) return fn(n) as T;
  if (n.kind === 'cond') return n;
  const children = n.children.map((c) => mapNode(c, id, fn));
  // a subtree keeps its identity when no node of it changed
  return children.some((c, i) => c !== n.children[i]) ? { ...n, children } : n;
}

export function removeNode(n: GroupNode, id: string): GroupNode {
  const children = n.children
    .filter((c) => c.id !== id)
    .map((c) => (c.kind === 'group' ? removeNode(c, id) : c));
  const same =
    children.length === n.children.length && children.every((c, i) => c === n.children[i]);
  return same ? n : { ...n, children };
}

export function countConditions(n: RuleNode): number {
  return n.kind === 'cond' ? 1 : n.children.reduce((sum, c) => sum + countConditions(c), 0);
}

export const isEmptyGroup = (n: GroupNode) => countConditions(n) === 0;
