export const locales = ['en', 'vi'] as const;

export type Locale = (typeof locales)[number];

export const defaultLocale: Locale = 'en';

/** Cookie that keeps the locale picked in the language switcher */
export const LOCALE_COOKIE = 'NEXT_LOCALE';

/** Native name of each locale, shown in the language switcher */
export const LOCALE_LABEL: Record<Locale, string> = {
  en: 'English',
  vi: 'Tiếng Việt',
};

export function isLocale(value: string | undefined): value is Locale {
  return locales.includes(value as Locale);
}
