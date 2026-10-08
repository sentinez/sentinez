'use client';

import { Languages } from 'lucide-react';
import { useLocale, useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { useTransition } from 'react';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';
import { Button } from '@sentinez/ui/components/button';
import { LOCALE_LABEL, isLocale, locales } from '@/i18n/config';
import { setUserLocale } from '@/i18n/locale';

export default function Language() {
  const t = useTranslations('Language');
  const locale = useLocale();
  const router = useRouter();
  const [pending, startTransition] = useTransition();

  const handleChange = (value: string) => {
    if (!isLocale(value) || value === locale) return;
    startTransition(async () => {
      await setUserLocale(value);
      router.refresh();
    });
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="text-xs text-muted-foreground"
          size={'sm'}
          disabled={pending}
        >
          <Languages /> {LOCALE_LABEL[locale]}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-36">
        <DropdownMenuGroup>
          <DropdownMenuLabel>{t('label')}</DropdownMenuLabel>
          <DropdownMenuRadioGroup value={locale} onValueChange={handleChange}>
            {locales.map((l) => (
              <DropdownMenuRadioItem key={l} value={l}>
                {LOCALE_LABEL[l]}
              </DropdownMenuRadioItem>
            ))}
          </DropdownMenuRadioGroup>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
