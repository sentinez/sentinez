import type { ReactNode } from 'react';
import { cn } from '@sentinez/ui/lib/utils';

// literal classes, so tailwind sees them
const COLUMNS = {
  2: 'sm:grid-cols-2',
  3: 'sm:grid-cols-3',
} as const;

interface Props {
  children: ReactNode;
  /** Fields per row from `sm` up; one per row below. */
  cols?: keyof typeof COLUMNS;
  className?: string;
}

/**
 * A row of short form fields (selects, numbers) side by side, so they do not
 * stretch across the whole page width; stacked on narrow screens.
 */
export function FieldGrid({ children, cols = 2, className }: Props) {
  return <div className={cn('grid grid-cols-1 gap-4', COLUMNS[cols], className)}>{children}</div>;
}
