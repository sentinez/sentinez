'use client';

import { useTranslations } from 'next-intl';
import { ArrowRight } from 'lucide-react';

import { Badge } from '@sentinez/ui/components/badge';
import { type Location, type Upstream } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { balanceLabelKey, schemeOf } from '@/lib/type/edge';

/**
 * The locations of a resource, as the edge routes them: the path prefix,
 * its rewrite, the upstreams and the headers set on the proxied requests.
 */
export function ResourceRouting({ locations }: { locations: Location[] }) {
  const t = useTranslations('ResourceSetting');

  if (locations.length === 0) {
    return (
      <div className="text-sm text-muted-foreground p-4 border rounded-md">{t('noLocations')}</div>
    );
  }

  return (
    <ul className="divide-y rounded-md border">
      {locations.map((l, i) => (
        <LocationItem key={`${l.location}-${i}`} location={l} />
      ))}
    </ul>
  );
}

// One line per location: path (-> rewrite) and its upstreams, then its
// headers when it sets any
function LocationItem({ location: l }: { location: Location }) {
  const t = useTranslations('ResourceSetting');
  const headers = Object.entries(l.proxySetHeaders);

  return (
    <li className="flex flex-col gap-2 p-3 min-w-0">
      <div className="flex flex-wrap items-center gap-2 text-sm">
        <span className="font-mono font-semibold break-all">{l.location}</span>
        {l.proxyRewrite && l.proxyRewrite !== l.location && (
          <span className="flex items-center gap-1 font-mono break-all" title={t('rewrite')}>
            <ArrowRight className="size-3.5 text-muted-foreground" aria-label={t('rewrite')} />
            {l.proxyRewrite}
          </span>
        )}
        <span className="text-muted-foreground" aria-hidden>
          &middot;
        </span>
        {l.proxyPass.length === 0 ? (
          <span className="text-muted-foreground">{t('noUpstreams')}</span>
        ) : (
          l.proxyPass.map((u, i) => <UpstreamBadge key={`${u.server}-${i}`} upstream={u} />)
        )}
        {/* The strategy only picks between several upstreams */}
        {l.proxyPass.length > 1 && (
          <Badge variant="outline" className="ml-auto">
            {t(balanceLabelKey(l.balanceStrategy))}
          </Badge>
        )}
      </div>

      {headers.length > 0 && (
        <dl
          aria-label={t('headers')}
          className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-0.5 font-mono text-xs"
        >
          {headers.map(([name, value]) => (
            <div key={name} className="contents">
              <dt className="text-muted-foreground">{name}</dt>
              <dd className="break-all">{value}</dd>
            </div>
          ))}
        </dl>
      )}
    </li>
  );
}

function UpstreamBadge({ upstream: u }: { upstream: Upstream }) {
  return (
    <Badge variant="secondary" className="font-mono max-w-full">
      <span className="text-muted-foreground">{schemeOf(u.protocol)}://</span>
      <span className="truncate">{u.server}</span>
    </Badge>
  );
}
