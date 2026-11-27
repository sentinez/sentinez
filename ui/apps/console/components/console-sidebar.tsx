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
  useSidebar,
} from '@sentinez/ui/components/sidebar';
import Image from 'next/image';
import { useTranslations } from 'next-intl';
import { NavItem } from '@/lib/default';
import { DOMAIN_NAV, ROOT_NAV, withDomain } from '@/lib/nav';
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
import { useDebounced } from '@/hooks/use-debounced';
import { useDomain } from '@/hooks/use-domain';
import { isUnder } from '@/lib/session';
import { cn } from '@sentinez/ui/lib/utils';
import { BAR_HEIGHT, PAGE_HEADER_HEIGHT } from '@/lib/layout-bars';
import { ResizableSidebarRail } from './resizable-sidebar';

type SidebarProps = ComponentProps<typeof Sidebar>;

/** Sidebar of the tenant-level pages (`/d/console/(root)`, `/d/setting`). */
export function RootSidebar(props: SidebarProps) {
  return <ConsoleSidebar navMain={ROOT_NAV} {...props} />;
}

/** Sidebar of the domain pages (`/d/console/[domain]/...`). */
export function DomainSidebar(props: SidebarProps) {
  const domain = useDomain();
  const navMain = useMemo(() => (domain ? withDomain(DOMAIN_NAV, domain) : DOMAIN_NAV), [domain]);

  return <ConsoleSidebar navMain={navMain} {...props} />;
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

/**
 * Two-level console sidebar: an icon rail with `navMain`, and a panel with the
 * children of the active item (searchable, one level of collapsible sub items).
 */
export function ConsoleSidebar({ navMain, ...props }: SidebarProps & { navMain: NavItem[] }) {
  const t = useTranslations('Nav');
  const ts = useTranslations('Sidebar');
  // filters the items of the active tab
  const [search, setSearch] = useState('');
  const pathname = usePathname();

  const [, startTransitionNavMain] = useTransition();
  const [, startTransitionChildren] = useTransition();

  const [loading, setLoading] = useState(true);
  const [navIndex, setNavIndex] = useState(0);
  const activeItem = navMain[navIndex];
  // The url of the selected tab or sub tab; set on click so the selection
  // moves at once, then synced from the pathname.
  const [activeUrl, setActiveUrl] = useState('');
  const { setOpen, setOpenMobile, isMobile } = useSidebar();
  const user = useLocalUser();

  // the items of the active tab, filtered by the search once typing stops
  const words = useDebounced(search).trim().toLowerCase();
  const childItems = useMemo(() => {
    const items = activeItem?.items ?? [];
    return words ? items.filter((item) => t(item.titleKey).toLowerCase().includes(words)) : items;
  }, [activeItem, words, t]);

  useEffect(() => {
    let best: { index: number; url: string } | undefined;
    for (const [index, item] of navMain.entries()) {
      const url = bestUrl(item, pathname);
      if (url && url.length > (best?.url.length ?? 0)) best = { index, url };
    }

    setActiveUrl(best?.url ?? '');
    if (best) setNavIndex(best.index);

    // Always leave the skeleton state, even when no nav item/child matched.
    setLoading(false);
  }, [pathname, navMain]);

  // The tabs are Links (they navigate and prefetch); the click handlers only
  // move the selection at once and close the mobile panel.
  const handlerSidebarNavMainClick = (item: NavItem, index: number) => {
    startTransitionNavMain(() => {
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
    <Sidebar collapsible="icon" {...props}>
      {/* The two sidebars side by side. A row of its own, not the inner block
          of the shadcn Sidebar: on mobile that block is another element of
          the sheet, which is portaled out of the wrapper that declares
          --sidebar-width-icon, so it is declared again here */}
      <div
        className="flex size-full min-h-0 overflow-hidden"
        style={{ '--sidebar-width-icon': '3rem' } as CSSProperties}
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
                        <Link
                          href={item.url}
                          onClick={() => handlerSidebarNavMainClick(item, index)}
                        >
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
          {/* Two blocks as tall as the bars of the page column, so that their
            borders line up: the team over PreviewHeader, the title over
            PageLayoutHeader */}
          <SidebarHeader
            // its bottom border inside its height, as the bars have theirs: on
            // the line under PreviewHeader
            className={cn('shrink-0 justify-center border-b px-2 py-0 md:px-3', BAR_HEIGHT)}
          >
            <TeamSwitcher />
          </SidebarHeader>
          <SidebarHeader
            className={cn(
              // under the border of the team block; the height (its bottom
              // border included) lines that one up with PageLayoutHeader.
              // Inside, a 16px title, 8px and the 32px search are centred
              // between the two borders
              'shrink-0 justify-center gap-2 border-b px-2 py-0 md:px-3',
              PAGE_HEADER_HEIGHT,
            )}
          >
            {loading ? (
              <SidebarLoading className="h-4" />
            ) : (
              <div className="text-foreground text-center text-base leading-none font-medium">
                {activeItem && t(activeItem.titleKey)}
              </div>
            )}
            <SidebarGroup className="p-0">
              <SidebarGroupContent>
                <div className="relative">
                  <Search className="absolute top-1/2 left-2.5 size-4 -translate-y-1/2" />
                  <SidebarInput
                    placeholder={ts('search')}
                    className="pl-8"
                    onChange={(e) => setSearch(e.target.value)}
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
                                      <Link
                                        href={subItem.url}
                                        onClick={() => selectUrl(subItem.url)}
                                      >
                                        {subItem.icon && <subItem.icon />}
                                        <span
                                          className={
                                            activeUrl === subItem.url ? 'font-semibold' : ''
                                          }
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
      </div>
      <ResizableSidebarRail />
    </Sidebar>
  );
}

export function ConsoleSidebarInset({ children }: { children: ReactNode }) {
  const user = useLocalUser();

  return (
    // min-w-0: a wide table scrolls inside the page instead of widening it
    <SidebarInset className="min-w-0">
      <div className="flex min-h-dvh flex-col">
        <PreviewHeader username={user.name} />
        <TabLoading>{children}</TabLoading>
      </div>
      <PreviewFooter />
    </SidebarInset>
  );
}
