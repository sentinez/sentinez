'use client';

import { useMemo } from 'react';
import { useTranslations } from 'next-intl';

import type { FieldChange } from '@sentinez/proto/sentinez/types/audit/v1/audit';
import { cn } from '@sentinez/ui/lib/utils';
import { collapse, lineDiff } from '@/lib/line-diff';

// The text of a protojson value, undefined when unset
const textOf = (value: unknown): string | undefined =>
  value === undefined || value === null
    ? undefined
    : typeof value === 'string'
      ? value
      : JSON.stringify(value);

/** The value of a field before and after an update: a line diff for a text of several lines */
export function FieldChangeValue({ change }: { change: FieldChange }) {
  const before = textOf(change.before);
  const after = textOf(change.after);
  if (before?.includes('\n') || after?.includes('\n')) {
    return <TextDiff before={before ?? ''} after={after ?? ''} />;
  }
  return (
    <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
      <Value text={before} className="bg-red-500/10 text-red-700 line-through dark:text-red-400" />
      <span className="text-muted-foreground">→</span>
      <Value text={after} className="bg-green-500/10 text-green-700 dark:text-green-400" />
    </div>
  );
}

// A value, "unset" muted when there is none
function Value({ text, className }: { text?: string; className: string }) {
  const t = useTranslations('AuditLog');
  if (text === undefined) return <span className="text-muted-foreground italic">{t('unset')}</span>;
  return <span className={cn('break-all rounded px-1', className)}>{text}</span>;
}

const LINE_CLASSES = {
  same: 'text-muted-foreground',
  removed: 'bg-red-500/10 text-red-700 dark:text-red-400',
  added: 'bg-green-500/10 text-green-700 dark:text-green-400',
};

const LINE_SIGNS = { same: ' ', removed: '-', added: '+' };

// The lines of two texts, like git diff, unchanged runs collapsed
function TextDiff({ before, after }: { before: string; after: string }) {
  const t = useTranslations('AuditLog');
  const hunks = useMemo(
    () => collapse(lineDiff(before.split('\n'), after.split('\n'))),
    [before, after],
  );
  return (
    <pre className="max-h-80 overflow-auto rounded border text-[11px] leading-5">
      {hunks.map((h, i) =>
        h.kind === 'skipped' ? (
          <div key={i} className="bg-muted/50 text-muted-foreground px-2 italic">
            {t('unchangedLines', { count: h.count })}
          </div>
        ) : (
          <div key={i} className={cn('whitespace-pre-wrap break-all px-2', LINE_CLASSES[h.kind])}>
            {LINE_SIGNS[h.kind]} {h.text}
          </div>
        ),
      )}
    </pre>
  );
}
