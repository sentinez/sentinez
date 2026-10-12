'use client';

import { useCallback, useEffect, useState } from 'react';

import { SENTINEZ_TENANT_KEY } from '@/lib/const';

/** Remembers the selected tenant (picked on /tenant or in the team switcher) */
export function selectTenant(tenantId: string) {
  localStorage.setItem(SENTINEZ_TENANT_KEY, tenantId);
}

/**
 * The selected tenant id, read from localStorage after mount ('' until then
 * or when nothing was picked yet), and a setter that also persists it.
 */
export function useTenantId() {
  const [tenantId, setTenantId] = useState('');

  useEffect(() => {
    setTenantId(localStorage.getItem(SENTINEZ_TENANT_KEY) ?? '');
  }, []);

  const select = useCallback((id: string) => {
    selectTenant(id);
    setTenantId(id);
  }, []);

  return [tenantId, select] as const;
}
