'use client';

import { EventExplorer } from '../_explorer/event-explorer';
import { useIngressConfig } from '../_explorer/kinds/ingress';

export default function View() {
  const config = useIngressConfig();
  return <EventExplorer config={config} />;
}
