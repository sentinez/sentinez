// What the browser islands read of the packages: nothing of node here.

export type Kind = 'ruleset' | 'secrule' | 'limiter' | 'hook' | 'function';

/** What the search reads of a package */
export type PackageSummary = {
  path: string;
  synopsis: string;
  version?: string;
  updated: string;
  keywords: string[];
  kinds: { kind: Kind; count: number }[];
  /** its rules: the anchor, the name, and the condition, doc and actions to search */
  rules: { id: string; name: string; text: string }[];
  files: string[];
};

/** The kinds of rules, in the order of the package page */
export const KINDS: Record<Kind, { label: string; title: string }> = {
  ruleset: { label: 'Ruleset', title: 'Rulesets' },
  secrule: { label: 'SecRule', title: 'SecRules' },
  limiter: { label: 'Rate limit', title: 'Rate limits' },
  hook: { label: 'Hook', title: 'Hooks' },
  function: { label: 'Function', title: 'Functions' },
};

const DATE = new Intl.DateTimeFormat('en', { dateStyle: 'medium' });
export const formatDate = (d: Date | string) => DATE.format(new Date(d));
