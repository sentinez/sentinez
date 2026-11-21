'use client';

import { NavUser } from '@/components/nav-user';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarInput,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarRail,
  useSidebar,
} from '@sentinez/ui/components/sidebar';
import Image from 'next/image';
import { useTranslations } from 'next-intl';
import { dashboard, NavItem } from '@/lib/default';
import { usePathname, useRouter } from 'next/navigation';
import { TeamSwitcher } from './team-switcher';
import { cn } from '@sentinez/ui/lib/utils';
import Link from 'next/link';
import PreviewHeader from './preview-header';
import PreviewFooter from './preview-footer';
import { ChevronLeft, Search } from 'lucide-react';
import SidebarLoading from './sidebar-loading';
import { TabLoading } from './tab-loading';
import { ComponentProps, ReactNode, useEffect, useMemo, useState, useTransition } from 'react';
import { useLocalUser } from '@/components/user-menu';
import { isUnder } from '@/lib/session';

type SidebarProps = ComponentProps<typeof Sidebar>;

/** Sidebar of the tenant-level pages (`/d/console/(root)`, `/d/setting`). */
export function RootSidebar(props: SidebarProps) {
  return <ConsoleSidebar navMain={_rootNavMain} {...props} />;
}

/** Sidebar of the domain pages (`/d/console/[domain]/...`). */
export function DomainSidebar(props: SidebarProps) {
  const domain = usePathname().split('/')[3];
  const navMain = useMemo(
    () => (domain ? withDomain(_domainNavMain, domain) : _domainNavMain),
    [domain],
  );

  return <ConsoleSidebar navMain={navMain} {...props} />;
}

/** Drops the items marked `hidden`, at every level. */
function visible(items: NavItem[]): NavItem[] {
  return items
    .filter((item) => !item.hidden)
    .map((item) => (item.items ? { ...item, items: visible(item.items) } : item));
}

const _rootNavMain = visible(dashboard.rootNavMain);
const _domainNavMain = visible(dashboard.domainNavMain);

/** Rewrites `/d/console/<path>` urls to `/d/console/<domain>/<path>`. */
function withDomain(items: NavItem[], domain: string): NavItem[] {
  return items.map((item) => ({
    ...item,
    url: domainUrl(item.url, domain),
    items: item.items ? withDomain(item.items, domain) : undefined,
  }));
}

/**
 * The longest url among `item`, its children and their sub items that
 * `pathname` matches: `sec-rule` must not win over `sec-ruleset`.
 */
function bestUrl(item: NavItem, pathname: string): string | undefined {
  let best = isUnder(pathname, item.url) ? item.url : undefined;
  for (const child of item.items ?? []) {
    const url = bestUrl(child, pathname);
    if (url && url.length > (best?.length ?? 0)) best = url;
  }
  return best;
}

function domainUrl(url: string, domain: string) {
  if (!url.startsWith('/d/console/')) return url;
  const parts = url.split('/');
  if (parts[3] === domain) return url;
  return `/d/console/${domain}/${parts.slice(3).join('/')}`;
}

/**
 * Two-level console sidebar: an icon rail with `navMain`, and a panel with the
 * children of the active item (searchable, one level of collapsible sub items).
 */
