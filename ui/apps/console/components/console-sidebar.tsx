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
  SidebarTrigger,
  useSidebar,
} from '@sentinez/ui/components/sidebar';
import Image from 'next/image';
import { useTranslations } from 'next-intl';
import { dashboard, NavItem } from '@/lib/default';
import { usePathname, useRouter } from 'next/navigation';
import { TeamSwitcher } from './team-switcher';
import { Separator } from '@sentinez/ui/components/separator';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbSeparator,
} from '@sentinez/ui/components/breadcrumb';
import { cn } from '@sentinez/ui/lib/utils';
import Link from 'next/link';
import PreviewHeader from './preview-header';
import PreviewFooter from './preview-footer';
import { ChevronLeft, ChevronRight, Search } from 'lucide-react';
import SidebarLoading from './sidebar-loading';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@sentinez/ui/components/collapsible';
import {
  ComponentProps,
  Fragment,
  ReactNode,
  useEffect,
  useMemo,
  useState,
  useTransition,
} from 'react';
import { useLocalUser } from '@/components/user-menu';

type SidebarProps = ComponentProps<typeof Sidebar>;

/** Sidebar of the tenant-level pages (`/d/console/(root)`, `/d/setting`). */
export function RootSidebar(props: SidebarProps) {
  return <ConsoleSidebar navMain={dashboard.rootNavMain} {...props} />;
}

/** Sidebar of the domain pages (`/d/console/[domain]/...`). */
export function DomainSidebar(props: SidebarProps) {
  const domain = usePathname().split('/')[3];
  const navMain = useMemo(
    () => (domain ? withDomain(dashboard.domainNavMain, domain) : dashboard.domainNavMain),
    [domain],
  );

  return <ConsoleSidebar navMain={navMain} {...props} />;
}

/** Rewrites `/d/console/<path>` urls to `/d/console/<domain>/<path>`. */
function withDomain(items: NavItem[], domain: string): NavItem[] {
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
  const [tabIndex, setTabIndex] = useState(-1);
  const [subTabIndex, setSubTabIndex] = useState(-1);
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
    for (const [index, item] of navMain.entries()) {
      const matchedChild = item.items?.find((subItem) => pathname.startsWith(subItem.url));
      if (pathname.startsWith(item.url) || matchedChild) {
        setActiveItem(item);
        setNavIndex(index);
        setChildItems(item.items);

        if (item.items && matchedChild) {
          setTabIndex(item.items.findIndex((sub) => sub.url === matchedChild.url));
          setSubTabIndex(
            matchedChild.items?.findIndex((sub) => pathname.startsWith(sub.url)) ?? -1,
          );
        }

        break;
      }
    }

    // Always leave the skeleton state, even when no nav item/child matched.
    setLoading(false);
  }, [pathname, navMain]);

  const handlerSidebarNavMainClick = (item: NavItem, index: number) => {
    router.push(item.url);

    startTransitionNavMain(() => {
      setActiveItem(item);
      setChildItems(item.items);
      setTabIndex(0);
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

  const handlerSidebarChildrenClick = (item: NavItem, index: number) => {
    router.push(item.url);

    startTransitionChildren(() => {
      setTabIndex(index);
      if (isMobile && (!item.items || item.items.length === 0)) {
        setOpenMobile(false);
      }
    });
  };

  const handlerSidebarSubChildrenClick = (item: NavItem, index: number) => {
    router.push(item.url);

    startTransitionChildren(() => {
      setSubTabIndex(index);
      if (isMobile) {
        setOpenMobile(false);
      }
    });
  };

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
                      src="/assets/sntz.png"
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
        <SidebarHeader className="gap-3.5 border-b p-4">
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
                  childItems?.map((item, index) => (
                    <SidebarMenuItem key={item.titleKey}>
                      <SidebarMenu>
                        <Collapsible
                          key={item.titleKey}
                          asChild
                          defaultOpen={item.isActive}
                          className="group/collapsible"
                        >
                          <SidebarMenuItem>
                            <CollapsibleTrigger asChild>
                              <SidebarMenuButton
                                tooltip={t(item.titleKey)}
                                size="default"
                                onClick={() => handlerSidebarChildrenClick(item, index)}
                                className="px-2.5 md:px-2 cursor-pointer"
                                isActive={tabIndex === index && subTabIndex === -1}
                              >
                                {item.icon && <item.icon />}
                                <span>{t(item.titleKey)}</span>
                                {!!item.items?.length && (
                                  <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                                )}
                              </SidebarMenuButton>
                            </CollapsibleTrigger>
                            {!!item.items?.length && (
                              <CollapsibleContent>
                                <SidebarMenuSub>
                                  {item.items.map((subItem, subIndex) => (
                                    <SidebarMenuSubItem key={subItem.titleKey}>
                                      <SidebarMenuSubButton
                                        asChild
                                        isActive={subTabIndex === subIndex}
                                        onClick={() =>
                                          handlerSidebarSubChildrenClick(subItem, subIndex)
                                        }
                                        className="px-2.5 md:px-2 cursor-pointer"
                                      >
                                        <div>
                                          {subItem.icon && <subItem.icon />}
                                          <span
                                            className={
                                              subTabIndex === subIndex ? 'font-semibold' : ''
                                            }
                                          >
                                            {t(subItem.titleKey)}
                                          </span>
                                        </div>
                                      </SidebarMenuSubButton>
                                    </SidebarMenuSubItem>
                                  ))}
                                </SidebarMenuSub>
                              </CollapsibleContent>
                            )}
                          </SidebarMenuItem>
                        </Collapsible>
                      </SidebarMenu>
                    </SidebarMenuItem>
                  ))
                )}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
      </Sidebar>
    </Sidebar>
  );
}

export function ConsoleSidebarInset({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const user = useLocalUser();

  const segments = useMemo(() => pathname.split('/').filter((path) => path !== ''), [pathname]);

  return (
    <SidebarInset>
      <div className="flex min-h-screen flex-col">
        <PreviewHeader username={user.name} />
        <div className="bg-background sticky top-0 flex shrink-0 items-center gap-2 border-b p-2 z-2">
          <SidebarTrigger className="-ml-1 cursor-pointer" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              {segments.map((path, index) => (
                <Fragment key={index}>
                  <BreadcrumbItem
                    className={cn(
                      'hidden md:block',
                      index === segments.length - 1 && 'text-black dark:text-white',
                    )}
                  >
                    {path === 'd' ? 'dashboard' : path}
                  </BreadcrumbItem>
                  {index < segments.length - 1 && (
                    <BreadcrumbSeparator className="hidden md:block" />
                  )}
                </Fragment>
              ))}
            </BreadcrumbList>
          </Breadcrumb>
        </div>
        <div className="flex flex-1 flex-col gap-4">{children}</div>
      </div>
      <PreviewFooter />
    </SidebarInset>
  );
}
