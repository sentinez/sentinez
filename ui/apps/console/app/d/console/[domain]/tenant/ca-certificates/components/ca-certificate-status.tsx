'use client';

import { useTranslations } from 'next-intl';
import { CaCertificateUsage } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { Badge } from '@sentinez/ui/components/badge';

/** A certificate that ends within this many days is about to expire */
export const EXPIRING_SOON_DAYS = 30;
export const DAY_MS = 24 * 3600 * 1000;

export type CaExpiry = 'expired' | 'soon' | 'valid';

/** Whether the certificate has expired at `now`, ends soon, or is valid for long */
export function caExpiryOf(notAfter: Date | undefined, now = Date.now()): CaExpiry {
  if (!notAfter) return 'valid';

  const left = notAfter.getTime() - now;
  if (left <= 0) return 'expired';

  return left < EXPIRING_SOON_DAYS * DAY_MS ? 'soon' : 'valid';
}

/** Nothing for a certificate that is valid for long */
export function CaExpiryBadge({ notAfter, now }: { notAfter: Date | undefined; now?: number }) {
  const t = useTranslations('CaCertificate');
  const expiry = caExpiryOf(notAfter, now);

  if (expiry === 'expired') return <Badge variant="destructive">{t('expired')}</Badge>;
  if (expiry === 'soon') return <Badge variant="secondary">{t('expiringSoon')}</Badge>;

  return null;
}

/** The usages in the order they are offered, with the message keys that describe each */
export const CA_USAGES = [
  {
    value: CaCertificateUsage.CA_CERTIFICATE_USAGE_UPSTREAM,
    label: 'usageUpstream',
    help: 'usageUpstreamHelp',
  },
  {
    value: CaCertificateUsage.CA_CERTIFICATE_USAGE_CLIENT,
    label: 'usageClient',
    help: 'usageClientHelp',
  },
  {
    value: CaCertificateUsage.CA_CERTIFICATE_USAGE_SERVER,
    label: 'usageServer',
    help: 'usageServerHelp',
  },
] as const;

/** The usage entry, the upstream one for an unspecified usage */
const caUsageOf = (usage: CaCertificateUsage) =>
  CA_USAGES.find((u) => u.value === usage) ?? CA_USAGES[0];

/** What the certificate is for, as the user reads it */
export function caUsageLabel(
  t: ReturnType<typeof useTranslations<'CaCertificate'>>,
  usage: CaCertificateUsage,
): string {
  return t(caUsageOf(usage).label);
}

/** What a certificate of usage is for, as a sentence */
export function caUsageHelp(
  t: ReturnType<typeof useTranslations<'CaCertificate'>>,
  usage: CaCertificateUsage,
): string {
  return t(caUsageOf(usage).help);
}
