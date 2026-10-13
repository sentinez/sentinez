'use client';

import { useSyncExternalStore } from 'react';

import { SENTINEZ_TENANT_KEY } from '@/lib/const';

// Fired on window when the selected tenant changes in this tab
const TENANT_CHANGE_EVENT = 'sentinez:tenant-change';

/** Remembers the selected tenant (picked on /tenant or in the team switcher) */
export function selectTenant(tenantId: string) {
  localStorage.setItem(SENTINEZ_TENANT_KEY, tenantId);
  window.dispatchEvent(new Event(TENANT_CHANGE_EVENT));
}

// `storage` covers a tenant picked in another tab
function subscribe(onChange: () => void) {
  window.addEventListener(TENANT_CHANGE_EVENT, onChange);
  window.addEventListener('storage', onChange);
  return () => {
    window.removeEventListener(TENANT_CHANGE_EVENT, onChange);
    window.removeEventListener('storage', onChange);
  };
}

const currentTenantId = () => localStorage.getItem(SENTINEZ_TENANT_KEY) ?? '';
const serverTenantId = () => '';

/**
 * The selected tenant id from localStorage ('' while rendering on the server
 * or when nothing was picked yet), and a setter that also persists it. Every
 * caller follows a change made anywhere, e.g. in the team switcher.
 */
export function useTenantId() {
  const tenantId = useSyncExternalStore(subscribe, currentTenantId, serverTenantId);

  return [tenantId, selectTenant] as const;
}
