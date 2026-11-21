'use client';

import { ChevronDown, Languages, LoaderCircle } from 'lucide-react';
import { useLocale, useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { useOptimistic, useTransition } from 'react';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';
import { Button } from '@sentinez/ui/components/button';
import { LOCALE_LABEL, isLocale, locales } from '@/i18n/config';
import { setUserLocale } from '@/i18n/locale';

export default function Language() {
  const t = useTranslations('Language');
  const current = useLocale();
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  // Shows the picked language at once; it falls back to the real locale if
  // the change fails.
  const [locale, setLocale] = useOptimistic(current);

  const label = LOCALE_LABEL[locale];

  const handleChange = (value: string) => {
    if (!isLocale(value) || value === locale) return;
    startTransition(async () => {
      setLocale(value);
      await setUserLocale(value);
      router.refresh();
    });
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="gap-1.5 text-xs text-muted-foreground"
          aria-label={`${t('label')}: ${label}`}
          aria-busy={pending}
          disabled={pending}
        >
          {pending ? <LoaderCircle className="animate-spin" /> : <Languages />}
          <span className="hidden sm:inline">{label}</span>
          <span className="uppercase sm:hidden">{locale}</span>
          <ChevronDown className="size-3 opacity-60" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-40">
        <DropdownMenuLabel className="text-xs text-muted-foreground">{t('label')}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuRadioGroup value={locale} onValueChange={handleChange}>
          {locales.map((l) => (
            <DropdownMenuRadioItem key={l} value={l} lang={l}>
              {LOCALE_LABEL[l]}
              <span className="ml-auto pl-3 text-xs uppercase text-muted-foreground">{l}</span>
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
