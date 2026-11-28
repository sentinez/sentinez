'use client';

import type { Notification } from '@sentinez/proto/sentinez/apps/communication/v1/model';
import { NotiType } from '@sentinez/proto/sentinez/types/notify/v1/notify';
import { Bell, CircleAlert, CircleCheck, CircleX, Info, type LucideIcon } from 'lucide-react';
import { useFormatter, useTranslations } from 'next-intl';
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
import { Skeleton } from '@sentinez/ui/components/skeleton';
import { cn } from '@sentinez/ui/lib/utils';

import { useNotifications } from '@/hooks/use-notifications';

type NotificationIcon = { icon: LucideIcon; className: string };

// The icon of a notification and its color, by its type; _defaultIcon for
// the others (unspecified, unknown)
const _defaultIcon: NotificationIcon = { icon: Info, className: 'text-muted-foreground' };
const _icons: Partial<Record<NotiType, NotificationIcon>> = {
  [NotiType.NOTI_TYPE_INFO]: { icon: Info, className: 'text-blue-600 dark:text-blue-400' },
  [NotiType.NOTI_TYPE_SUCCESS]: {
    icon: CircleCheck,
    className: 'text-emerald-600 dark:text-emerald-400',
  },
  [NotiType.NOTI_TYPE_WARNING]: {
    icon: CircleAlert,
    className: 'text-amber-600 dark:text-amber-400',
  },
  [NotiType.NOTI_TYPE_ERROR]: { icon: CircleX, className: 'text-destructive' },
};

/**
 * The notifications of the selected tenant: a bell icon, badged with those
 * pushed since the drawer was last opened, opening them in a drawer from
 * the right (from the bottom on a phone).
 */
export function Notifications() {
  const t = useTranslations('Sidebar');
  const { isMobile } = useSidebar();
  const [open, setOpen] = useState(false);
  const { notifications, isLoading, unread, markRead } = useNotifications();

  const onOpenChange = (next: boolean) => {
    setOpen(next);
    if (next) markRead();
  };

  return (
    <>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="relative size-8 cursor-pointer text-muted-foreground"
        aria-label={
          unread
            ? `${t('notifications')} (${t('notificationsUnread', { count: unread })})`
            : t('notifications')
        }
        onClick={() => onOpenChange(true)}
      >
        <Bell />
        {unread > 0 && (
          <span
            aria-hidden
            className="bg-destructive absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] leading-none font-medium text-white"
          >
            {unread > 9 ? '9+' : unread}
          </span>
        )}
      </Button>
      <Drawer direction={isMobile ? 'bottom' : 'right'} open={open} onOpenChange={onOpenChange}>
        <DrawerContent className="data-[vaul-drawer-direction=bottom]:min-h-[50dvh]">
          <DrawerHeader>
            <DrawerTitle>{t('notifications')}</DrawerTitle>
            <DrawerDescription>{t('notificationsDescription')}</DrawerDescription>
          </DrawerHeader>
          {isLoading && notifications.length === 0 ? (
            <NotificationSkeleton />
          ) : notifications.length === 0 ? (
            <div className="text-muted-foreground flex flex-1 flex-col items-center justify-center gap-2 p-4 text-sm">
              <Bell className="size-8 opacity-50" />
              {t('notificationsEmpty')}
            </div>
          ) : (
            <ul className="flex flex-1 flex-col gap-1 overflow-y-auto p-2">
              {notifications.map((n) => (
                <NotificationItem key={n.id} notification={n} />
              ))}
            </ul>
          )}
        </DrawerContent>
      </Drawer>
    </>
  );
}

function NotificationItem({ notification }: { notification: Notification }) {
  const format = useFormatter();
  const { event, metadata } = notification;
  const { icon: Icon, className } = (event && _icons[event.notiType]) ?? _defaultIcon;
  const createdAt = metadata?.createdAt;

  return (
    <li className="hover:bg-muted/50 flex gap-3 rounded-md p-2">
      <Icon className={cn('mt-0.5 size-4 shrink-0', className)} aria-hidden />
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <div className="flex items-baseline justify-between gap-2">
          <span className="truncate text-sm font-medium">{event?.title}</span>
          {createdAt && (
            <time
              dateTime={createdAt.toISOString()}
              title={format.dateTime(createdAt, { dateStyle: 'medium', timeStyle: 'short' })}
              className="text-muted-foreground shrink-0 text-xs"
            >
              {format.relativeTime(createdAt)}
            </time>
          )}
        </div>
        {event?.message && (
          <p className="text-muted-foreground text-xs break-words">{event.message}</p>
        )}
      </div>
    </li>
  );
}

function NotificationSkeleton() {
  return (
    <div className="flex flex-col gap-3 p-4" aria-busy>
      {Array.from({ length: 3 }, (_, i) => (
        <div key={i} className="flex gap-3">
          <Skeleton className="size-4 rounded-full" />
          <div className="flex flex-1 flex-col gap-1.5">
            <Skeleton className="h-4 w-2/3" />
            <Skeleton className="h-3 w-full" />
          </div>
        </div>
      ))}
    </div>
  );
}
