'use client';

import { useMemo, useState, type ReactNode } from 'react';
import { useFormatter, useTranslations } from 'next-intl';
import { Check, Copy } from 'lucide-react';

import { Button } from '@sentinez/ui/components/button';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@sentinez/ui/components/sheet';
import { useCopy } from '@/hooks/use-copy';
import { InfoRow } from '@/components/info-card';
import { CodeBlock } from '@/components/code-block';
import { RequestLink } from './request-link';
import type { DetailConfig, DetailSection } from './types';

type View = 'table' | 'json';

// A field of the event: dense, its value in monospace
function Row({ name, children }: { name: string; children: ReactNode }) {
  return (
    <InfoRow label={name} mono compact>
      {children || '—'}
    </InfoRow>
  );
}

function Section({ title, pairs }: DetailSection) {
  if (pairs.length === 0) return null;
  return (
    <div className="flex flex-col">
      <p className="pb-1 pt-4 text-sm font-semibold">{title}</p>
      <dl className="flex flex-col">
        {pairs.map(([k, v], i) => (
          <Row key={`${k}-${i}`} name={k}>
            {v}
          </Row>
        ))}
      </dl>
    </div>
  );
}

interface Props<T> {
  config: DetailConfig<T>;
  event: T | null;
  onClose: () => void;
}

/** The fields of an event, as a table or JSON (Kibana expanded document) */
export function EventDetail<T>({ config, event, onClose }: Props<T>) {
  const t = useTranslations('Explorer');
  const format = useFormatter();
  const [view, setView] = useState<View>('table');
  const { detail } = config;
  const json = useMemo(
    () => (event ? JSON.stringify(detail.toJSON(event), null, 2) : ''),
    [event, detail],
  );
  const request = event ? config.requestOf?.(event) : undefined;

  const { copied, copy } = useCopy();

  return (
    <Sheet open={!!event} onOpenChange={(open) => !open && onClose()}>
      <SheetContent className="w-full gap-0 sm:max-w-2xl">
        <SheetHeader className="border-b">
          <SheetTitle className="flex min-w-0 items-center gap-2 pr-6">
            {event && detail.title(event)}
          </SheetTitle>
          <SheetDescription className="font-mono text-xs">
            {event ? config.idOf(event) : ''}
          </SheetDescription>
          <div className="flex flex-wrap items-center gap-1 pt-2">
            {(['table', 'json'] as const).map((v) => (
              <Button
                key={v}
                size="sm"
                variant={view === v ? 'secondary' : 'ghost'}
                aria-pressed={view === v}
                onClick={() => setView(v)}
              >
                {t(`detail.${v}`)}
              </Button>
            ))}
            {view === 'json' && (
              <Button size="sm" variant="ghost" className="ml-auto" onClick={() => copy(json)}>
                {copied ? <Check /> : <Copy />}
                {copied ? t('detail.copied') : t('detail.copy')}
              </Button>
            )}
          </div>
        </SheetHeader>

        {event && (
          <div className="flex-1 overflow-y-auto px-4 pb-6">
            {view === 'json' ? (
              <CodeBlock className="mt-4 max-h-none" value={json} />
            ) : (
              <>
                <dl className="flex flex-col pt-2">
                  <Row name={t('time')}>
                    {format.dateTime(new Date(config.timeOf(event)), {
                      dateStyle: 'medium',
                      timeStyle: 'long',
                    })}
                  </Row>
                  {detail.rows(event).map(([name, value]) => (
                    <Row key={name} name={name}>
                      {value}
                    </Row>
                  ))}
                  {request?.id && (
                    <Row name={t('requestId')}>
                      <RequestLink request={request} />
                    </Row>
                  )}
                </dl>
                {detail.sections?.(event).map((section) => (
                  <Section key={section.title} {...section} />
                ))}
              </>
            )}
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
