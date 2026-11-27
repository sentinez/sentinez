import { dashboard, NavItem, NavTitleKey } from '@/lib/default';

/** Drops the items marked `hidden`, at every level. */
function visible(items: NavItem[]): NavItem[] {
  return items
    .filter((item) => !item.hidden)
    .map((item) => (item.items ? { ...item, items: visible(item.items) } : item));
}

/** The tabs of the tenant-level pages (`/d/console/(root)`, `/d/setting`). */
export const ROOT_NAV = visible(dashboard.rootNavMain);

/** The tabs of the domain pages, with `/d/console/<path>` urls. */
export const DOMAIN_NAV = visible(dashboard.domainNavMain);

/** Rewrites `/d/console/<path>` urls to `/d/console/<domain>/<path>`. */
export function withDomain(items: NavItem[], domain: string): NavItem[] {
  return items.map((item) => ({
    ...item,
    url: domainUrl(item.url, domain),
    items: item.items ? withDomain(item.items, domain) : undefined,
  }));
}

function domainUrl(url: string, domain: string) {
  if (!url.startsWith('/d/console/')) return url;
  const parts = url.split('/');
  if (parts[3] === domain) return url;
  return `/d/console/${domain}/${parts.slice(3).join('/')}`;
}

/** A page of the navigation, with the titles of the items above it. */
export type NavPage = { item: NavItem; trail: NavTitleKey[] };

/** The pages of the items: each one without sub items, at every level. */
export function navPages(items: NavItem[], trail: NavTitleKey[] = []): NavPage[] {
  return items.flatMap((item) =>
    item.items?.length ? navPages(item.items, [...trail, item.titleKey]) : [{ item, trail }],
  );
}
