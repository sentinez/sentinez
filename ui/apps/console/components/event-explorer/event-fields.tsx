'use client';

import { useFormatter, useTranslations } from 'next-intl';

import { cn } from '@sentinez/ui/lib/utils';
import { Skeleton } from '@sentinez/ui/components/skeleton';
import type { GetEventTopValuesResponse } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { dimmed } from './dimmed';
import type { FieldDef } from './types';

interface Props<T> {
  fields: FieldDef<T>[];
  /** The top values of the range, counted by the API; null until loaded */
  data: GetEventTopValuesResponse | null;
  /** Refetching: the last top values stay, dimmed */
  loading: boolean;
  /** Adds `field:value` to the filter */
  onFilter: (field: string, value: string) => void;
}

/**
 * Kibana-style field list: the most frequent values of each field among
 * the events of the time range, and their share, as the API counts them.
 */
export function EventFields<T>({ fields, data, loading, onFilter }: Props<T>) {
  const t = useTranslations('Explorer');
  const format = useFormatter();

  if (!data) return <Skeleton className="h-96 w-full" />;

  return (
    <aside className={cn('flex flex-col gap-4', dimmed(loading))} aria-busy={loading}>
      <div>
        <p className="text-sm font-semibold">{t('fields')}</p>
        <p className="text-xs text-muted-foreground">
          {t('fieldsDescription', { count: data.total })}
        </p>
      </div>
      {data.fields.map(({ field, values }) => {
        const def = fields.find((f) => f.name === field);
        if (!def) return null;
        return (
          <div key={field} className="flex flex-col gap-1">
            <p className="text-xs font-medium text-muted-foreground">{def.label}</p>
            {values.length === 0 && <p className="text-xs text-muted-foreground">—</p>}
            {values.map(({ value, share }) => {
              const label = def.format?.(value) ?? value;
              return (
                <button
                  key={value}
                  type="button"
                  title={t('filterFor', { field: def.label, value: label })}
                  className="relative flex items-center justify-between gap-2 overflow-hidden rounded px-2 py-1 text-left text-xs hover:bg-muted"
                  onClick={() => onFilter(field, value)}
                >
                  <span
                    aria-hidden
                    className="absolute inset-y-0 left-0 bg-primary/10 transition-[width] duration-300"
                    style={{ width: `${Math.min(share, 1) * 100}%` }}
                  />
                  <span className="relative truncate font-mono">{label}</span>
                  <span className="relative shrink-0 tabular-nums text-muted-foreground">
                    {format.number(share, { style: 'percent', maximumFractionDigits: 1 })}
                  </span>
                </button>
              );
            })}
          </div>
        );
      })}
    </aside>
  );
}
