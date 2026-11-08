'use client';

import type { ComponentProps } from 'react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { ArrowUpRight } from 'lucide-react';

import { cn } from '@sentinez/ui/lib/utils';
import { useDomain } from '@/hooks/use-domain';
import type { RequestRef } from './types';

// The query params of the ingress page opening a request (see the focus
// of EventExplorer)
export const REQUEST_PARAM = 'request';
export const AT_PARAM = 'at';

// The ingress page of the domain, opened on the request of ref
export function requestHref(domain: string, ref: RequestRef): string {
  const params = new URLSearchParams({ [REQUEST_PARAM]: ref.id, [AT_PARAM]: String(ref.time) });
  return `/d/console/${domain}/analytic/logs/ingress?${params}`;
}

/**
 * A link to the request an event is about, on the ingress page: its id, or
 * an icon only in a table; a click does not open the row it is in.
 */
export function RequestLink({
  request,
  iconOnly = false,
}: {
  request: RequestRef;
  iconOnly?: boolean;
}) {
  const t = useTranslations('Explorer');
  const domain = useDomain();
  return (
    <RowLink
      href={requestHref(domain, request)}
      title={t('viewRequest')}
      aria-label={iconOnly ? t('viewRequest') : undefined}
      className="inline-flex min-w-0 items-center gap-1"
    >
      {!iconOnly && <span className="truncate">{request.id}</span>}
      <ArrowUpRight className="size-3.5 shrink-0" />
    </RowLink>
  );
}

/** A link in a row of the table: a click on it does not open the row */
export function RowLink({ className, ...props }: ComponentProps<typeof Link>) {
  return (
    <Link
      {...props}
      className={cn('font-mono text-xs text-blue-700 underline dark:text-blue-400', className)}
      onClick={(e) => e.stopPropagation()}
    />
  );
}
