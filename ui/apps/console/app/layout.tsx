import { Roboto, Roboto_Mono } from 'next/font/google';

import '@sentinez/ui/globals.css';
import '@sentinez/ui/custom.css';

import { Providers } from '@/components/providers';
import { Metadata, Viewport } from 'next';
import { headers } from 'next/headers';
import { NextIntlClientProvider } from 'next-intl';
import { getLocale, getTranslations } from 'next-intl/server';
import LoadingEffect from '@/components/loading-effect';
import { SessionRedirect } from '@/components/session-redirect';

const fontSans = Roboto({
  subsets: ['latin', 'vietnamese'],
  variable: '--font-sans',
});

const fontMono = Roboto_Mono({
  subsets: ['latin', 'vietnamese'],
  variable: '--font-mono',
});

// Lay the page out on the width of the device, whatever its pixel density; the
// user can still pinch-zoom
export const viewport: Viewport = { width: 'device-width', initialScale: 1 };

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('Metadata');
  const title = t('title');
  const description = t('description');
  const image = '/assets/sz-og.png';

  return {
    metadataBase: await requestOrigin(),
    title,
    description,

    openGraph: {
      title,
      description,
      url: 'https://console.s6z.io.vn/',
      siteName: 'Sentinéz',
      type: 'website',
      images: [
        {
          url: image,
          width: 1200,
          height: 630,
          alt: title,
        },
      ],
    },

    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: [image],
    },
  };
}

// requestOrigin resolves the origin the console is served from, so the
// metadata image links back to this app's own public assets.
async function requestOrigin(): Promise<URL | null> {
  const h = await headers();
  const host = h.get('x-forwarded-host') ?? h.get('host');
  if (!host) return null;

  const proto = h.get('x-forwarded-proto') ?? 'https';
  return new URL(`${proto}://${host}`);
}

export default async function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const locale = await getLocale();

  return (
    <html lang={locale} suppressHydrationWarning>
      <body className={`${fontSans.variable} ${fontMono.variable} font-sans antialiased `}>
        <NextIntlClientProvider>
          <Providers>
            <SessionRedirect />
            <LoadingEffect>{children}</LoadingEffect>
          </Providers>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
