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
import { dashboard } from '@/lib/default';
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
import { ChevronRight, Search } from 'lucide-react';
import SidebarLoading from './sidebar-loading';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@sentinez/ui/components/collapsible';
import {
  ComponentProps,
  Fragment,
  JSX,
  ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useState,
  useTransition,
} from 'react';

export function DomainSidebar({ ...props }: ComponentProps<typeof Sidebar>) {
  // Note: I'm using state to show active item.
  // IRL you should use the url/router.
  const router = useRouter();
  const pathname = usePathname();
  const domain = pathname.split('/')[2];

  const formatUrl = useCallback((url: string, d: string) => {
    if (!d || !url) return url;
    if (url.startsWith('/console/')) {
      const parts = url.split('/');
      if (parts[2] !== d) {
        return `/console/${d}/${parts.slice(2).join('/')}`;
      }
    }
    return url;
  }, []);

  const mapNavItems = useCallback(
    (items: any[], d: string): any[] => {
      if (!items) return items;
      return items.map((item) => ({
        ...item,
        url: formatUrl(item.url, d),
        items: item.items ? mapNavItems(item.items, d) : undefined,
      }));
    },
    [formatUrl],
  );

  const navMain = useMemo(() => {
    if (!domain) return dashboard.domainNavMain;
    return mapNavItems(dashboard.domainNavMain, domain);
  }, [domain, mapNavItems]);

  const [, startTransitionNavMain] = useTransition();
  const [, startTransitionChildren] = useTransition();

  const [loading, setLoading] = useState(true);
  const [activeItem, setActiveItem] = useState(navMain[0]);
  const [childItems, setChildItems] = useState(navMain[0]?.items);
  const [navIndex, setNavIndex] = useState(0);
  const [tabIndex, setTabIndex] = useState(-1);
  const [subTabIndex, setSubTabIndex] = useState(-1);
  const [search, setSearch] = useState('');
  const { setOpen } = useSidebar();

  useEffect(() => {
    const handler = setTimeout(() => {
      const originalItems = navMain[navIndex]?.items || [];
      if (search.trim() === '') {
        setChildItems(originalItems);
      } else {
        const filtered = originalItems.filter((item: any) =>
          item.title.toLowerCase().includes(search.toLowerCase()),
        );
        setChildItems(filtered);
      }
    }, 300); // debounce 300ms

    return () => clearTimeout(handler); // cleanup
  }, [search, activeItem, navIndex]);

  useEffect(() => {
    for (const [index, item] of navMain.entries()) {
      const matchedChild = item.items?.find((subItem: any) => pathname.startsWith(subItem.url));
      if (pathname.startsWith(item.url) || matchedChild) {
        setActiveItem(item);
        setNavIndex(index);
        setChildItems(item.items);

        if (matchedChild) {
          const subIndex = item.items.findIndex((sub: any) => sub.url === matchedChild.url);
          setTabIndex(subIndex);

          const matchedSubChild = matchedChild.items?.find((subSubItem: any) =>
            pathname.startsWith(subSubItem.url),
          );
          if (matchedSubChild) {
            const subSubIndex = matchedChild.items.findIndex(
              (sub: any) => sub.url === matchedSubChild.url,
            );
            setSubTabIndex(subSubIndex);
          } else {
            setSubTabIndex(-1);
          }

          setLoading(false);
        }

        break;
      }
    }
  }, [pathname, navMain]);

  const handlerSidebarNavMainClick = (item: any, index: number) => {
    router.push(item.url);

    startTransitionNavMain(() => {
      setActiveItem(item);
      setChildItems(item.items);
      setTabIndex(0);
      setNavIndex(index);
      setOpen(true);
    });
  };

  const handlerSidebarChildrenClick = (item: any, index: number) => {
    router.push(item.url);

    startTransitionChildren(() => {
      setTabIndex(index);
    });
  };

  const handlerSidebarSubChildrenClick = (item: any, index: number) => {
    router.push(item.url);

    startTransitionChildren(() => {
      setSubTabIndex(index);
    });
  };

  return (
    <Sidebar
      collapsible="icon"
      className="overflow-hidden *:data-[sidebar=sidebar]:flex-row"
      {...props}
    >
      {/* This is the first sidebar */}
      {/* We disable collapsible and adjust width to icon. */}
      {/* This will make the sidebar appear as icons. */}
      <Sidebar collapsible="none" className="w-[calc(var(--sidebar-width-icon)+1px)]! border-r">
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild className="md:h-8 md:p-0">
                <Link href="/console">
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
                {navMain.map((item: any, index: number) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton
                      tooltip={{
                        children: item.title,
                        hidden: false,
                      }}
                      onClick={() => handlerSidebarNavMainClick(item, index)}
                      isActive={activeItem?.title === item.title}
                      className="px-2.5 md:px-2 cursor-pointer"
                    >
                      <item.icon />
                      <span>{item.title}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          <NavUser user={dashboard.user} />
        </SidebarFooter>
      </Sidebar>

      {/* This is the second sidebar */}
      {/* We disable collapsible and let it fill remaining space */}
      <Sidebar collapsible="none" className="hidden flex-1 md:flex">
        <SidebarHeader className="gap-3.5 border-b p-4">
          <TeamSwitcher teams={dashboard.tenant} />
          {loading ? (
            <SidebarLoading />
          ) : (
            <div className="text-foreground text-base font-medium">{activeItem?.title}</div>
          )}
          <SidebarGroup>
            <SidebarGroupContent>
              <div className="relative">
                <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2" />

                <SidebarInput
                  placeholder="Search..."
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
                  childItems?.map((item: any, index: number) => (
                    <SidebarMenuItem key={item.title}>
                      <SidebarMenu>
                        <Collapsible
                          key={item.title}
                          asChild
                          defaultOpen={item.isActive}
                          className="group/collapsible"
                        >
                          <SidebarMenuItem>
                            <CollapsibleTrigger asChild>
                              <SidebarMenuButton
                                tooltip={item.title}
                                size="default"
                                onClick={() => handlerSidebarChildrenClick(item, index)}
                                className="px-2.5 md:px-2 cursor-pointer"
                                isActive={tabIndex === index}
                              >
                                {item.icon && <item.icon />}
                                <span>{item.title}</span>
                                {item.items?.length > 0 && (
                                  <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                                )}
                              </SidebarMenuButton>
                            </CollapsibleTrigger>
                            <CollapsibleContent>
                              <SidebarMenuSub>
                                {item.items?.map((subItem: any, subIndex: number) => (
                                  <SidebarMenuSubItem key={subItem.title}>
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
                                        <span>{subItem.title}</span>
                                      </div>
                                    </SidebarMenuSubButton>
                                  </SidebarMenuSubItem>
                                ))}
                              </SidebarMenuSub>
                            </CollapsibleContent>
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

export function DomainSidebarInset({ children }: { children: ReactNode }) {
  const [breadcrumbs, setBreadcrumbs] = useState<JSX.Element[]>([]);
  const pathname = usePathname();

  useEffect(() => {
    const pathArray = pathname.split('/').filter((path) => path !== '');
    const components = pathArray.map((path, index) => {
      return (
        <>
          <BreadcrumbItem
            className={cn(
              'hidden md:block',
              index == pathArray.length - 1 && 'text-black dark:text-white',
            )}
          >
            {path}
          </BreadcrumbItem>
          {index < pathArray.length - 1 && <BreadcrumbSeparator className="hidden md:block" />}
        </>
      );
    });

    setBreadcrumbs(components);
  }, [pathname]);

  return (
    <SidebarInset>
      <PreviewHeader username={dashboard.user.name} />
      <div className="bg-background sticky top-0 flex shrink-0 items-center gap-2 border-b p-2 z-2">
        <SidebarTrigger className="-ml-1 cursor-pointer" />
        <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
        <Breadcrumb>
          <BreadcrumbList>
            {breadcrumbs.map((item, index) => (
              <Fragment key={index}>{item}</Fragment>
            ))}
          </BreadcrumbList>
        </Breadcrumb>
      </div>
      <div className="flex flex-1 flex-col gap-4 p-4 max-w-7xl">{children}</div>
      <PreviewFooter />
    </SidebarInset>
  );
}
