'use client';

import type { Notification } from '@sentinez/proto/sentinez/apps/communication/v1/model';
import axios from 'axios';
import { createContext, type ReactNode, useCallback, useContext, useEffect, useState } from 'react';

import { useTenantId } from '@/hooks/use-tenant-id';
import { StreamError, listNotifications, streamNotifications } from '@/lib/api/communication';
import { sleep } from '@/lib/delays';

// The notifications kept: the latest page, then those streamed since
const KEPT = 50;
// Reconnect delays of the stream, doubled after each failure
const RETRY_MIN_MILLIS = 1_000;
const RETRY_MAX_MILLIS = 30_000;

interface NotificationFeed {
  notifications: Notification[];
  isLoading: boolean;
  /** The notifications pushed since markRead */
  unread: number;
  markRead: () => void;
}

const NotificationsContext = createContext<NotificationFeed | null>(null);

/**
 * Holds the notifications of the selected tenant for the pages under it: the
 * latest stored ones, then those the server pushes, the newest first. It sits
 * in a layout above the section layouts, so moving between them keeps the
 * stream open and the unread count.
 */
export function NotificationsProvider({ children }: { children: ReactNode }) {
  const [tenantId] = useTenantId();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [unread, setUnread] = useState(0);

  useEffect(() => {
    if (!tenantId) return;
    const controller = new AbortController();
    const { signal } = controller;
    setIsLoading(true);
    setNotifications([]);
    setUnread(0);

    // the stored ones; one pushed meanwhile is kept once
    listNotifications(
      { tenantIds: [tenantId], page: { index: 1, size: KEPT, total: false } },
      { signal },
    )
      .then((resp) => setNotifications((pushed) => merge(pushed, resp.notifications)))
      .catch((err: unknown) => {
        // the drawer shows what it has; the bell must not toast on every page
        if (!axios.isCancel(err)) console.warn('[notifications] list failed:', err);
      })
      .finally(() => {
        if (!signal.aborted) setIsLoading(false);
      });

    // the pushed ones, of this tenant only (the server narrows the stream)
    void follow(tenantId, signal, (n) => {
      setNotifications((list) => merge([n], list));
      setUnread((count) => count + 1);
    });

    return () => controller.abort();
  }, [tenantId]);

  const markRead = useCallback(() => setUnread(0), []);

  return (
    <NotificationsContext.Provider value={{ notifications, isLoading, unread, markRead }}>
      {children}
    </NotificationsContext.Provider>
  );
}

/** The notifications of the NotificationsProvider above */
export function useNotifications(): NotificationFeed {
  const feed = useContext(NotificationsContext);
  if (!feed) throw new Error('useNotifications outside NotificationsProvider');
  return feed;
}

/** a, then the ones of b not in a, at most KEPT */
function merge(a: Notification[], b: Notification[]): Notification[] {
  const ids = new Set(a.map((n) => n.id));
  return [...a, ...b.filter((n) => !ids.has(n.id))].slice(0, KEPT);
}

/**
 * Keeps the notification stream of tenantId open until signal aborts:
 * reconnects when it ends or fails, after a delay growing up to
 * RETRY_MAX_MILLIS; gives up when the server refuses the user (401, 403).
 */
async function follow(
  tenantId: string,
  signal: AbortSignal,
  onNotification: (n: Notification) => void,
) {
  let delay = RETRY_MIN_MILLIS;
  while (!signal.aborted) {
    try {
      await streamNotifications(onNotification, {
        tenantIds: [tenantId],
        signal,
        onOpen: () => (delay = RETRY_MIN_MILLIS),
      });
    } catch (err) {
      if (signal.aborted) return;
      if (err instanceof StreamError && (err.status === 401 || err.status === 403)) return;
    }
    await sleep(delay, signal);
    delay = Math.min(delay * 2, RETRY_MAX_MILLIS);
  }
}
