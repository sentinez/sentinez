import { Skeleton } from '@sentinez/ui/components/skeleton';
import { cn } from '@sentinez/ui/lib/utils';

export default function SidebarLoading({ className }: { className?: string }) {
  return <Skeleton className={cn('h-6 w-full', className)} />;
}
