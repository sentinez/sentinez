'use client';

import { Bell } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useState } from 'react';

import { Button } from '@sentinez/ui/components/button';
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from '@sentinez/ui/components/drawer';
import { useSidebar } from '@sentinez/ui/components/sidebar';

/**
 * The notifications: a bell icon opening them in a drawer, from the right
 * (from the bottom on a phone).
 */
export function Notifications() {
  const t = useTranslations('Sidebar');
  const { isMobile } = useSidebar();
  const [open, setOpen] = useState(false);

  return (
    <>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="size-8 cursor-pointer text-muted-foreground"
        aria-label={t('notifications')}
        onClick={() => setOpen(true)}
      >
        <Bell />
      </Button>
      <Drawer direction={isMobile ? 'bottom' : 'right'} open={open} onOpenChange={setOpen}>
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
    </>
  );
}
