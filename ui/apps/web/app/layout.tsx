import { Geist, Geist_Mono } from 'next/font/google';

import '@sentinez/ui/globals.css';
import { Providers } from '@/components/providers';
import { Metadata } from 'next';

const fontSans = Geist({
  subsets: ['latin'],
  variable: '--font-sans',
});

const fontMono = Geist_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
});

export const metadata: Metadata = {
  title: 'Sentinéz — On Your Side',
  description: 'Security and protection for modern applications.',

  openGraph: {
    title: 'Sentinéz — On Your Side',
    description: 'Security and protection for modern applications.',
    url: 'https://s6z.io.vn/',
    siteName: 'Sentinéz',
    type: 'website',
    images: [
      {
        url: 'https://s6z.io.vn/images/sntz.png',
        width: 720,
        height: 720,
        alt: 'Sentinéz — On Your Side',
      },
    ],
  },

  twitter: {
    card: 'summary_large_image',
    title: 'Sentinéz — On Your Side',
    description: 'Security and protection for modern applications.',
    images: ['https://s6z.io.vn/images/sntz.png'],
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${fontSans.variable} ${fontMono.variable} font-sans antialiased `}>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
