'use client';

import { Button } from '@sentinez/ui/components/button';
import { CheckIcon, CopyIcon } from 'lucide-react';

import { useCopy } from '@/hooks/use-copy';

/** An icon button copying `text`, its icon a check for a moment after */
export function CopyButton({ text, label }: { text: string; label: string }) {
  const { copied, copy } = useCopy();
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-sm"
      aria-label={label}
      title={label}
      onClick={() => copy(text)}
    >
      {copied ? <CheckIcon /> : <CopyIcon />}
    </Button>
  );
}
