'use client';

import { useTranslations } from 'next-intl';
import { ArrowRight } from 'lucide-react';

import { Badge } from '@sentinez/ui/components/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import {
  BalanceStrategy,
  ProxyProtocol,
  type Location,
  type Upstream,
} from '@sentinez/proto/sentinez/dmz/edge/v1/setting';

// The scheme of the protocol of an upstream, https when unspecified as the
// edge defaults it
const schemeOf = (protocol: ProxyProtocol) =>
  protocol === ProxyProtocol.PROXY_PROTOCOL_HTTP ? 'http' : 'https';

/**
 * The locations of a resource, as the edge routes them: the path prefix,
 * its rewrite, the upstreams and the headers set on the proxied requests.
 */
export function ResourceRouting({ locations }: { locations: Location[] }) {
  const t = useTranslations('Resource');

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('routing')}</CardTitle>
        <CardDescription>{t('routingDescription', { count: locations.length })}</CardDescription>
      </CardHeader>
      <CardContent>
        {locations.length === 0 ? (
          <div className="text-sm text-muted-foreground p-4 border rounded-md">
            {t('noLocations')}
          </div>
        ) : (
          <ul className="flex flex-col gap-3">
            {locations.map((l, i) => (
              <LocationItem key={`${l.location}-${i}`} location={l} />
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

function LocationItem({ location: l }: { location: Location }) {
  const t = useTranslations('Resource');
  const headers = Object.entries(l.proxySetHeaders);
  const balance =
    l.balanceStrategy === BalanceStrategy.BALANCE_STRATEGY_ROUND_ROBIN
      ? t('balanceRoundRobin')
      : t('balanceDefault');

  return (
    <li className="rounded-md border p-3 sm:p-4 space-y-3 min-w-0">
      <div className="flex flex-wrap items-center gap-2 font-mono text-sm">
        <span className="font-semibold break-all">{l.location}</span>
        {l.proxyRewrite && l.proxyRewrite !== l.location && (
          <>
            <ArrowRight className="size-3.5 text-muted-foreground" aria-hidden />
            <span className="break-all" title={t('rewrite')}>
              {l.proxyRewrite}
            </span>
          </>
        )}
        {l.proxyPass.length > 1 && (
          <Badge variant="outline" className="ml-auto font-sans">
            {balance}
          </Badge>
        )}
      </div>

      <Row label={t('upstreams')}>
        {l.proxyPass.length === 0 ? (
          <span className="text-muted-foreground">{t('noUpstreams')}</span>
        ) : (
          <div className="flex flex-wrap gap-1.5">
            {l.proxyPass.map((u, i) => (
              <UpstreamBadge key={`${u.server}-${i}`} upstream={u} />
            ))}
          </div>
        )}
      </Row>

      {headers.length > 0 && (
        <Row label={t('headers')}>
          <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 font-mono text-xs">
            {headers.map(([name, value]) => (
              <div key={name} className="contents">
                <dt className="text-muted-foreground">{name}</dt>
                <dd className="break-all">{value}</dd>
              </div>
            ))}
          </dl>
        </Row>
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

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-[8rem_1fr] gap-1 sm:gap-4 text-sm">
      <span className="font-semibold">{label}</span>
      <div className="min-w-0">{children}</div>
    </div>
  );
}
