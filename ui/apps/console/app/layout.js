import { Roboto, Roboto_Mono } from 'next/font/google';
import '@sentinez/ui/globals.css';
import { Providers } from '@/components/providers';
const fontSans = Roboto({
    subsets: ['latin'],
    variable: '--font-sans',
});
const fontMono = Roboto_Mono({
    subsets: ['latin'],
    variable: '--font-mono',
});
export const metadata = {
    title: 'Sentinez',
    description: 'Sentinez',
};
export default function RootLayout({ children, }) {
    return (<html lang="en" suppressHydrationWarning>
      <body className={`${fontSans.variable} ${fontMono.variable} font-sans antialiased `}>
        <Providers>{children}</Providers>
      </body>
    </html>);
}
