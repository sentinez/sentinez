'use client';

import { DomainEventExplorer } from '../domain-explorer';
import { useRulesetConfig } from '@/components/event-explorer/kinds/ruleset';

export default function View() {
  const config = useRulesetConfig();
  return <DomainEventExplorer config={config} />;
}
