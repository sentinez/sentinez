'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';

import { toast } from '@/lib/toast';

// How long `copied` stays true after a copy
const COPIED_MILLIS = 1500;

/**
 * Copies text to the clipboard: `copied` is true for a moment after, for a
 * check mark; a failure (no permission, insecure context) is toasted.
 */
export function useCopy() {
  const t = useTranslations('Common');
  const [copied, setCopied] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => () => clearTimeout(timer.current), []);

  const copy = useCallback(
    async (text: string) => {
      try {
        await navigator.clipboard.writeText(text);
        setCopied(true);
        clearTimeout(timer.current);
        timer.current = setTimeout(() => setCopied(false), COPIED_MILLIS);
      } catch {
        toast.error(t('copyFailed'));
      }
    },
    [t],
  );

  return { copied, copy };
}
