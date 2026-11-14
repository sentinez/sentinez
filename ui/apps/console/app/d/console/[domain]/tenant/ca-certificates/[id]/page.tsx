'use client';

import { type ReactNode, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useFormatter, useTranslations } from 'next-intl';
import { Copy, Download, FileKey2, KeyRound, ShieldCheck } from 'lucide-react';
import { CaCertificate, CaCertificateUsage } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import IsLoading from '@sentinez/ui/components/common/loading';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@sentinez/ui/components/dialog';
import { Progress } from '@sentinez/ui/components/progress';
import { cn } from '@sentinez/ui/lib/utils';
import { toast } from '@/lib/toast';
import { handleStatusError } from '@/lib/error-handler';
import { deleteCaCertificate, getCaCertificate } from '@/lib/api/tenant';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { InfoCard, InfoRow } from '@/components/info-card';
import { CaExpiryBadge, DAY_MS, caExpiryOf, caUsageHelp, caUsageLabel } from '../components';

/** Saves text as a file, in the browser */
function download(name: string, text: string) {
  const url = URL.createObjectURL(new Blob([text], { type: 'application/x-pem-file' }));
  const link = document.createElement('a');
  link.href = url;
  link.download = name;
  link.click();
  URL.revokeObjectURL(url);
}

