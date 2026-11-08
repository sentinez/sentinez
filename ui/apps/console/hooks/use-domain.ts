'use client';

import { useParams } from 'next/navigation';

/** The `[domain]` route segment, as in the URL (encoded); empty outside one */
export function useDomain(): string {
  const { domain = '' } = useParams<{ domain?: string }>();
  return domain;
}
