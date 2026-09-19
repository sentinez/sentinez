'use client';
import { Button } from '@sentinez/ui/components/button';
import { ChevronLeft } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { ReactNode } from 'react';

export interface TitleProps {
  children?: ReactNode;
  title: string;
  subtitle: string;
}

export default function Title({ children, title, subtitle }: TitleProps) {
  const router = useRouter();

  return (
    <div className="flex max-w-7xl p-2.25 sticky top-11.25 bg-background border-b">
      <div className="h-full items-center flex">
        <Button variant="ghost" size="icon" onClick={() => router.back()}>
          <ChevronLeft className="w-5 h-5" />
        </Button>
      </div>
      <div className="flex justify-between p-3 w-full">
        <div className="flex flex-col">
          <p className="text-lg font-semibold tracking-tight sm:text-lg">{title}</p>
          <p className="text-[1.05rem] text-muted-foreground sm:text-base sm:text-balance">
            {subtitle}
          </p>
        </div>

        <div className="flex items-center">{children}</div>
      </div>
    </div>
  );
}
