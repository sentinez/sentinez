'use client';

import { EventExplorer } from '../_explorer/event-explorer';
import { useSecRuleConfig } from '../_explorer/kinds/sec-rule';

export default function View() {
  const config = useSecRuleConfig('rateLimit');
  return <EventExplorer config={config} />;
}
