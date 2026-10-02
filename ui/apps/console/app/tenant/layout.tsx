import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Tenant | Sentinez',
  description: 'Sentinez Tenant',
};

export default function Layout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return <>{children}</>;
}
