// The packages of pkg/: a folder each, of YAML and Lua security rules. Read
// at build time (and again by `astro dev` when a file of pkg/ changes).
import { execFileSync } from 'node:child_process';
import { statSync } from 'node:fs';
import path from 'node:path';

import { marked } from 'marked';

import { fileDoc, luaRules, yamlRules, type Lang, type Rule } from './rule';
import { KINDS, type Kind, type PackageSummary } from './summary';

export type SourceFile = { path: string; name: string; content: string; lines: number; lang: Lang };

export type Package = {
  /** the folder under pkg/, e.g. `auth/login` */
  id: string;
  synopsis: string;
  /** the doc of the package, as HTML */
  docHtml: string;
  version?: string;
  license: string;
  keywords: string[];
  updated: Date;
  files: SourceFile[];
  rules: Rule[];
};

const PKG_DIR = path.resolve('pkg');

// every rule file under pkg/
const RAW = import.meta.glob<string>('/pkg/**/*.{lua,yaml,yml}', { query: '?raw', import: 'default', eager: true });

const dirOf = (p: string) => p.slice(0, p.lastIndexOf('/'));

/**
 * The date of the last commit of each folder under pkg/ (its sub folders
 * included), from one `git log`; empty out of a git checkout.
 */
function commitDates(): Map<string, Date> {
  const dates = new Map<string, Date>();
  let out = '';
  try {
    out = execFileSync('git', ['log', '--format=%x00%cI', '--name-only', '--relative', '--', '.'], {
      cwd: PKG_DIR,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
  } catch {
    return dates;
  }
  // newest commit first: the first date seen for a folder is its last one
  for (const commit of out.split('\0').slice(1)) {
    const [date, ...files] = commit.trim().split('\n');
    for (const file of files) {
      for (let dir = dirOf(file); dir; dir = dirOf(dir)) if (!dates.has(dir)) dates.set(dir, new Date(date));
    }
  }
  return dates;
}

/** The first sentence of the doc */
const synopsisOf = (text: string) =>
  text.split(/\n\s*\n/)[0].replace(/\s+/g, ' ').match(/^.*?[.!?](\s|$)/)?.[0].trim() ?? text.split('\n')[0];

function load(): Package[] {
  // sorted by path, so the folders come out in order
  const files: SourceFile[] = Object.entries(RAW)
    .map(([file, content]) => {
      const p = file.replace(/^\/pkg\//, '');
      const name = p.slice(p.lastIndexOf('/') + 1);
      return { path: p, name, content, lines: content.split('\n').length, lang: name.endsWith('.lua') ? 'lua' : 'yaml' } as const;
    })
    // a file right in pkg/ belongs to no package
    .filter((f) => dirOf(f.path))
    .sort((a, b) => a.path.localeCompare(b.path));

  const byDir = new Map<string, SourceFile[]>();
  for (const f of files) {
    const dir = dirOf(f.path);
    const list = byDir.get(dir);
    if (list) list.push(f);
    else byDir.set(dir, [f]);
  }
  const dates = commitDates();

  return [...byDir].map(([id, files]) => {
    // the doc of the package: the file docs, in file order
    const docs = files.map((f) => fileDoc(f.content, f.lang));
    const text = docs.map((d) => d.text).filter(Boolean).join('\n\n');
    const tags = Object.assign({}, ...docs.map((d) => d.tags)) as Record<string, string>;
    const rules = files.flatMap((f, i) =>
      f.lang === 'lua' ? luaRules(f.content, f.path) : yamlRules(f.content, f.path, docs[i].lines),
    );
    return {
      id,
      synopsis: synopsisOf(text) || `${rules.length} security rules.`,
      docHtml: text ? (marked.parse(text, { async: false }) as string) : '',
      version: tags.version,
      license: tags.license ?? 'Apache-2.0',
      keywords: tags.keywords?.split(/[,\s]+/).filter(Boolean) ?? [],
      // uncommitted: the newest of its files on disk
      updated: dates.get(id) ?? new Date(Math.max(...files.map((f) => statSync(path.join(PKG_DIR, f.path)).mtimeMs))),
      files,
      rules,
    };
  });
}

let cache: Package[] | undefined;

/** The packages, by path */
export function getPackages(): Package[] {
  cache ??= load();
  return cache;
}

/** The packages one level under `pkg` */
export function subPackages(pkg: Package): Package[] {
  const dir = `${pkg.id}/`;
  return getPackages().filter((p) => p.id.startsWith(dir) && !p.id.slice(dir.length).includes('/'));
}

/** The rules by kind, in the order of KINDS, the kinds without rules left out */
export function rulesByKind(rules: Rule[]): { kind: Kind; rules: Rule[] }[] {
  const groups = new Map<Kind, Rule[]>(Object.keys(KINDS).map((k) => [k as Kind, []]));
  for (const r of rules) groups.get(r.kind)!.push(r);
  return [...groups].filter(([, list]) => list.length).map(([kind, rules]) => ({ kind, rules }));
}

export function summaries(all: Package[]): PackageSummary[] {
  return all.map((p) => ({
    path: p.id,
    synopsis: p.synopsis,
    version: p.version,
    updated: p.updated.toISOString(),
    keywords: p.keywords,
    kinds: rulesByKind(p.rules).map(({ kind, rules }) => ({ kind, count: rules.length })),
    rules: p.rules.map((r) => ({ id: r.id, name: r.name, text: [r.when, r.doc, ...r.actions].filter(Boolean).join(' ') })),
    files: p.files.map((f) => f.name),
  }));
}
