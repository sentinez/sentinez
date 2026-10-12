'use client';

import { Check, ChevronsUpDown, Plus, Users } from 'lucide-react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@sentinez/ui/components/sidebar';
import { Skeleton } from '@sentinez/ui/components/skeleton';

import { useApi } from '@/hooks/use-api';
import { usePlanLabel } from '@/hooks/use-plan-label';
import { useTenantId } from '@/hooks/use-tenant-id';
import { listTenants } from '@/lib/api/tenant';

export function TeamSwitcher() {
  const t = useTranslations('Sidebar');
  const planLabel = usePlanLabel();
  const { isMobile } = useSidebar();
  const { data, isLoading } = useApi(listTenants, {});
  const [tenantId, selectTenant] = useTenantId();

  const tenants = data?.tenants ?? [];
  // Falls back to the first tenant when nothing was picked yet
  const activeTenant = tenants.find((tenant) => tenant.id === tenantId) ?? tenants[0];

  if (isLoading && !data) {
    return <TeamSwitcherSkeleton />;
  }

  if (!activeTenant) {
    return null;
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-sm">
                <Users className="size-4" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">{activeTenant.name || activeTenant.id}</span>
                <span className="truncate text-xs">{planLabel(activeTenant.plan)}</span>
              </div>
              <ChevronsUpDown className="ml-auto" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-sm"
            align="start"
            side={isMobile ? 'bottom' : 'right'}
            sideOffset={4}
          >
            <DropdownMenuLabel className="text-muted-foreground text-xs">
              {t('tenants')}
            </DropdownMenuLabel>
            {tenants.map((tenant) => (
              <DropdownMenuItem
                key={tenant.id}
                onClick={() => selectTenant(tenant.id)}
                className="gap-2 p-2"
              >
                <div className="flex size-6 items-center justify-center rounded-sm border">
                  <Users className="size-3.5 shrink-0" />
                </div>
                <span className="truncate">{tenant.name || tenant.id}</span>
                {tenant.id === activeTenant.id && <Check className="ml-auto size-4" />}
              </DropdownMenuItem>
            ))}
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild className="gap-2 p-2">
              <Link href="/tenant">
                <div className="flex size-6 items-center justify-center rounded-sm border bg-transparent">
                  <Plus className="size-4" />
                </div>
                <div className="text-muted-foreground font-medium">{t('addTeam')}</div>
              </Link>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}

/** Same footprint as the switcher button, shown while the tenants load */
function TeamSwitcherSkeleton() {
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <div className="flex h-12 items-center gap-2 p-2 group-data-[collapsible=icon]:p-0">
          <Skeleton className="bg-sidebar-accent size-8 shrink-0 rounded-sm" />
          <div className="grid flex-1 gap-1.5 group-data-[collapsible=icon]:hidden">
            <Skeleton className="bg-sidebar-accent h-3.5 w-24" />
            <Skeleton className="bg-sidebar-accent h-3 w-14" />
          </div>
        </div>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
