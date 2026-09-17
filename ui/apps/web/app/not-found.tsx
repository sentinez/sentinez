import { Button } from '@sentinez/ui/components/button';
import { MoveLeft } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="flex flex-col gap-2 items-center justify-center h-screen">
      <div className="flex flex-col gap-5">
        <div className="flex gap-2 items-center">
          <Image src="/sntz.png" alt="sen" width={150} height={150} loading="eager" />
          <div className=" flex flex-col gap-2">
            <div className="font-medium text-7xl font-mono mx-2">404</div>
            <div className="font-medium text-2xl font-mono mx-2">Not Found!</div>
          </div>
        </div>
      </div>
    </div>
  );
}
