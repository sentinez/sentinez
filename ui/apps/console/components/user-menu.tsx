'use client';

import { useEffect, useState } from 'react';

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

import { SENTINEZ_USER_KEY } from '@/lib/const';
import { toast } from '@/lib/toast';

interface LocalUser {
  name: string;
  email: string;
  avatar: string;
}

// The user the login forms stored, read after mount (empty until then)
function useLocalUser(): LocalUser {
  const [user, setUser] = useState<LocalUser>({ name: '', email: '', avatar: '' });

  useEffect(() => {
    try {
      const stored = localStorage.getItem(SENTINEZ_USER_KEY);
      if (stored) setUser((prev) => ({ ...prev, ...JSON.parse(stored) }));
    } catch {
      // a corrupt entry leaves the menu without user info; logout still works
    }
  }, []);

  return user;
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
