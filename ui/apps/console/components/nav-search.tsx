'use client';

import { Search } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { Fragment, useMemo, useState } from 'react';

import { Button } from '@sentinez/ui/components/button';
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@sentinez/ui/components/command';
import { cn } from '@sentinez/ui/lib/utils';

import { useDomain } from '@/hooks/use-domain';
import { NavItem, NavTitleKey } from '@/lib/default';
import { DOMAIN_NAV, ROOT_NAV, navPages, withDomain } from '@/lib/nav';

/** A page of the search, its titles translated: those of its trail first */
type SearchPage = { url: string; icon: NavItem['icon']; title: string; trail: string[] };

/**
 * The search of the top bar: a box opening the command menu in a dialog,
 * which finds the pages of the console navigation (lib/default) by their
 * title or the titles above them: the tenant-level ones and, on a domain
 * page, those of the domain; a group with its icons per tab, a separator
 * between two. Picking one goes to it.
 */
export function NavSearch({ className }: { className?: string }) {
  const t = useTranslations('Nav');
  const ts = useTranslations('Sidebar');
  const tc = useTranslations('Common');
  const router = useRouter();
  const domain = useDomain();
  const [open, setOpen] = useState(false);

  // the pages by their tab (the first title of their trail), their titles
  // translated once
  const groups = useMemo(() => {
    const pages = [
      ...navPages(ROOT_NAV),
      ...(domain ? navPages(withDomain(DOMAIN_NAV, domain)) : []),
    ];
    const byTab = new Map<NavTitleKey, SearchPage[]>();
    for (const { item, trail } of pages) {
      const tab = trail[0] ?? item.titleKey;
      const page = {
        url: item.url,
        icon: item.icon,
        title: t(item.titleKey),
        trail: trail.map((k) => t(k)),
      };
      const list = byTab.get(tab);
      if (list) list.push(page);
      else byTab.set(tab, [page]);
    }
    return [...byTab].map(([tab, list]) => ({ heading: t(tab), pages: list }));
  }, [domain, t]);

  const go = (url: string) => {
    setOpen(false);
    router.push(url);
  };

  return (
    <>
      <Button
        type="button"
        variant="outline"
        onClick={() => setOpen(true)}
        className={cn(
          'text-muted-foreground h-8 justify-start gap-2 px-2 text-xs font-normal shadow-none',
          className,
        )}
      >
        <Search className="size-3.5 shrink-0" aria-hidden />
        <span className="truncate">{ts('search')}</span>
      </Button>

      <CommandDialog open={open} onOpenChange={setOpen} title={ts('search')}>
        <CommandInput placeholder={ts('search')} />
        {/* built only while open: closed, the dialog renders nothing */}
        {open && (
          <CommandList>
            <CommandEmpty>{tc('noResults')}</CommandEmpty>
            {groups.map(({ heading, pages }, i) => (
              <Fragment key={heading}>
                {i > 0 && <CommandSeparator />}
                <CommandGroup heading={heading}>
                  {pages.map(({ url, icon: Icon, title, trail }) => (
                    <CommandItem
                      key={url}
                      // unique, and matched on every title of the trail
                      value={url}
                      keywords={[title, ...trail]}
                      onSelect={() => go(url)}
                    >
                      {Icon && <Icon />}
                      <span>{title}</span>
                      {trail.length > 1 && (
                        <span className="text-muted-foreground ml-auto text-xs">
                          {trail.slice(1).join(' › ')}
                        </span>
                      )}
                    </CommandItem>
                  ))}
                </CommandGroup>
              </Fragment>
            ))}
          </CommandList>
        )}
      </CommandDialog>
    </>
  );
}
