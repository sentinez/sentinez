'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { PencilIcon } from 'lucide-react';

import { Button } from '@sentinez/ui/components/button';
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import IsLoading from '@sentinez/ui/components/common/loading';
import type { Resource } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { updateResource } from '@/lib/api/tenant';
import { toast } from '@/lib/toast';
import { ResourceOverview } from '../../../components/resource-overview';
import { ResourceUnavailable } from '../../../components/resource-unavailable';
import { ResourceRouting } from '../../../components/resource-routing';
import { ResourceSettingJson } from '../../../components/resource-setting-json';
import {
  ServerFields,
  serverFormOf,
  serverOf,
  validateServerForm,
  type ServerFormValue,
} from '../../../components/server-fields';
import { useResource } from '../../resource-context';

/** The edge setting of the domain: its overview, then its routing, shown or edited in place */
export function SettingView() {
  const { domain, resource, isLoading, error, refresh } = useResource();

  if (isLoading) {
    return <IsLoading />;
  }

  if (error || !resource) {
    return <ResourceUnavailable domain={domain} error={error} />;
  }

  return (
    <div className="flex flex-col gap-4 w-full max-w-6xl p-4 animate-in fade-in duration-500">
      <ResourceOverview resource={resource} />
      {/* Keyed by the update time: a saved setting starts again from the stored one */}
      <RoutingCard
        key={resource.metadata?.updatedAt?.getTime()}
        resource={resource}
        onSaved={refresh}
      />
    </div>
  );
}

/** The locations of the resource; Edit swaps them for the form changing the whole Server */
function RoutingCard({ resource, onSaved }: { resource: Resource; onSaved: () => Promise<void> }) {
  const t = useTranslations('ResourceSetting');
  const tc = useTranslations('Common');
  const tv = useTranslations('Validation');
  const [form, setForm] = useState<ServerFormValue | null>(null);
  const [saving, setSaving] = useState(false);
  const locations = resource.server?.locations ?? [];

  const save = async (value: ServerFormValue) => {
    const invalid = validateServerForm(value, tv);
    if (invalid) {
      toast.error(invalid);
      return;
    }

    setSaving(true);
    try {
      // UpdateResource replaces both the server and the status
      await updateResource({
        id: resource.id,
        server: serverOf(value, resource.server),
        status: resource.status,
      });
      toast.success(t('saved'));
      await onSaved();
    } catch {
      toast.error(t('saveFailed'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card className="gap-4">
      <CardHeader>
        <CardTitle>{t('routing')}</CardTitle>
        <CardDescription>
          {form ? t('editDescription') : t('routingDescription', { count: locations.length })}
        </CardDescription>
        {!form && (
          <CardAction>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setForm(serverFormOf(resource.server))}
            >
              <PencilIcon className="w-4 h-4" />
              {t('edit')}
            </Button>
          </CardAction>
        )}
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {form ? (
          <ServerFields
            value={form}
            onChange={(patch) => setForm((f) => f && { ...f, ...patch })}
          />
        ) : (
          <>
            <ResourceRouting locations={locations} />
            <ResourceSettingJson server={resource.server} />
          </>
        )}
      </CardContent>
      {form && (
        <CardFooter className="flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-muted-foreground max-w-xl">{t('syncHint')}</p>
          <div className="flex gap-2">
            <Button type="button" variant="outline" disabled={saving} onClick={() => setForm(null)}>
              {tc('cancel')}
            </Button>
            <Button type="button" disabled={saving} onClick={() => save(form)}>
              {saving ? tc('saving') : tc('save')}
            </Button>
          </div>
        </CardFooter>
      )}
    </Card>
  );
}
