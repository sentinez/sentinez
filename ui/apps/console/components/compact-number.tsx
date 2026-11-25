'use client';

import { useFormatter } from 'next-intl';
import { compactNumber } from '@/lib/number';

// CompactNumber a count shown short (1.2K, 3.4M), the whole number on hover
export function CompactNumber({ value, className }: { value: number; className?: string }) {
  const format = useFormatter();
  return (
    <span className={className} title={format.number(value)}>
      {compactNumber(value)}
    </span>
  );
}
