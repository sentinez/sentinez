'use client';

import { useEffect, useState } from 'react';
import { Button } from '@sentinez/ui/components/button';
import { cn } from '@sentinez/ui/lib/utils';
import { RefreshCw } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { MIN_LOADING_MILLIS } from '@/lib/delays';

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
    const timer = setTimeout(() => setHolding(false), MIN_LOADING_MILLIS);
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
