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
    <div className="flex w-full sticky top-[45px] bg-background border-b">
      <div className="h-full items-center flex">
        <Button variant="ghost" size="icon" onClick={() => router.back()}>
          <ChevronLeft className="w-5 h-5" />
        </Button>
      </div>
      <div className="flex justify-between flex-col w-full p-3">
        <div className="flex justify-between">
          <p className="text-lg font-semibold tracking-tight sm:text-2xl">{title}</p>
          {children}
        </div>

        <p className="text-[1.05rem] text-muted-foreground sm:text-base sm:text-balance md:max-w-[80%]">
          {subtitle}
        </p>
      </div>
    </div>
  );
}
