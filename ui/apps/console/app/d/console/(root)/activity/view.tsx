'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';

import IsLoading from '@sentinez/ui/components/common/loading';
import { Button } from '@sentinez/ui/components/button';
import { ControlPlane } from '@sentinez/proto/sentinez/types/v1/known';
import { EventExplorer } from '@/components/event-explorer/event-explorer';
import { useLocalUser } from '@/components/user-menu';
import { useTenantId } from '@/hooks/use-tenant-id';
import { useAuditLogConfig } from './audit-log-config';

// The resources of the logs: all of those of the tenant
const ALL_RESOURCES: string[] = [];

// The logs of the page: those of the selected tenant, or, on the admin
// control plane, those of no tenant (logins, calls refused early...)
type Scope = 'tenant' | 'none';

/**
 * The audit log of the selected tenant: the calls of the control plane that
 * changed it or its resources, their histogram per result and the top
 * values of their fields. The admin control plane also sees the calls about
 * no tenant.
 */
export default function View() {
  const t = useTranslations('AuditLog');
  const [tenantId] = useTenantId();
  const isAdmin = useLocalUser().controlPlane === ControlPlane.CONTROL_PLANE_ADMIN;
  const [picked, setPicked] = useState<Scope>('tenant');
  // An admin without a selected tenant sees the logs of no tenant
  const scope: Scope = isAdmin && (picked === 'none' || !tenantId) ? 'none' : 'tenant';
  const config = useAuditLogConfig(scope === 'none' ? '' : tenantId);

  // Wait for the selected tenant: it is unknown while rendering on the server
  if (scope === 'tenant' && !tenantId) return <IsLoading />;

  const scopes = isAdmin && (
    <div role="group" aria-label={t('scope.label')} className="flex shrink-0 gap-1">
      {(['tenant', 'none'] as const).map((s) => (
        <Button
          key={s}
          size="sm"
          variant={scope === s ? 'secondary' : 'ghost'}
          aria-pressed={scope === s}
          disabled={s === 'tenant' && !tenantId}
          onClick={() => setPicked(s)}
        >
          {t(`scope.${s}`)}
        </Button>
      ))}
    </div>
  );

  return <EventExplorer config={config} resourceIds={ALL_RESOURCES} actions={scopes} />;
}
