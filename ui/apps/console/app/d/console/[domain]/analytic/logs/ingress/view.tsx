'use client';

import { DomainEventExplorer } from '../domain-explorer';
import { useIngressConfig } from '@/components/event-explorer/kinds/ingress';

export default function View() {
  const config = useIngressConfig();
  return <DomainEventExplorer config={config} />;
}
