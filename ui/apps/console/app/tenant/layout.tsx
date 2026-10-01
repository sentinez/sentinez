import LoadingEffect from '@/components/loading-effect';
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
  return <LoadingEffect>{children}</LoadingEffect>;
}