export default function CaCertificatePage({ params }: { params: Promise<{ id: string }> }) {
  const t = useTranslations('CaCertificate');
  const tc = useTranslations('Common');
  const te = useTranslations('Errors');
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [cert, setCert] = useState<CaCertificate | null>(null);
  // when the certificate was read, for the validity left
  const [now, setNow] = useState(0);
  const [confirming, setConfirming] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    async function load() {
      try {
        const { id } = await params;
        const { caCertificate } = await getCaCertificate({ id });
        setCert(caCertificate ?? null);
        setNow(Date.now());
      } catch {
        toast.error(t('loadDetailFailed'));
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [params, t]);

  const copy = async (text: string) => {
    await navigator.clipboard.writeText(text);
    toast.success(t('copied'));
  };

  const handleDelete = async () => {
    if (!cert) return;
    setDeleting(true);
    try {
      await deleteCaCertificate({ id: cert.id });
      toast.success(t('deleted'));
      router.back();
    } catch (err) {
      handleStatusError(err, te);
    } finally {
      setDeleting(false);
    }
  };

  if (loading) return <IsLoading />;

  return (
    <PageLayout>
      <PageLayoutHeader title={cert?.name || t('detailTitle')} subtitle={t('immutableNote')}>
        {cert && (
          <Button variant="destructive" onClick={() => setConfirming(true)}>
            {tc('delete')}
          </Button>
        )}
      </PageLayoutHeader>

      <PageLayoutContent>
        {!cert && (
          <p className="py-10 text-center text-muted-foreground">{t('loadDetailFailed')}</p>
        )}
        {cert && (
          <>
            <SummaryCard cert={cert} now={now} />

            <div className="grid gap-6 lg:grid-cols-2">
              <InfoCard title={t('identity')}>
                <InfoRow label={t('subject')} mono>
                  {cert.subject}
                </InfoRow>
                <InfoRow label={t('issuer')} mono>
                  {cert.issuer}
                </InfoRow>
                <InfoRow label={t('provider')}>{cert.provider || '—'}</InfoRow>
                <InfoRow label={t('serialNumber')} mono>
                  {cert.serialNumber}
                </InfoRow>
                <InfoRow label={t('fingerprint')} mono>
                  <CopyValue
                    value={cert.fingerprintSha256}
                    label={t('copyFingerprint')}
                    onCopy={copy}
                  />
                </InfoRow>
              </InfoCard>

              <InfoCard title={t('details')}>
                <InfoRow label={t('id')} mono>
                  <CopyValue value={cert.id} label={t('copyId')} onCopy={copy} />
                </InfoRow>
                <InfoRow label={t('description')}>{cert.description || '—'}</InfoRow>
                <InfoRow label={t('usage')}>
                  <span className="font-medium">{caUsageLabel(t, cert.usage)}</span>
                  <span className="mt-1 block text-muted-foreground">
                    {caUsageHelp(t, cert.usage)}
                  </span>
                </InfoRow>
                <InfoRow label={t('privateKey')}>
                  {cert.hasPrivateKey ? t('keyStoredHelp') : t('none')}
                </InfoRow>
                <InfoRow label={t('selfSigned')}>{cert.selfSigned ? t('yes') : t('no')}</InfoRow>
              </InfoCard>
            </div>

            <Card>
              <CardHeader className="border-b">
                <CardTitle>{t('pem')}</CardTitle>
                <CardDescription>{t('pemDetailHelp')}</CardDescription>
                <CardAction className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={() => copy(cert.pem)}>
                    <Copy className="h-4 w-4" />
                    {t('copyPem')}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => download(`${cert.name || cert.id}.pem`, cert.pem)}
                  >
                    <Download className="h-4 w-4" />
                    {t('downloadPem')}
                  </Button>
                </CardAction>
              </CardHeader>
              <CardContent>
                <pre className="max-h-80 overflow-auto rounded-md bg-muted p-4 font-mono text-xs leading-relaxed">
                  {cert.pem}
                </pre>
              </CardContent>
            </Card>
          </>
        )}
      </PageLayoutContent>

      <Dialog open={confirming} onOpenChange={setConfirming}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('deleteTitle')}</DialogTitle>
            <DialogDescription>{tc('deleteConfirm', { name: cert?.name ?? '' })}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setConfirming(false)}>
              {tc('cancel')}
            </Button>
            <Button variant="destructive" disabled={deleting} onClick={handleDelete}>
              {deleting ? tc('deleting') : tc('delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  );
}

/** The name, the kind of certificate, and how much of its validity is left */
function SummaryCard({ cert, now }: { cert: CaCertificate; now: number }) {
  const t = useTranslations('CaCertificate');
  const server = cert.usage === CaCertificateUsage.CA_CERTIFICATE_USAGE_SERVER;
  const Icon = server ? FileKey2 : ShieldCheck;

  return (
    <Card className="py-5">
      <CardContent className="grid gap-6 px-5 md:grid-cols-[minmax(0,1fr)_minmax(0,22rem)] md:items-center">
        <div className="flex min-w-0 items-start gap-4">
          <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <Icon className="h-5 w-5" />
          </div>
          <div className="min-w-0 space-y-2">
            <div>
              <h2 className="truncate text-lg font-semibold">{cert.name}</h2>
              <p className="truncate text-sm text-muted-foreground">
                {cert.provider || cert.issuer}
              </p>
            </div>
            <div className="flex flex-wrap gap-1.5">
              <Badge variant="outline">{caUsageLabel(t, cert.usage)}</Badge>
              {cert.selfSigned && <Badge variant="outline">{t('rootCa')}</Badge>}
              {cert.hasPrivateKey && (
                <Badge variant="outline">
                  <KeyRound className="h-3 w-3" />
                  {t('privateKey')}
                </Badge>
              )}
              <CaExpiryBadge notAfter={cert.notAfter} now={now} />
            </div>
          </div>
        </div>

        <Validity notBefore={cert.notBefore} notAfter={cert.notAfter} now={now} />
      </CardContent>
    </Card>
  );
}

/** The validity period as a bar of the time elapsed, with the days left */
function Validity({
  notBefore,
  notAfter,
  now,
}: {
  notBefore?: Date;
  notAfter?: Date;
  now: number;
}) {
  const t = useTranslations('CaCertificate');
  const format = useFormatter();
  if (!notBefore || !notAfter) return null;

  const start = notBefore.getTime();
  const end = notAfter.getTime();
  const elapsed = end > start ? ((now - start) / (end - start)) * 100 : 100;
  const daysLeft = Math.floor((end - now) / DAY_MS);
  const date = (d: Date) => format.dateTime(d, { dateStyle: 'medium' });

  const expiry = caExpiryOf(notAfter, now);
  const expired = expiry === 'expired';
  const soon = expiry === 'soon';

  return (
    <div className="grid gap-2">
      <div className="flex items-baseline justify-between gap-2 text-sm">
        <span className="text-muted-foreground">{t('validity')}</span>
        <span
          className={cn(
            'font-medium',
            expired && 'text-destructive',
            soon && 'text-amber-600 dark:text-amber-500',
          )}
        >
          {expired
            ? t('expiredDaysAgo', { count: Math.max(0, -daysLeft) })
            : t('daysLeft', { count: daysLeft })}
        </span>
      </div>
      <Progress
        value={Math.min(100, Math.max(0, elapsed))}
        aria-label={t('validity')}
        className={cn(
          expired && '[&>[data-slot=progress-indicator]]:bg-destructive',
          soon && '[&>[data-slot=progress-indicator]]:bg-amber-500',
        )}
      />
      <div className="flex justify-between gap-2 text-xs text-muted-foreground">
        <span>{date(notBefore)}</span>
        <span>{date(notAfter)}</span>
      </div>
    </div>
  );
}

/** A long value, such as an id, with a button that copies it */
function CopyValue({
  value,
  label,
  onCopy,
}: {
  value: string;
  label: string;
  onCopy: (value: string) => void;
}): ReactNode {
  return (
    <span className="flex items-start gap-2">
      <span className="min-w-0 break-all">{value}</span>
      <Button
        variant="ghost"
        size="icon"
        className="h-6 w-6 shrink-0"
        aria-label={label}
        onClick={() => onCopy(value)}
      >
        <Copy className="h-3.5 w-3.5" />
      </Button>
    </span>
  );
}
