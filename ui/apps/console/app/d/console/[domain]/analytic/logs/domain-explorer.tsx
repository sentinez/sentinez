'use client';

import { useMemo } from 'react';
import { EventExplorer } from '@/components/event-explorer/event-explorer';
import type { EventConfig } from '@/components/event-explorer/types';
import { useResourceId } from '@/hooks/use-resource-id';

/** The explorer of the events of the resource of the `[domain]` segment */
export function DomainEventExplorer<T>({ config }: { config: EventConfig<T> }) {
  const { resourceId } = useResourceId();
  const resourceIds = useMemo(() => [resourceId], [resourceId]);
  return <EventExplorer config={config} resourceIds={resourceIds} pending={!resourceId} />;
}
