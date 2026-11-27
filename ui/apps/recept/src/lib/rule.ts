// Reads the security rules of a package file: the YAML of the `security`
// block of proxy.yaml (secrules, limiters, ruleset) or a Lua script
// (match, key, ---@stage hooks). The doc comments become the documentation.
import { isMap, isSeq, parseDocument, type Node, type YAMLMap, type YAMLSeq } from 'yaml';

import { KINDS, type Kind } from './summary';

export type Lang = 'yaml' | 'lua';

export type Rule = {
  kind: Kind;
  /** the anchor on the package page */
  id: string;
  name: string;
  doc?: string;
  /** active / disabled */
  status?: string;
  /** what a match does: block, challenge… */
  actions: string[];
  /** the condition, readable */
  when?: string;
  /** a limiter: 100 requests / 5s */
  limit?: string;
  /** a hook: the phase/stage lines */
  stages: string[];
  /** a ruleset: its settings */
  details: { label: string; value: string }[];
  code: string;
  lang: Lang;
  file: string;
};

/** The doc of a file: its first comment block, followed by a blank line */
export type FileDoc = { text: string; tags: Record<string, string>; lines: number };

const COMMENT: Record<Lang, RegExp> = { lua: /^\s*--+ ?/, yaml: /^\s*# ?/ };

const slug = (s: string) =>
  s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '') || 'rule';

/**
 * The first comment block of the file when a blank line follows it (else
 * it documents what comes next), its `@tag value` lines apart.
 */
export function fileDoc(source: string, lang: Lang): FileDoc {
  const marker = COMMENT[lang];
  const lines = source.split('\n');
  const block: string[] = [];
  let i = 0;
  for (; i < lines.length && marker.test(lines[i]); i++) block.push(lines[i].replace(marker, ''));
  if (!block.length || (i < lines.length && lines[i].trim() !== '')) return { text: '', tags: {}, lines: 0 };
  const tags: Record<string, string> = {};
  const text = block
    .filter((l) => {
      const m = l.match(/^\s*@(\w+)\s+(.*)$/);
      if (m) tags[m[1]] = m[2].trim();
      return !m;
    })
    .join('\n')
    .trim();
  return { text, tags, lines: block.length };
}

// ---------------------------------------------------------------- YAML

const OPERATORS: Record<string, string> = {
  OPERATOR_EQ: '==',
  OPERATOR_NE: '!=',
  OPERATOR_CONTAINS: 'contains',
  OPERATOR_PREFIX: 'starts with',
  OPERATOR_SUFFIX: 'ends with',
  OPERATOR_GT: '>',
  OPERATOR_GTE: '>=',
  OPERATOR_LT: '<',
  OPERATOR_LTE: '<=',
  OPERATOR_MATCHES: 'matches',
  OPERATOR_IN: 'in',
  OPERATOR_NOT_IN: 'not in',
};

type Cond = { source?: string; key?: string; operator?: string; value?: unknown };
type Group = { rules?: { condition?: Cond }[]; orCondition?: Group[]; andCondition?: Group[] };

function condText(c: Cond): string {
  // FIELD_SOURCE_HEADER + key → header("key"), FIELD_SOURCE_PATH → path
  const field = (c.source ?? '?').replace('FIELD_SOURCE_', '').toLowerCase() + (c.key && /HEADER|QUERY/.test(c.source ?? '') ? `("${c.key}")` : '');
  const op = OPERATORS[c.operator ?? ''] ?? c.operator ?? '?';
  return `${field} ${op} ${JSON.stringify(c.value ?? '')}`;
}

// A condition as text, and how it is joined: an `or` inside an `and` takes
// parentheses, an `and` inside an `or` too, for reading
type Text = { t: string; op: 'and' | 'or' | 'leaf' };

function join(op: 'and' | 'or', parts: Text[]): Text {
  if (parts.length === 1) return parts[0];
  const t = parts.map((p) => (p.op !== 'leaf' && p.op !== op ? `(${p.t})` : p.t)).join(` ${op} `);
  return { t, op };
}

/** A group: its conditions, its or groups and its and groups, all matching */
function groupText(g: Group): Text {
  const parts: Text[] = (g.rules ?? [])
    .filter((r) => r.condition)
    .map((r) => ({ t: condText(r.condition!), op: 'leaf' }));
  if (g.orCondition?.length) parts.push(orText(g.orCondition));
  for (const a of g.andCondition ?? []) parts.push(groupText(a));
  return join('and', parts);
}

