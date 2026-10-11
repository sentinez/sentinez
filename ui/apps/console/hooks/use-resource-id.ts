'use client';

import { useParams } from 'next/navigation';
import { getResourceByDomain } from '@/lib/api/tenant';
import { useApi } from '@/hooks/use-api';

/**
 * The id of the resource behind the `[domain]` route segment, the
 * `metadata.resourceId` of the rows that belong to it. Empty while loading or
 * when the domain has no resource.
 */
export function useResourceId() {
  const { domain = '' } = useParams<{ domain?: string }>();
  const { data, isLoading } = useApi(getResourceByDomain, {
    resourceDomain: decodeURIComponent(domain),
  });

  return { resourceId: data?.resource?.id ?? '', isLoading };
}
