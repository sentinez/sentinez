import { LoaderCircle } from 'lucide-react';

export default function IsLoading() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center text-center px-4">
      <LoaderCircle className="animate-spin text-slate-500 font-bold" size={30} />
    </div>
  );
}
