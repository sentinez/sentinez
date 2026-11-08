'use client';

import { EventExplorer } from '../_explorer/event-explorer';
import { useEgressConfig } from '../_explorer/kinds/egress';

export default function View() {
  const config = useEgressConfig();
  return <EventExplorer config={config} />;
}
