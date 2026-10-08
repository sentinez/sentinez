'use server';

import { cookies, headers } from 'next/headers';
import { LOCALE_COOKIE, defaultLocale, isLocale, locales, type Locale } from './config';

const ONE_YEAR = 60 * 60 * 24 * 365;

/** Cookie first, then the browser's Accept-Language, then the default */
export async function getUserLocale(): Promise<Locale> {
  const cookie = (await cookies()).get(LOCALE_COOKIE)?.value;
  if (isLocale(cookie)) return cookie;

  const accepted = (await headers()).get('accept-language') ?? '';
  for (const part of accepted.split(',')) {
    const lang = part.split(';')[0]?.trim().toLowerCase().split('-')[0];
    const match = locales.find((l) => l === lang);
    if (match) return match;
  }
  return defaultLocale;
}

export async function setUserLocale(locale: Locale) {
  if (!isLocale(locale)) return;
  (await cookies()).set(LOCALE_COOKIE, locale, { path: '/', maxAge: ONE_YEAR, sameSite: 'lax' });
}
