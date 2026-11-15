'use client';

import { Separator } from '@sentinez/ui/components/separator';
import IsLoading from '@sentinez/ui/components/common/loading';
import BadgeStatus from '@/components/badge-status';
import { statusLabel } from '@/lib/type/security';
import { useResource } from '../[domain]/resource-context';
import { ResourceLatency } from './resource-latency';
import { ResourceUnavailable } from './resource-unavailable';
import { ResourceVisitors } from './resource-visitors';

export function ResourceView() {
  const { domain, resource, isLoading, error } = useResource();

  if (isLoading) {
    return <IsLoading />;
  }

  if (error || !resource) {
    return <ResourceUnavailable domain={domain} error={error} />;
  }

  return (
    <div className="flex flex-col gap-6 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{resource.server?.name}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {resource.server?.resourceDomain} &mdash; {resource.id}
          </p>
        </div>
        <BadgeStatus status={statusLabel(resource.status)} value={statusLabel(resource.status)} />
      </div>

      <Separator />

      <ResourceLatency resourceId={resource.id} />
      <ResourceVisitors resourceId={resource.id} />
    </div>
  );
}
