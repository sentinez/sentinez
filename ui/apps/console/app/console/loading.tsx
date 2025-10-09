import Image from 'next/image';
import { LoaderCircle } from 'lucide-react';

export default function Page() {
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="flex flex-col gap-2">
        <Image src="/assets/sntz.png" alt="sen" width={100} height={100} />
        <div className="font-medium text-xl font-mono mx-2">The Sentinez WAF is loading.</div>
        <div className=" mx-2">
          <LoaderCircle className="animate-spin text-slate-500 font-bold" size={30} />
        </div>
      </div>
    </div>
  );
}
