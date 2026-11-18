'use client';

import { useMemo, useState } from 'react';
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
import { Resource } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { TlsMode } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import { edgeResourceOf, getResource, updateResource } from '@/lib/api/tenant';
import { FieldGrid } from '@/components/field-grid';
import { useTlsModes } from '@/hooks/use-tls-modes';
import { OptionSelect, statusOptions } from '@/app/d/console/(root)/domain/components';
import { toast } from '@/lib/toast';
import { ResourceOverview } from '../../../components/resource-overview';
import { ResourceUnavailable } from '../../../components/resource-unavailable';
import { ResourceRouting } from '../../../components/resource-routing';
import { ResourceSettingJson } from '../../../components/resource-setting-json';
import {
  EdgeResourceFields,
  edgeResourceFormOf,
  edgeResourceOfForm,
  validateEdgeResourceForm,
  type EdgeResourceFormValue,
} from '../../../components/edge-resource-fields';
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
    <div className="flex flex-col gap-4 animate-in fade-in duration-500">
      <ResourceOverview resource={resource} />
      <TlsCard
        key={`tls-${resource.metadata?.updatedAt?.getTime()}`}
        resource={resource}
        onSaved={refresh}
      />
      {/* Keyed by the update time: a saved setting starts again from the stored one */}
      <RoutingCard
        key={resource.metadata?.updatedAt?.getTime()}
        resource={resource}
        onSaved={refresh}
      />
    </div>
  );
}

/** The locations of the resource; Edit swaps them for the form changing the whole resource */
function RoutingCard({ resource, onSaved }: { resource: Resource; onSaved: () => Promise<void> }) {
  const t = useTranslations('ResourceSetting');
  const tc = useTranslations('Common');
  const tv = useTranslations('Validation');
  const [form, setForm] = useState<EdgeResourceFormValue | null>(null);
  const [saving, setSaving] = useState(false);
  const locations = resource.locations ?? [];
  const edge = useMemo(() => edgeResourceOf(resource), [resource]);

  const save = async (value: EdgeResourceFormValue) => {
    const invalid = validateEdgeResourceForm(value, tv);
    if (invalid) {
      toast.error(invalid);
      return;
    }

    setSaving(true);
    try {
      // The form edits the name and the routes, not the domain; the TLS mode is in TlsCard
      const { name, locations } = edgeResourceOfForm(value, edge);
      await updateResource({
        id: resource.id,
        resource: Resource.fromPartial({ name, locations }),
        updateMask: ['name', 'locations'],
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
              onClick={() => setForm(edgeResourceFormOf(edge))}
            >
              <PencilIcon className="w-4 h-4" />
              {t('edit')}
            </Button>
          </CardAction>
        )}
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {form ? (
          <EdgeResourceFields
            value={form}
            onChange={(patch) => setForm((f) => f && { ...f, ...patch })}
          />
        ) : (
          <>
            <ResourceRouting locations={locations} />
            <ResourceSettingJson resource={edge} />
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

// The status the select shows: pending is active, waiting for certificates
const wantedStatus = (status: Status) =>
  status === Status.STATUS_PENDING ? Status.STATUS_ACTIVE : status || Status.STATUS_DISABLE;

/**
 * The status and the TLS mode of the resource. Asked active, the backend
 * keeps it pending until its certificates fit the mode, then activates it
 * on its own when one is uploaded.
 */
function TlsCard({ resource, onSaved }: { resource: Resource; onSaved: () => Promise<void> }) {
  const t = useTranslations('ResourceSetting');
  const tf = useTranslations('ResourceFields');
  const ts = useTranslations('Enum.status');
  const tc = useTranslations('Common');
  const { options: tlsModes } = useTlsModes();
  const initialStatus = wantedStatus(resource.status);
  const [status, setStatus] = useState<Status>(initialStatus);
  const [tlsMode, setTlsMode] = useState<TlsMode>(resource.tlsMode);
  const [saving, setSaving] = useState(false);
  const changed = status !== initialStatus || tlsMode !== resource.tlsMode;
  const statuses = useMemo(() => statusOptions(tf, ts), [tf, ts]);

  const save = async () => {
    setSaving(true);
    try {
      await updateResource({
        id: resource.id,
        resource: Resource.fromPartial({ status, tlsMode }),
        updateMask: ['status', 'tlsMode'],
      });
      const stored = (await getResource({ id: resource.id }))?.resource;
      if (stored?.status === Status.STATUS_PENDING) {
        toast.info(t('savedPending'), t('tlsPending'));
      } else {
        toast.success(t('saved'));
      }
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
        <CardTitle>{t('tls')}</CardTitle>
        <CardDescription>{t('tlsDescription')}</CardDescription>
        {resource.status === Status.STATUS_PENDING && (
          <p className="text-sm text-amber-700 dark:text-amber-400">{t('tlsPending')}</p>
        )}
      </CardHeader>
      <CardContent>
        <FieldGrid>
          <OptionSelect
            label={tf('status')}
            value={status}
            options={statuses}
            onChange={setStatus}
          />
          <OptionSelect
            label={t('tlsMode')}
            value={tlsMode}
            options={tlsModes}
            onChange={setTlsMode}
          />
        </FieldGrid>
      </CardContent>
      {changed && (
        <CardFooter className="flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            disabled={saving}
            onClick={() => {
              setStatus(initialStatus);
              setTlsMode(resource.tlsMode);
            }}
          >
            {tc('cancel')}
          </Button>
          <Button type="button" disabled={saving} onClick={save}>
            {saving ? tc('saving') : tc('save')}
          </Button>
        </CardFooter>
      )}
    </Card>
  );
}
