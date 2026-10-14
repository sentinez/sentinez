'use client';

import { useState } from 'react';

import { AccountResponse, Provider } from '@sentinez/proto/sentinez/apps/iam/v1/model';
import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Skeleton } from '@sentinez/ui/components/skeleton';
import { Fingerprint, KeyRound, LucideIcon } from 'lucide-react';
import { useFormatter, useTranslations } from 'next-intl';

import { useApi } from '@/hooks/use-api';
import { PasskeyLink } from '@/lib/api/iam/passkey';
import { ListAccounts } from '@/lib/api/iam/user';
import { handleStatusError } from '@/lib/error-handler';
import { toast } from '@/lib/toast';

/** The providers a user can sign in with, in display order */
const PROVIDERS: { provider: Provider; icon: LucideIcon }[] = [
  { provider: Provider.PROVIDER_LOCAL, icon: KeyRound },
  { provider: Provider.PROVIDER_PASSKEY, icon: Fingerprint },
];

/** Adds a passkey to the signed-in user, creating its passkey account if needed */
function LinkPasskeyButton({ onLinked }: Readonly<{ onLinked: () => void }>) {
  const t = useTranslations('Setting');
  const ta = useTranslations('Auth');
  const te = useTranslations('Errors');
  const [linking, setLinking] = useState(false);

  const link = async () => {
    setLinking(true);
    try {
      await PasskeyLink();
      toast.success(t('passkeyLinked'));
      onLinked();
    } catch (err) {
      // the browser prompt was dismissed
      if (err instanceof Error && err.name === 'NotAllowedError') {
        toast.error(ta('registerCancelled'), ta('registerCancelledDescription'));
      } else {
        handleStatusError(err, te);
      }
    } finally {
      setLinking(false);
    }
  };

  return (
    <Button size="sm" variant="outline" disabled={linking} onClick={link}>
      {linking ? t('linkingPasskey') : t('linkPasskey')}
    </Button>
  );
}

function AccountLine({ account }: Readonly<{ account: AccountResponse }>) {
  const t = useTranslations('Setting');
  const format = useFormatter();
  const createdAt = account.metadata?.createdAt;
  // a local account without a username stores its email there
  const showUsername = account.username && account.username !== account.email;

  return (
    <li className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 text-sm">
      <span className="min-w-0 break-all">
        {showUsername && <span className="font-medium">{account.username} · </span>}
        <span className={showUsername ? 'text-muted-foreground' : 'font-medium'}>
          {account.email}
        </span>
      </span>
      {createdAt && (
        <span className="text-muted-foreground text-xs">
          {t('linkedAt', { date: format.dateTime(createdAt, { dateStyle: 'medium' }) })}
        </span>
      )}
    </li>
  );
}

/** The accounts of the signed-in user, one row per provider */
export function AccountsCard({ userId }: Readonly<{ userId: string }>) {
  const t = useTranslations('Setting');
  const { data, isLoading, refresh } = useApi(ListAccounts, userId ? { userIds: [userId] } : null);
  const accounts = data?.accounts ?? [];

  const providerLabel = (provider: Provider) =>
    provider === Provider.PROVIDER_PASSKEY ? t('providerPasskey') : t('providerLocal');

  return (
    <Card>
      <CardHeader className="border-b">
        <CardTitle>{t('accountsTitle')}</CardTitle>
        <CardDescription>{t('accountsSubtitle')}</CardDescription>
      </CardHeader>
      <CardContent className="divide-y">
        {PROVIDERS.map(({ provider, icon: Icon }) => {
          const linked = accounts.filter((a) => a.provider === provider);
          return (
            <div key={provider} className="flex gap-4 py-4 first:pt-0 last:pb-0">
              <div className="bg-muted grid size-9 shrink-0 place-items-center rounded-md">
                <Icon className="size-4" aria-hidden />
              </div>
              <div className="grid min-w-0 flex-1 gap-2">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{providerLabel(provider)}</span>
                    {!isLoading && linked.length === 0 && (
                      <Badge variant="secondary">{t('notLinked')}</Badge>
                    )}
                  </div>
                  {provider === Provider.PROVIDER_PASSKEY && (
                    <LinkPasskeyButton onLinked={() => refresh()} />
                  )}
                </div>
                {isLoading ? (
                  <Skeleton className="h-4 w-48" />
                ) : (
                  linked.length > 0 && (
                    <ul className="grid gap-1">
                      {linked.map((account) => (
                        <AccountLine key={account.id} account={account} />
                      ))}
                    </ul>
                  )
                )}
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
