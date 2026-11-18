'use client';

import type { ReactNode } from 'react';
import { X } from 'lucide-react';

import { Badge } from '@sentinez/ui/components/badge';
import { cn } from '@sentinez/ui/lib/utils';

interface Props {
  children: ReactNode;
  /** The label of the remove button */
  removeLabel: string;
  onRemove: () => void;
  className?: string;
  title?: string;
}

/** A badge of a filter, its text truncated, with a button removing it */
export function RemovableBadge({ children, removeLabel, onRemove, className, title }: Props) {
  return (
    <Badge
      variant="secondary"
      className={cn('max-w-full gap-1 font-normal', className)}
      title={title}
    >
      <span className="truncate">{children}</span>
      <button
        type="button"
        aria-label={removeLabel}
        className="shrink-0 rounded-full hover:text-foreground"
        onClick={onRemove}
      >
        <X />
      </button>
    </Badge>
  );
}