export function ConsoleSidebar({ navMain, ...props }: SidebarProps & { navMain: NavItem[] }) {
  const t = useTranslations('Nav');
  const ts = useTranslations('Sidebar');
  const router = useRouter();
  const pathname = usePathname();

  const [, startTransitionNavMain] = useTransition();
  const [, startTransitionChildren] = useTransition();

  const [loading, setLoading] = useState(true);
  const [activeItem, setActiveItem] = useState(navMain[0]);
  const [childItems, setChildItems] = useState(navMain[0]?.items);
  const [navIndex, setNavIndex] = useState(0);
  // The url of the selected tab or sub tab; set on click so the selection
  // moves at once, then synced from the pathname.
  const [activeUrl, setActiveUrl] = useState('');
  const [search, setSearch] = useState('');
  const [mobileTab, setMobileTab] = useState<'main' | 'domain'>('domain');
  const { setOpen, setOpenMobile, isMobile } = useSidebar();
  const user = useLocalUser();

  useEffect(() => {
    const handler = setTimeout(() => {
      const originalItems = navMain[navIndex]?.items || [];
      if (search.trim() === '') {
        setChildItems(originalItems);
      } else {
        const filtered = originalItems.filter((item) =>
          t(item.titleKey).toLowerCase().includes(search.toLowerCase()),
        );
        setChildItems(filtered);
      }
    }, 300); // debounce 300ms

    return () => clearTimeout(handler); // cleanup
  }, [search, activeItem, navIndex, navMain, t]);

  useEffect(() => {
    let best: { item: NavItem; index: number; url: string } | undefined;
    for (const [index, item] of navMain.entries()) {
      const url = bestUrl(item, pathname);
      if (url && url.length > (best?.url.length ?? 0)) best = { item, index, url };
    }

    setActiveUrl(best?.url ?? '');
    if (best) {
      setActiveItem(best.item);
      setNavIndex(best.index);
      setChildItems(best.item.items);
    }

    // Always leave the skeleton state, even when no nav item/child matched.
    setLoading(false);
  }, [pathname, navMain]);

  const handlerSidebarNavMainClick = (item: NavItem, index: number) => {
    router.push(item.url);

    startTransitionNavMain(() => {
      setActiveItem(item);
      setChildItems(item.items);
      setActiveUrl(item.url);
      setNavIndex(index);
      setOpen(true);
      if (isMobile) {
        if (item.items && item.items.length > 0) {
          setMobileTab('domain');
        } else {
          setOpenMobile(false);
        }
      }
    });
  };

  const navigateTo = (url: string) => {
    router.push(url);

    startTransitionChildren(() => {
      setActiveUrl(url);
      if (isMobile) {
        setOpenMobile(false);
      }
    });
  };

  /** A tab with sub tabs opens its first sub tab. */
  const handlerSidebarChildrenClick = (item: NavItem) =>
    navigateTo(item.items?.[0]?.url ?? item.url);

  return (
    <Sidebar
      collapsible="icon"
      className="overflow-hidden *:data-[sidebar=sidebar]:flex-row flex-col md:flex-row"
      {...props}
    >
      {isMobile && (
        <div className="flex items-center border-b p-2 bg-sidebar text-sm font-medium gap-1 shrink-0 w-full">
          <button
            type="button"
            onClick={() => setMobileTab('main')}
            className={cn(
              'flex-1 py-1.5 px-3 rounded-md text-center transition-colors cursor-pointer text-xs font-semibold',
              mobileTab === 'main'
                ? 'bg-accent text-accent-foreground shadow-xs'
                : 'text-muted-foreground hover:text-foreground hover:bg-accent/50',
            )}
          >
            {ts('main')}
          </button>
          <button
            type="button"
            onClick={() => setMobileTab('domain')}
            className={cn(
              'flex-1 py-1.5 px-3 rounded-md text-center transition-colors cursor-pointer text-xs font-semibold truncate',
              mobileTab === 'domain'
                ? 'bg-accent text-accent-foreground shadow-xs'
                : 'text-muted-foreground hover:text-foreground hover:bg-accent/50',
            )}
          >
            {activeItem ? t(activeItem.titleKey) : ts('domain')}
          </button>
        </div>
      )}

      {/* This is the first sidebar */}
      {/* We disable collapsible and adjust width to icon. */}
      {/* This will make the sidebar appear as icons. */}
      <Sidebar
        collapsible="none"
        className={cn(
          'w-[calc(var(--sidebar-width-icon)+1px)]! border-r',
          mobileTab === 'main' ? 'flex w-full' : 'hidden md:flex',
        )}
      >
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild className="md:h-8 md:p-0">
                <Link href="/d/console">
                  <div className="text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                    <Image
                      src="/assets/sz.png"
                      alt="logo"
                      width={100}
                      height={100}
                      loading="eager"
                    />
                  </div>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent className="px-1.5 md:px-0">
              <SidebarMenu>
                {navMain.map((item, index) => (
                  <SidebarMenuItem key={item.titleKey}>
                    <SidebarMenuButton
                      tooltip={{
                        children: t(item.titleKey),
                        hidden: false,
                      }}
                      onClick={() => handlerSidebarNavMainClick(item, index)}
                      isActive={activeItem?.titleKey === item.titleKey}
                      className="px-2.5 md:px-2 cursor-pointer"
                    >
                      <item.icon />
                      <span>{t(item.titleKey)}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          <NavUser user={user} />
        </SidebarFooter>
      </Sidebar>

      {/* This is the second sidebar */}
      {/* We disable collapsible and let it fill remaining space */}
      <Sidebar
        collapsible="none"
        className={cn('flex-1', mobileTab === 'domain' ? 'flex w-full' : 'hidden md:flex')}
      >
        <SidebarHeader className=" gap-[4.8px] border-b p-3">
          {isMobile && (
            <button
              type="button"
              onClick={() => setMobileTab('main')}
              className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground cursor-pointer transition-colors font-medium -mb-1"
            >
              <ChevronLeft className="size-4" />
              <span>{ts('backToMain')}</span>
            </button>
          )}
          <TeamSwitcher />
          {loading ? (
            <SidebarLoading />
          ) : (
            <div className="text-foreground text-base font-medium">
              {activeItem && t(activeItem.titleKey)}
            </div>
          )}
          <SidebarGroup>
            <SidebarGroupContent>
              <div className="relative">
                <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2" />

                <SidebarInput
                  placeholder={ts('search')}
                  className="pl-8"
                  onChange={(e) => {
                    setSearch(e.target.value);
                  }}
                />
              </div>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup className="px-0">
            <SidebarGroupContent className="px-1.5 md:px-0 flex justify-center">
              <SidebarMenu className="w-11/12">
                {loading ? (
                  <SidebarLoading />
                ) : (
                  childItems?.map((item) => (
                    <SidebarMenuItem key={item.titleKey}>
                      <SidebarMenu>
                        {/* A tab with sub tabs always shows them */}
                        <SidebarMenuItem>
                          <SidebarMenuButton
                            tooltip={t(item.titleKey)}
                            size="default"
                            onClick={() => handlerSidebarChildrenClick(item)}
                            className="px-2.5 md:px-2 cursor-pointer"
                            isActive={activeUrl === item.url}
                          >
                            {item.icon && <item.icon />}
                            <span>{t(item.titleKey)}</span>
                          </SidebarMenuButton>
                          {!!item.items?.length && (
                            <SidebarMenuSub>
                              {item.items.map((subItem) => (
                                <SidebarMenuSubItem key={subItem.titleKey}>
                                  <SidebarMenuSubButton
                                    asChild
                                    isActive={activeUrl === subItem.url}
                                    onClick={() => navigateTo(subItem.url)}
                                    className="px-2.5 md:px-2 cursor-pointer"
                                  >
                                    <div>
                                      {subItem.icon && <subItem.icon />}
                                      <span
                                        className={activeUrl === subItem.url ? 'font-semibold' : ''}
                                      >
                                        {t(subItem.titleKey)}
                                      </span>
                                    </div>
                                  </SidebarMenuSubButton>
                                </SidebarMenuSubItem>
                              ))}
                            </SidebarMenuSub>
                          )}
                        </SidebarMenuItem>
                      </SidebarMenu>
                    </SidebarMenuItem>
                  ))
                )}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
      </Sidebar>
      {/* Inside the edge: this sidebar clips its overflow, which would cut the
          default rail (centred on the edge) in half */}
      <SidebarRail className="translate-x-0 group-data-[side=left]:right-0 after:left-auto after:right-0" />
    </Sidebar>
  );
}

export function ConsoleSidebarInset({ children }: { children: ReactNode }) {
  const user = useLocalUser();

  return (
    // min-w-0: a wide table scrolls inside the page instead of widening it
    <SidebarInset className="min-w-0">
      <div className="flex min-h-screen flex-col">
        <PreviewHeader username={user.name} />
        <TabLoading>{children}</TabLoading>
      </div>
      <PreviewFooter />
    </SidebarInset>
  );
}