const orText = (groups: Group[]): Text => join('or', groups.map(groupText));

// the words of an enum value that read as acronyms
const ACRONYMS: Record<string, string> = { owasp: 'OWASP', lfi: 'LFI', rfi: 'RFI', rce: 'RCE', xss: 'XSS', sqli: 'SQLi', php: 'PHP', iis: 'IIS', sql: 'SQL' };

/**
 * An enum value as words: STATUS_ACTIVE → active; with `caps`, for a label,
 * CATEGORY_SQLI → SQLi, ENGINE_MODE_DETECTION_ONLY → Detection only.
 */
function enumText(v: unknown, prefix: string, caps = false): string | undefined {
  if (typeof v !== 'string') return undefined;
  const words = v.replace(prefix, '').toLowerCase().split('_');
  if (!caps) return words.join(' ');
  const text = words.map((w) => ACRONYMS[w] ?? w).join(' ');
  return text[0].toUpperCase() + text.slice(1);
}

/** The lines from the start of the line of `from` to `to`, dedented */
function sliceLines(source: string, from: number, to: number): string {
  const start = source.lastIndexOf('\n', from - 1) + 1;
  const lines = source.slice(start, to).trimEnd().split('\n');
  const indent = lines[0].match(/^ */)![0].length;
  return lines.map((l) => l.slice(Math.min(indent, l.match(/^ */)![0].length))).join('\n');
}

/** The comment above a sequence item or a map key */
function commentOf(node: Node | null | undefined, fallback?: string | null): string | undefined {
  const c = (node as { commentBefore?: string | null } | null)?.commentBefore ?? fallback;
  return c?.split('\n').map((l) => l.replace(/^ ?/, '')).join('\n').trim() || undefined;
}

type Ingress = { name?: string; status?: string; expr?: { orCondition?: Group[] }; script?: string };

function ingressRule(kind: 'secrule' | 'limiter', item: Record<string, unknown>, doc: string | undefined, code: string, file: string, index: number): Rule {
  const ingress = (item.ingress ?? {}) as Ingress;
  const name = ingress.name ?? `${KINDS[kind].label} ${index + 1}`;
  const action = (item.action as { type?: string } | undefined)?.type;
  const script = ingress.script ? luaRules(ingress.script, file)[0] : undefined;
  const limit =
    kind === 'limiter' && item.maxRequests
      ? `${item.maxRequests} requests / ${item.timeWindow ?? '?'}${item.timeout ? `, then blocked ${item.timeout}` : ''}`
      : undefined;
  return {
    kind,
    id: slug(name),
    name,
    doc,
    status: enumText(ingress.status, 'STATUS_'),
    actions: action ? [enumText(action, 'ACTION_TYPE_')!] : (script?.actions ?? (kind === 'limiter' ? ['rate limit'] : [])),
    when: ingress.expr?.orCondition ? orText(ingress.expr.orCondition).t : ingress.script ? 'the Lua script below' : undefined,
    limit,
    stages: [],
    details: [],
    code,
    lang: 'yaml',
    file,
  };
}

function rulesetRule(value: Record<string, unknown>, doc: string | undefined, code: string, file: string): Rule {
  const crs = (value.crs ?? {}) as Record<string, unknown>;
  const threshold = (crs.threshold ?? {}) as Record<string, unknown>;
  const count = (v: unknown) => (Array.isArray(v) ? String(v.length) : undefined);
  const details = [
    { label: 'Provider', value: enumText(value.provider, 'PROVIDER_', true) },
    { label: 'Mode', value: enumText(crs.mode, 'ENGINE_MODE_', true) },
    { label: 'CRS version', value: enumText(crs.version, 'CRS_VERSION_')?.replace(/ /g, '.') },
    { label: 'Paranoia level', value: crs.paranoiaLevel?.toString() },
    { label: 'Detection paranoia', value: crs.detectionParanoiaLevel?.toString() },
    { label: 'Inbound threshold', value: threshold.inbound?.toString() },
    { label: 'Outbound threshold', value: threshold.outbound?.toString() },
    { label: 'Categories', value: Array.isArray(crs.categories) ? (crs.categories as string[]).map((c) => enumText(c, 'CATEGORY_', true)).join(', ') : 'All' },
    { label: 'Exclusions', value: count(crs.exclusions) },
    { label: 'Overrides', value: count(crs.overrides) },
  ].filter((d): d is { label: string; value: string } => !!d.value);
  const mode = enumText(crs.mode, 'ENGINE_MODE_');
  return {
    kind: 'ruleset',
    id: 'ruleset',
    name: 'OWASP CRS ruleset',
    doc,
    status: enumText(value.status, 'STATUS_'),
    actions: mode === 'on' ? ['block'] : mode === 'detection only' ? ['log'] : [],
    stages: [],
    details,
    code,
    lang: 'yaml',
    file,
  };
}

