'use client';

import { DomainEventExplorer } from '../domain-explorer';
import { useSecRuleConfig } from '@/components/event-explorer/kinds/sec-rule';

export default function View() {
  const config = useSecRuleConfig('secRule');
  return <DomainEventExplorer config={config} />;
}
