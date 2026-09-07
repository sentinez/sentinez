'use client';

import * as React from 'react';
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

export function RootSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  // Note: I'm using state to show active item.
  // IRL you should use the url/router.
  const router = useRouter();
  const pathname = usePathname();
  const [, startTransitionNavMain] = React.useTransition();
  const [, startTransitionChildren] = React.useTransition();

  const rootNavMain = dashboard.rootNavMain;

  const [activeItem, setActiveItem] = React.useState(rootNavMain[0]);
  const [childItems, setChildItems] = React.useState(rootNavMain[0]?.items);
  const [navIndex, setNavIndex] = React.useState(0);
  const [tabIndex, setTabIndex] = React.useState(-1);
  const [search, setSearch] = React.useState('');
  const { setOpen } = useSidebar();

  React.useEffect(() => {
    const handler = setTimeout(() => {
      const originalItems = rootNavMain[navIndex]?.items || [];
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

  React.useEffect(() => {
    for (const [index, item] of rootNavMain.entries()) {
      // Use exact match for root-level URLs to avoid false positives
      const itemMatches =
        item.url === '/console'
          ? pathname === '/console' || pathname.startsWith('/console/')
          : pathname.startsWith(item.url);
      if (itemMatches) {
        setActiveItem(item);
        setNavIndex(index);
        setChildItems(item.items);

        const matchedChild = item.items?.find((subItem: any) =>
          subItem.url === '/console' ? pathname === '/console' : pathname.startsWith(subItem.url),
        );
        if (matchedChild) {
          const subIndex = item.items.findIndex((sub: any) => sub.url === matchedChild.url);
          setTabIndex(subIndex);
        } else {
          setTabIndex(-1);
        }

        break;
      }
    }
  }, [pathname]);

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
                    <Image src="/assets/sntz.png" alt="logo" width={100} height={100} />
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
                {rootNavMain.map((item: any, index: number) => (
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
          <div className="flex w-full items-center justify-between">
            <div className="text-foreground text-base font-medium">{activeItem?.title}</div>
          </div>
          <SidebarInput
            placeholder="type to search..."
            onChange={(e) => {
              setSearch(e.target.value);
            }}
          />
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup className="px-0">
            <SidebarGroupContent className="px-1.5 md:px-0 flex justify-center">
              <SidebarMenu className="w-11/12">
                {childItems?.map((item: any, index: number) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton
                      asChild
                      size="default"
                      onClick={() => handlerSidebarChildrenClick(item, index)}
                      className="px-2.5 md:px-2 cursor-pointer"
                      isActive={tabIndex === index}
                    >
                      <div>
                        <item.icon />
                        <span>{item.title}</span>
                      </div>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
      </Sidebar>
    </Sidebar>
  );
}

export function RootSidebarInset({ children }: { children: React.ReactNode }) {
  const [breadcrumbs, setBreadcrumbs] = React.useState<React.JSX.Element[]>([]);
  const pathname = usePathname();

  React.useEffect(() => {
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
      <PreviewHeader />
      <header className="bg-background sticky top-0 flex shrink-0 items-center gap-2 border-b p-2">
        <SidebarTrigger className="-ml-1 cursor-pointer" />
        <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
        <Breadcrumb>
          <BreadcrumbList>
            {breadcrumbs.map((item, index) => (
              <React.Fragment key={index}>{item}</React.Fragment>
            ))}
          </BreadcrumbList>
        </Breadcrumb>
      </header>
      <div className="flex flex-1 flex-col gap-4 p-4">{children}</div>
      <PreviewFooter />
    </SidebarInset>
  );
}
