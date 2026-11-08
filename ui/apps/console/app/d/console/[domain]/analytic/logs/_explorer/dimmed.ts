import { cn } from '@sentinez/ui/lib/utils';

// The classes of a block refetching: its last data stays, dimmed and inert
export const dimmed = (loading: boolean) =>
  cn('transition-opacity duration-200', loading && 'pointer-events-none opacity-50');
