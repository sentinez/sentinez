'use client';

import { useTranslations } from 'next-intl';

import { RemovableBadge } from './removable-badge';
import type { FieldDef, FieldFilter } from './types';

interface Props<T> {
  fields: FieldDef<T>[];
  filters: FieldFilter[];
  onRemove: (index: number) => void;
}

/** The field filters of the page, Kibana-like pills, each removable */
export function FieldFilters<T>({ fields, filters, onRemove }: Props<T>) {
  const t = useTranslations('Explorer');
  if (filters.length === 0) return null;

  return (
    <div className="flex flex-wrap items-center gap-2">
      {filters.map(({ field, value }, i) => {
        const def = fields.find((f) => f.name === field);
        const label = `${def?.label ?? field}: ${def?.format?.(value) ?? value}`;
        return (
          <RemovableBadge
            key={`${field}=${value}`}
            className="font-mono"
            title={label}
            removeLabel={t('removeFilter', { filter: label })}
            onRemove={() => onRemove(i)}
          >
            {label}
          </RemovableBadge>
        );
      })}
    </div>
  );
}
