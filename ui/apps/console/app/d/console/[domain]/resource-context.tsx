'use client';

import { createContext, useContext, useMemo, type ReactNode } from 'react';

import type { Resource } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { useApi } from '@/hooks/use-api';
import { getResourceByDomain } from '@/lib/api/tenant';

type ResourceState = {
  /** The `[domain]` segment, URL-encoded as in the route */
  domain: string;
  /** The resource of the domain; undefined while loading, or when it has none */
  resource: Resource | undefined;
  /** True until the resource is first fetched: a refresh keeps showing it */
  isLoading: boolean;
  error: Error | null;
  /** Fetches the resource again, e.g. after it was updated */
  refresh: () => Promise<void>;
};

const ResourceContext = createContext<ResourceState | null>(null);

/**
 * Fetches the resource of the `[domain]` segment once, for every page and
 * hook under it: the Resource and Setting tabs, and the pages scoped to a
 * resource. Mount it with `key={domain}` so a domain switch starts clean.
 */
export function ResourceProvider({ domain, children }: { domain: string; children: ReactNode }) {
  const { data, isLoading, error, refresh } = useApi(getResourceByDomain, {
    resourceDomain: decodeURIComponent(domain),
  });

  const value = useMemo<ResourceState>(
    () => ({
      domain,
      resource: data?.resource,
      isLoading: isLoading && !data,
      error,
      refresh,
    }),
    [domain, data, isLoading, error, refresh],
  );

  return <ResourceContext.Provider value={value}>{children}</ResourceContext.Provider>;
}

/** The resource of the `[domain]` segment, see ResourceProvider */
export function useResource(): ResourceState {
  const value = useContext(ResourceContext);
  if (!value) {
    throw new Error('useResource must be used inside a ResourceProvider');
  }

  return value;
}
