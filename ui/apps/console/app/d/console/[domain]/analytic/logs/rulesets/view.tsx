'use client';

import { EventExplorer } from '../_explorer/event-explorer';
import { useCoreRuleConfig } from '../_explorer/kinds/core-rule';

export default function View() {
  const config = useCoreRuleConfig();
  return <EventExplorer config={config} />;
}