/**
 * The rules of a YAML file: `security:` (or `setting.security`) or its
 * content at the root. The first `docLines` lines, the file doc, are left
 * out, so no rule takes them for its own comment.
 */
export function yamlRules(source: string, file: string, docLines: number): Rule[] {
  const src = source.split('\n').map((l, i) => (i < docLines ? '' : l)).join('\n');
  const doc = parseDocument(src);
  if (doc.errors.length) throw new Error(`${file}: ${doc.errors[0].message}`);
  let root: unknown = doc.contents;
  for (const key of ['setting', 'security']) {
    const next = isMap(root) ? (root as YAMLMap).get(key, true) : undefined;
    if (isMap(next)) root = next;
  }
  if (!isMap(root)) return [];

  const rules: Rule[] = [];
  for (const pair of (root as YAMLMap).items) {
    const keyNode = pair.key as Node;
    const key = String((keyNode as { value?: unknown }).value);
    if (key === 'ruleset' && isMap(pair.value)) {
      const value = pair.value as YAMLMap;
      const code = sliceLines(src, keyNode.range![0], value.range![1]);
      rules.push(rulesetRule(value.toJSON(), commentOf(keyNode), code, file));
    } else if ((key === 'secrules' || key === 'limiters') && isSeq(pair.value)) {
      const seq = pair.value as YAMLSeq;
      seq.items.forEach((item, i) => {
        if (!isMap(item)) return;
        const node = item as YAMLMap;
        const doc = commentOf(node, i === 0 ? seq.commentBefore : undefined);
        const code = sliceLines(src, node.range![0], node.range![1]);
        rules.push(ingressRule(key === 'secrules' ? 'secrule' : 'limiter', node.toJSON(), doc, code, file, i));
      });
    }
  }
  return rules;
}

// ----------------------------------------------------------------- Lua

const LUA_COMMENT = COMMENT.lua;

/** The global functions of a Lua script, with the comment lines right above them */
export function luaRules(source: string, file: string): Rule[] {
  const lines = source.split('\n');
  const rules: Rule[] = [];
  for (let i = 0; i < lines.length; i++) {
    const m = lines[i].match(/^function\s+([A-Za-z_][\w.:]*)\s*\(([^)]*)\)/);
    if (!m) continue;
    let start = i;
    while (start > 0 && LUA_COMMENT.test(lines[start - 1])) start--;
    let end = i;
    while (end < lines.length - 1 && !/^end\b/.test(lines[end])) end++;

    const comments = lines.slice(start, i);
    const stages = comments.map((l) => l.match(/---@stage\s+(\S+)/)?.[1]).filter((s): s is string => !!s);
    const doc = comments
      .filter((l) => !/^\s*---@/.test(l))
      .map((l) => l.replace(LUA_COMMENT, ''))
      .join('\n')
      .trim();
    const body = lines.slice(i, end + 1).join('\n');
    const name = m[1];
    const kind: Kind = stages.length ? 'hook' : name === 'match' ? 'secrule' : name === 'key' ? 'limiter' : 'function';
    // the actions a match returns: SZ_BLOCK or {type = SZ_REDIRECT, …}
    const actions = [...new Set([...body.matchAll(/return\s+(?:\{\s*type\s*=\s*)?SZ_([A-Z_]+)/g)].map((a) => a[1].toLowerCase().replace(/_/g, ' ')))];
    if (/ctx:respond\(/.test(body)) actions.push('respond');
    rules.push({
      kind,
      id: slug(`${file.split('/').pop()?.replace(/\.lua$/, '')}-${name}`),
      name: `${name}(${m[2]})`,
      doc: doc || undefined,
      actions,
      when: kind === 'limiter' ? 'key(ctx) returns true: counted under the client IP' : undefined,
      stages,
      details: [],
      code: lines.slice(start, end + 1).join('\n'),
      lang: 'lua',
      file,
    });
    i = end;
  }
  return rules;
}
