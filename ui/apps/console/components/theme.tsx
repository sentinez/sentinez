'use client';

import { Moon, Sun } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useTheme } from 'next-themes';

import { Button } from '@sentinez/ui/components/button';

// Theme switches between light and dark (next-themes, which puts the .dark
// class on <html>, see Providers); never the theme of the OS. The icon
// follows the .dark class through CSS rather than useTheme, which is unknown
// on the server: no hydration mismatch.
export default function Theme() {
  const t = useTranslations('Theme');
  const { resolvedTheme, setTheme } = useTheme();

  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-sm"
      className="text-muted-foreground"
      aria-label={t('toggle')}
      title={t('toggle')}
      onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
    >
      <Sun className="dark:hidden" />
      <Moon className="hidden dark:block" />
    </Button>
  );
}
