'use client';

import { useMemo, useSyncExternalStore } from 'react';

import { ControlPlane } from '@sentinez/proto/sentinez/types/v1/known';
import { LogOut } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { Avatar, AvatarFallback, AvatarImage } from '@sentinez/ui/components/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';

import { avatarOf } from '@/lib/avatar';
import { SENTINEZ_USER_KEY } from '@/lib/const';
import { subscribeUser } from '@/lib/session';
import { toast } from '@/lib/toast';
import { UserDataLocal } from '@/lib/type/common';

const NO_USER: UserDataLocal = {
  id: '',
  name: '',
  email: '',
  avatar: '',
  controlPlane: ControlPlane.CONTROL_PLANE_UNSPECIFIED,
  createdAt: '',
  updatedAt: '',
  fullName: ''
};

const storedUser = () => localStorage.getItem(SENTINEZ_USER_KEY);
const serverUser = () => null;

/**
 * The user the login forms stored, with a generated avatar (empty while
 * rendering on the server)
 */
export function useLocalUser(): UserDataLocal {
  const stored = useSyncExternalStore(subscribeUser, storedUser, serverUser);

  return useMemo(() => {
    try {
      if (!stored) return NO_USER;
      const user: UserDataLocal = { ...NO_USER, ...JSON.parse(stored) };
      // seeded by the id so the avatar is stable; older sessions only have the email
      return { ...user, avatar: avatarOf(user.fullName || user.id) };
    } catch {
      // a corrupt entry leaves the menu without user info; logout still works
      return NO_USER;
    }
  }, [stored]);
}

/** Clears the session and returns to the login page */
export function useLogout() {
  const t = useTranslations('Sidebar');
  const router = useRouter();

  return () => {
    localStorage.clear();
    toast.success(t('loggedOut'), t('loggedOutDescription'));
    router.push('/auth');
  };
}

/**
 * Avatar button for pages without a sidebar (e.g. /tenant): opens a menu with
 * the signed-in user's name and email, and the logout action.
 */
export function UserMenu() {
  const t = useTranslations('Sidebar');
  const user = useLocalUser();
  const logout = useLogout();
  const initial = (user.name || user.email).charAt(0).toUpperCase();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        aria-label={t('account')}
        className="focus-visible:ring-ring cursor-pointer rounded-full outline-none focus-visible:ring-2"
      >
        <Avatar className="size-8">
          <AvatarImage src={user.avatar || undefined} alt="" />
          <AvatarFallback>{initial}</AvatarFallback>
        </Avatar>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="min-w-56" align="end" sideOffset={4}>
        <DropdownMenuLabel className="grid font-normal">
          <span className="truncate text-sm font-medium">{user.name}</span>
          <span className="text-muted-foreground truncate text-xs">{user.email}</span>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={logout} className="cursor-pointer">
          <LogOut />
          {t('logout')}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
