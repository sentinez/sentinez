import { Geist, Geist_Mono } from 'next/font/google';
import '@sentinez/ui/globals.css';
import { Providers } from '@/components/providers';
const fontSans = Geist({
    subsets: ['latin'],
    variable: '--font-sans',
});
const fontMono = Geist_Mono({
    subsets: ['latin'],
    variable: '--font-mono',
});
export const metadata = {
    title: 'Sentinez WAF',
    description: 'Sentinez Web Application Firewall',
};
export default function RootLayout({ children, }) {
    return (<html lang="en" suppressHydrationWarning>
      <body className={`${fontSans.variable} ${fontMono.variable} font-sans antialiased `}>
        <Providers>{children}</Providers>
      </body>
    </html>);
}
