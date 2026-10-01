import LoadingEffect from '@/components/loading-effect';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Authentication | Sentinez',
  description: 'Sentinez Central Authentication',
};

export default function Layout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <LoadingEffect>{children}</LoadingEffect>;
}
