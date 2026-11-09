'use client';

import IsLoading from '@sentinez/ui/components/common/loading';
import { useApi } from '@/hooks/use-api';
import { getResourceByDomain } from '@/lib/api/tenant';
import { ResourceUnavailable } from '../../../components/resource-unavailable';
import { ResourceRouting } from '../../../components/resource-routing';
import { ResourceSettingJson } from '../../../components/resource-setting-json';

/** The edge setting of the domain: its routing, then the raw JSON */
export function SettingView({ domain }: { domain: string }) {
  const { data, isLoading, error } = useApi(getResourceByDomain, { resourceDomain: domain });
  const resource = data?.resource;

  if (isLoading) {
    return <IsLoading />;
  }

  if (error || !resource) {
    return <ResourceUnavailable domain={domain} error={error} />;
  }

  return (
    <div className="flex flex-col gap-6 w-full max-w-6xl p-4 animate-in fade-in duration-500">
      <ResourceRouting locations={resource.server?.locations ?? []} />
      <ResourceSettingJson server={resource.server} />
    </div>
  );
}
