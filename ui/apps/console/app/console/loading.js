import Image from 'next/image';
import { LoaderCircle } from 'lucide-react';
export default function Loading() {
    return (<div className="fixed inset-0 flex items-center justify-center bg-background z-50">
      <div className="flex flex-col items-center gap-4">
        <Image src="/assets/sntz.png" alt="Sentinez" width={70} height={70}/>
        <div className="font-medium font-mono text-foreground">The Sentinez is loading.</div>
        <LoaderCircle className="animate-spin text-muted-foreground" size={28}/>
      </div>
    </div>);
}
