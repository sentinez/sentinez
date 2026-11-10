'use client';

import { useResource } from '@/app/d/console/[domain]/resource-context';

/**
 * The id of the resource behind the `[domain]` route segment and of its
 * tenant: the `owner.resourceId` / `owner.tenantId` of the rows that belong
 * to it. Empty while loading or when the domain has no resource.
 */
export function useResourceId() {
  const { resource, isLoading } = useResource();

  return {
    resourceId: resource?.id ?? '',
    tenantId: resource?.tenantId ?? '',
    isLoading,
  };
}
