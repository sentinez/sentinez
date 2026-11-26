'use client';

import { useTranslations } from 'next-intl';
import { BadgeCheck, Bell, Building2, LogOut } from 'lucide-react';
import Link from 'next/link';
import { useState } from 'react';

import { Avatar, AvatarFallback, AvatarImage } from '@sentinez/ui/components/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from '@sentinez/ui/components/drawer';
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@sentinez/ui/components/sidebar';
import { useLogout } from '@/components/user-menu';

export function NavUser({
  user,
}: {
  user: {
    name: string;
    email: string;
    avatar: string;
  };
}) {
  const t = useTranslations('Sidebar');
  const { isMobile } = useSidebar();
  const handleLogout = useLogout();
  const [notificationsOpen, setNotificationsOpen] = useState(false);

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="default"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground size-8 justify-center p-0"
            >
              <Avatar className="h-8 w-8 rounded-lg">
                <AvatarImage src={user.avatar || undefined} alt={user.name} />
                <AvatarFallback className="rounded-lg">
                  {(user.name || user.email).charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side={isMobile ? 'bottom' : 'right'}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                <Avatar className="h-8 w-8 rounded-lg">
                  <AvatarImage src={user.avatar || undefined} alt={user.name} />
                  <AvatarFallback className="rounded-lg">
                    {(user.name || user.email).charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">{user.name}</span>
                  <span className="truncate text-xs">{user.email}</span>
                </div>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem asChild className="cursor-pointer">
                <Link href="/d/setting/account">
                  <BadgeCheck />
                  {t('account')}
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild className="cursor-pointer">
                <Link href="/d/setting/tenant">
                  <Building2 />
                  {t('tenant')}
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem
                className="cursor-pointer"
                onSelect={() => setNotificationsOpen(true)}
              >
                <Bell />
                {t('notifications')}
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout} className="cursor-pointer">
              <LogOut />
              {t('logout')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
      <Drawer
        direction={isMobile ? 'bottom' : 'right'}
        open={notificationsOpen}
        onOpenChange={setNotificationsOpen}
      >
        <DrawerContent className="data-[vaul-drawer-direction=bottom]:min-h-[50dvh]">
          <DrawerHeader>
            <DrawerTitle>{t('notifications')}</DrawerTitle>
            <DrawerDescription>{t('notificationsDescription')}</DrawerDescription>
          </DrawerHeader>
          <div className="text-muted-foreground flex flex-1 flex-col items-center justify-center gap-2 p-4 text-sm">
            <Bell className="size-8 opacity-50" />
            {t('notificationsEmpty')}
          </div>
        </DrawerContent>
      </Drawer>
    </SidebarMenu>
  );
}
