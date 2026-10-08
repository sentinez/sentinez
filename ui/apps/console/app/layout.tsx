import { Roboto, Roboto_Mono } from 'next/font/google';

import '@sentinez/ui/globals.css';
import '@sentinez/ui/custom.css';

import { Providers } from '@/components/providers';
import { Metadata } from 'next';
import { NextIntlClientProvider } from 'next-intl';
import { getLocale, getTranslations } from 'next-intl/server';
import LoadingEffect from '@/components/loading-effect';

const fontSans = Roboto({
  subsets: ['latin', 'vietnamese'],
  variable: '--font-sans',
});

const fontMono = Roboto_Mono({
  subsets: ['latin', 'vietnamese'],
  variable: '--font-mono',
});

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('Metadata');
  const title = t('title');
  const description = t('description');

  return {
    title,
    description,

    openGraph: {
      title,
      description,
      url: 'https://s6z.io.vn/',
      siteName: 'Sentinéz',
      type: 'website',
      images: [
        {
          url: 'https://s6z.io.vn/images/sntz.png',
          width: 720,
          height: 720,
          alt: title,
        },
      ],
    },

    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: ['https://s6z.io.vn/images/sntz.png'],
    },
  };
}

export default async function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const locale = await getLocale();

  return (
    <html lang={locale} suppressHydrationWarning>
      <body className={`${fontSans.variable} ${fontMono.variable} font-sans antialiased `}>
        <NextIntlClientProvider>
          <Providers>
            <LoadingEffect>{children}</LoadingEffect>
          </Providers>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
