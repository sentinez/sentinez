'use client';

import { DomainEventExplorer } from '../domain-explorer';
import { useSecRuleConfig } from '@/components/event-explorer/kinds/sec-rule';

export default function View() {
  const config = useSecRuleConfig('rateLimit');
  return <DomainEventExplorer config={config} />;
}
