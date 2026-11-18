'use client';

import { DomainEventExplorer } from '../domain-explorer';
import { useEgressConfig } from '@/components/event-explorer/kinds/egress';

export default function View() {
  const config = useEgressConfig();
  return <DomainEventExplorer config={config} />;
}
