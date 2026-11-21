import { readFile, readdir } from 'node:fs/promises';
import path from 'node:path';

export const LANGS = ['en', 'vi'] as const;
export type Lang = (typeof LANGS)[number];

export const REPO_URL = 'https://github.com/sentinez/sentinez/blob/main';

// The docs live at the repo root: <repo>/docs (en) and <repo>/docs/vi.
const DOCS_ROOT = path.resolve(process.cwd(), '../../../docs');

export function isLang(v: string): v is Lang {
  return (LANGS as readonly string[]).includes(v);
}

export interface DocEntry {
  slug: string; // '' for the README (index)
  title: string;
}

function dir(lang: Lang) {
  return lang === 'vi' ? path.join(DOCS_ROOT, 'vi') : DOCS_ROOT;
}

function fileOf(lang: Lang, slug: string) {
  return path.join(dir(lang), `${slug || 'README'}.md`);
}

function titleOf(src: string, fallback: string) {
  return /^#\s+(.+)$/m.exec(src)?.[1].trim() ?? fallback;
}

/** All pages of a language, README first, then 01-… in order. */
export async function listDocs(lang: Lang): Promise<DocEntry[]> {
  const names = (await readdir(dir(lang)))
    .filter((n) => n.endsWith('.md'))
    .map((n) => n.slice(0, -3))
    .sort();
  const entries: DocEntry[] = [];
  for (const name of names) {
    const src = await readFile(fileOf(lang, name), 'utf8');
    entries.push({ slug: name === 'README' ? '' : name, title: titleOf(src, name) });
  }
  return entries.sort((a, b) => (a.slug === '' ? -1 : b.slug === '' ? 1 : 0));
}

export async function readDoc(lang: Lang, slug: string): Promise<string | null> {
  if (!/^[\w-]*$/.test(slug)) return null;
  try {
    return await readFile(fileOf(lang, slug), 'utf8');
  } catch {
    return null;
  }
}

/**
 * Rewrites a link of a markdown page: sibling pages become site routes,
 * repo paths (../internal/...) become GitHub links.
 */
export function resolveHref(href: string, lang: Lang): string {
  if (/^([a-z]+:|#|\/\/)/i.test(href)) return href;
  const [target, hash = ''] = href.split('#');
  const suffix = hash ? `#${hash}` : '';
  if (target === '') return suffix;

  // Where the page sits in the repo: docs/ or docs/vi/.
  const base = lang === 'vi' ? 'docs/vi' : 'docs';
  const parts = base.split('/');
  for (const seg of target.split('/')) {
    if (seg === '..') parts.pop();
    else if (seg !== '.') parts.push(seg);
  }
  const repoPath = parts.join('/');

  const m = /^docs\/(vi\/)?([\w-]+)\.md$/.exec(repoPath);
  if (m) {
    const l = m[1] ? 'vi' : 'en';
    return `/engineering/${l}${m[2] === 'README' ? '' : `/${m[2]}`}${suffix}`;
  }
  return `${REPO_URL}/${repoPath}${suffix}`;
}
