'use client';

import { useId } from 'react';
import { ArrowUpRightIcon, ShieldCheckIcon } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { Button } from '@sentinez/ui/components/button';
import {
  Card,
  CardAction,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';

const CRS_URL = 'https://coreruleset.org';

/**
 * Card telling that the ruleset runs the OWASP Core Rule Set, above its
 * create / update form. The link sits beside the text on a wide card and
 * drops below it on a narrow one.
 */
export function CrsNotice() {
  const t = useTranslations('CrsNotice');
  const titleId = useId();

  return (
    <Card role="region" aria-labelledby={titleId} className="min-w-0 gap-0 py-3 shadow-none">
      <CardHeader className="gap-x-3 gap-y-0.5 px-4 has-data-[slot=card-action]:grid-cols-1 @md/card-header:has-data-[slot=card-action]:grid-cols-[1fr_auto]">
        <CardTitle
          id={titleId}
          className="flex min-w-0 items-start gap-2 text-sm leading-5 font-medium"
        >
          <ShieldCheckIcon className="mt-0.5 size-4 shrink-0 text-primary" aria-hidden />
          <span className="min-w-0 break-words">{t('title')}</span>
        </CardTitle>
        <CardDescription className="pl-6 text-xs leading-4 break-words">
          {t('inUse')}
        </CardDescription>
        <CardAction className="col-start-1 row-span-1 row-start-3 mt-2 ml-6 max-w-[calc(100%-1.5rem)] justify-self-start @md/card-header:col-start-2 @md/card-header:row-span-2 @md/card-header:row-start-1 @md/card-header:mt-0 @md/card-header:ml-0 @md/card-header:max-w-48 @md/card-header:self-center @md/card-header:justify-self-end">
          <Button asChild variant="outline" size="xs" className="max-w-full">
            <a href={CRS_URL} target="_blank" rel="noopener noreferrer">
              <span className="truncate">{t('learnMore')}</span>
              <ArrowUpRightIcon aria-hidden />
            </a>
          </Button>
        </CardAction>
      </CardHeader>
    </Card>
  );
}
