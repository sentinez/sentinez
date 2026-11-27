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
import { usePathname } from 'next/navigation';
import { TeamSwitcher } from './team-switcher';
import Link from 'next/link';
import PreviewHeader from './preview-header';
import PreviewFooter from './preview-footer';
import { Search } from 'lucide-react';
import SidebarLoading from './sidebar-loading';
import { TabLoading } from './tab-loading';
import {
  ComponentProps,
  CSSProperties,
  ReactNode,
  useEffect,
  useMemo,
  useState,
  useTransition,
} from 'react';
import { useLocalUser } from '@/components/user-menu';
import { isUnder } from '@/lib/session';
import { cn } from '@sentinez/ui/lib/utils';
import { PAGE_HEADER_HEIGHT, TOP_BARS_HEIGHT } from '@/lib/layout-bars';
import { useHideOnScroll } from '@/hooks/use-hide-on-scroll';

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

/** A tab with sub tabs opens its first sub tab. */
const tabUrl = (item: NavItem) => item.items?.[0]?.url ?? item.url;

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

  // The tabs are Links (they navigate and prefetch); the click handlers only
  // move the selection at once and close the mobile panel.
  const handlerSidebarNavMainClick = (item: NavItem, index: number) => {
    startTransitionNavMain(() => {
      setActiveItem(item);
      setChildItems(item.items);
      setActiveUrl(item.url);
      setNavIndex(index);
      setOpen(true);
      // a tab without sub tabs has no panel to show: go there
      if (isMobile && !item.items?.length) setOpenMobile(false);
    });
  };

  const selectUrl = (url: string) => {
    startTransitionChildren(() => {
      setActiveUrl(url);
      if (isMobile) {
        setOpenMobile(false);
      }
    });
  };

  const handlerSidebarChildrenClick = (item: NavItem) => selectUrl(tabUrl(item));

  return (
    <Sidebar
      collapsible="icon"
      className="overflow-hidden *:data-[slot=sidebar-inner]:flex-row"
      {...props}
    >
      {/* This is the first sidebar */}
      {/* We disable collapsible and adjust width to icon. */}
      {/* This will make the sidebar appear as icons. */}
      <Sidebar
        collapsible="none"
        className="w-[calc(var(--sidebar-width-icon)+1px)]! shrink-0 border-r"
      >
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild className="h-8 p-0">
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
            <SidebarGroupContent className="px-0">
              <SidebarMenu>
                {navMain.map((item, index) => (
                  <SidebarMenuItem key={item.titleKey}>
                    <SidebarMenuButton
                      asChild
                      tooltip={{
                        children: t(item.titleKey),
                        hidden: false,
                      }}
                      isActive={activeItem?.titleKey === item.titleKey}
                      className="px-2.5 md:px-2 cursor-pointer"
                    >
                      <Link href={item.url} onClick={() => handlerSidebarNavMainClick(item, index)}>
                        <item.icon />
                        <span className="sr-only">{t(item.titleKey)}</span>
                      </Link>
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
      <Sidebar collapsible="none" className="min-w-0 flex-1">
        {/* Two blocks as tall as the bars of the page column, so that the
            bottom border lines up with the page header: the team over the top
            bars, the title and the search over the page header */}
        <SidebarHeader
          // its bottom border inside its height, as the bars have theirs: on
          // the line under BreadcrumbBar
          className={cn('shrink-0 justify-center border-b px-2 py-0 md:px-3', TOP_BARS_HEIGHT)}
        >
          <TeamSwitcher />
        </SidebarHeader>
        <SidebarHeader
          className={cn(
            // under the border of the team block; the height (its bottom
            // border included) lines that one up with PageLayoutHeader.
            // Inside, a 16px title, 8px and the 32px search are centred
            // between the two borders, 9px of room above and below
            'shrink-0 justify-center gap-2 border-b px-2 py-0 md:px-3',
            PAGE_HEADER_HEIGHT,
          )}
        >
          {loading ? (
            <SidebarLoading className="h-4" />
          ) : (
            <div className="text-foreground text-base leading-none font-medium">
              {activeItem && t(activeItem.titleKey)}
            </div>
          )}
          <SidebarGroup className="p-0">
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
              <SidebarMenu className="w-full px-1 md:w-11/12 md:px-0">
                {loading ? (
                  <SidebarLoading />
                ) : (
                  childItems?.map((item) => (
                    <SidebarMenuItem key={item.titleKey}>
                      <SidebarMenu>
                        {/* A tab with sub tabs always shows them */}
                        <SidebarMenuItem>
                          <SidebarMenuButton
                            asChild
                            tooltip={t(item.titleKey)}
                            size="default"
                            className="px-2.5 md:px-2 cursor-pointer"
                            isActive={activeUrl === item.url}
                          >
                            <Link
                              href={tabUrl(item)}
                              onClick={() => handlerSidebarChildrenClick(item)}
                            >
                              {item.icon && <item.icon />}
                              <span>{t(item.titleKey)}</span>
                            </Link>
                          </SidebarMenuButton>
                          {!!item.items?.length && (
                            <SidebarMenuSub>
                              {item.items.map((subItem) => (
                                <SidebarMenuSubItem key={subItem.titleKey}>
                                  <SidebarMenuSubButton
                                    asChild
                                    isActive={activeUrl === subItem.url}
                                    className="px-2.5 md:px-2 cursor-pointer"
                                  >
                                    <Link href={subItem.url} onClick={() => selectUrl(subItem.url)}>
                                      {subItem.icon && <subItem.icon />}
                                      <span
                                        className={activeUrl === subItem.url ? 'font-semibold' : ''}
                                      >
                                        {t(subItem.titleKey)}
                                      </span>
                                    </Link>
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
  // the top bar hides while the page is scrolled down, shows when scrolled up
  const previewHidden = useHideOnScroll();

  return (
    // min-w-0: a wide table scrolls inside the page instead of widening it.
    // --console-preview-offset: where the sticky PageLayoutHeader stops, under
    // the top bar while it shows
    <SidebarInset
      className="min-w-0"
      style={
        {
          '--console-preview-offset': previewHidden ? '0px' : 'var(--console-bar-height)',
        } as CSSProperties
      }
    >
      <div className="flex min-h-dvh flex-col">
        <PreviewHeader username={user.name} hidden={previewHidden} />
        <TabLoading>{children}</TabLoading>
      </div>
      <PreviewFooter />
    </SidebarInset>
  );
}
