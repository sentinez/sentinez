import { Badge } from '@sentinez/ui/components/badge';
import { useEffect, useMemo, useState } from 'react';

import { SearchBox } from '@/components/search-box';
import { formatDate, KINDS, type PackageSummary } from '@/lib/summary';

type Hit = { pkg: PackageSummary; score: number; rules: PackageSummary['rules'] };


type Rule = PackageSummary['rules'][number];

/** A package with its texts lowercased once, for the search */
type Indexed = { pkg: PackageSummary; path: string; synopsis: string; keywords: string[]; rules: [Rule, string][] };

const indexOf = (pkg: PackageSummary): Indexed => ({
  pkg,
  path: pkg.path.toLowerCase(),
  synopsis: pkg.synopsis.toLowerCase(),
  keywords: pkg.keywords.map((k) => k.toLowerCase()),
  rules: pkg.rules.map((r) => [r, `${r.name} ${r.text}`.toLowerCase()]),
});

/** Ranks a package for the words: its path first, then its doc, keywords, rules */
function rank(p: Indexed, words: string[]): Hit | null {
  let score = 0;
  const rules = new Set<Rule>();
  for (const w of words) {
    const matched = p.rules.filter(([, text]) => text.includes(w)).map(([r]) => r);
    let s = 0;
    if (p.path === w || p.path.endsWith(`/${w}`)) s += 50;
    else if (p.path.includes(w)) s += 20;
    if (p.synopsis.includes(w)) s += 8;
    if (p.keywords.some((k) => k.includes(w))) s += 6;
    if (matched.length) s += 4;
    if (!s) return null; // every word must match somewhere
    score += s;
    matched.forEach((m) => rules.add(m));
  }
  return { pkg: p.pkg, score, rules: [...rules].slice(0, 6) };
}

/** The results of /search?q=…, filtered in the browser */
export function PackageSearch({ packages }: { packages: PackageSummary[] }) {
  // read after hydration: the page is static, the query only lives in the url
  const [query, setQuery] = useState<string | null>(null);
  useEffect(() => setQuery(new URLSearchParams(location.search).get('q') ?? ''), []);
  const indexed = useMemo(() => packages.map(indexOf), [packages]);

  const hits = useMemo<Hit[]>(() => {
    const words = (query ?? '').toLowerCase().split(/\s+/).filter(Boolean);
    if (!words.length)
      return packages.map((pkg) => ({ pkg, score: 0, rules: [] }));
    return indexed
      .map((p) => rank(p, words))
      .filter((h): h is Hit => !!h)
      .sort((a, b) => b.score - a.score || a.pkg.path.localeCompare(b.pkg.path));
  }, [packages, indexed, query]);

  if (query === null) return null;

  return (
    <div className="grid gap-6">
      <SearchBox key={query} defaultValue={query} size="lg" />
      <p className="text-sm text-muted-foreground">
        {query ? (
          <>
            {hits.length} {hits.length === 1 ? 'result' : 'results'} for{' '}
            <span className="font-medium text-foreground">“{query}”</span>
          </>
        ) : (
          <>All {hits.length} packages</>
        )}
      </p>
      {hits.length ? (
        <ul className="divide-y border-y">
          {hits.map(({ pkg, rules }) => (
            <li key={pkg.path} className="grid gap-1.5 py-4">
              <a href={`/pkg/${pkg.path}`} className="w-fit font-mono text-base font-medium text-primary hover:underline">
                {pkg.path}
              </a>
              <p className="text-sm">{pkg.synopsis}</p>
              <p className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
                {pkg.kinds.map(({ kind, count }) => (
                  <span key={kind}>
                    {count} {KINDS[kind].label}
                  </span>
                ))}
                {pkg.version && <span className="font-mono">{pkg.version}</span>}
                <span>Updated {formatDate(pkg.updated)}</span>
              </p>
              {rules.length > 0 && (
                <div className="flex flex-wrap gap-1.5 pt-1">
                  {rules.map((r) => (
                    <Badge key={r.id} variant="outline" asChild>
                      <a href={`/pkg/${pkg.path}#${r.id}`}>{r.name}</a>
                    </Badge>
                  ))}
                </div>
              )}
            </li>
          ))}
        </ul>
      ) : (
        <div className="rounded-lg border border-dashed p-10 text-center text-sm text-muted-foreground">
          No package matches “{query}”. Try a shorter word, or{' '}
          <a href="/search" className="text-primary hover:underline">
            list every package
          </a>
          .
        </div>
      )}
    </div>
  );
}
