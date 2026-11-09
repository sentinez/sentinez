'use client';

import { useTranslations } from 'next-intl';

/** The resource of a domain could not be loaded: the error, or that it is gone */
export function ResourceUnavailable({ domain, error }: { domain: string; error?: Error | null }) {
  const t = useTranslations('Resource');

  return (
    <div className="p-8 flex flex-col items-center justify-center text-center space-y-4">
      <div className="text-muted-foreground">{error?.message || t('unavailable')}</div>
      <p className="text-sm text-muted-foreground max-w-xs">
        {t.rich('unavailableHint', {
          domain,
          strong: (chunks) => <strong>{chunks}</strong>,
        })}
      </p>
    </div>
  );
}
