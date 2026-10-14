'use client';

import { useEffect, useState } from 'react';
import { Button } from '@sentinez/ui/components/button';
import { cn } from '@sentinez/ui/lib/utils';
import { RefreshCw } from 'lucide-react';
import { useTranslations } from 'next-intl';

// A click keeps the icon spinning at least this long, so a fast fetch still shows
const MIN_SPIN_MILLIS = 500;

interface Props {
  onRefresh: () => void;
  /** The list is being fetched: the icon spins and the button is disabled */
  loading?: boolean;
  className?: string;
}

/** Icon button next to a list's search box that fetches the list again */
export function RefreshButton({ onRefresh, loading = false, className }: Props) {
  const t = useTranslations('Common');
  const [holding, setHolding] = useState(false);
  const busy = loading || holding;

  useEffect(() => {
    if (!holding) return;
    const timer = setTimeout(() => setHolding(false), MIN_SPIN_MILLIS);
    return () => clearTimeout(timer);
  }, [holding]);

  const handleClick = () => {
    setHolding(true);
    onRefresh();
  };

  return (
    <Button
      type="button"
      variant="outline"
      size="icon"
      aria-label={t('refresh')}
      title={t('refresh')}
      disabled={busy}
      className={className}
      onClick={handleClick}
    >
      <RefreshCw className={cn(busy && 'animate-spin')} />
    </Button>
  );
}
