'use client';

import { DomainEventExplorer } from '../domain-explorer';
import { useCoreRuleConfig } from '@/components/event-explorer/kinds/core-rule';

export default function View() {
  const config = useCoreRuleConfig();
  return <DomainEventExplorer config={config} />;
}
