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
import { toast } from '@/lib/toast';
import { RequestLink } from './request-link';
import type { DetailSection, EventConfig } from './types';

type View = 'table' | 'json';

function Row({ name, children }: { name: string; children: ReactNode }) {
  return (
    <div className="grid grid-cols-[6.5rem_minmax(0,1fr)] gap-3 border-b py-1.5 text-xs last:border-0 sm:grid-cols-[9rem_minmax(0,1fr)]">
      <span className="break-words font-medium text-muted-foreground">{name}</span>
      <span className="min-w-0 break-all font-mono">{children || '—'}</span>
    </div>
  );
}

function Section({ title, pairs }: DetailSection) {
  if (pairs.length === 0) return null;
  return (
    <div className="flex flex-col">
      <p className="pb-1 pt-4 text-sm font-semibold">{title}</p>
      {pairs.map(([k, v], i) => (
        <Row key={`${k}-${i}`} name={k}>
          {v}
        </Row>
      ))}
    </div>
  );
}

interface Props<T> {
  config: EventConfig<T>;
  event: T | null;
  onClose: () => void;
}

/** The fields of an event, as a table or JSON (Kibana expanded document) */
export function EventDetail<T>({ config, event, onClose }: Props<T>) {
  const t = useTranslations('Explorer');
  const format = useFormatter();
  const [view, setView] = useState<View>('table');
  const [copied, setCopied] = useState(false);
  const { detail } = config;
  const json = useMemo(
    () => (event ? JSON.stringify(detail.toJSON(event), null, 2) : ''),
    [event, detail],
  );
  const request = event ? config.requestOf?.(event) : undefined;

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(json);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      toast.error(t('copyFailed'));
    }
  };

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
              <Button size="sm" variant="ghost" className="ml-auto" onClick={copy}>
                {copied ? <Check /> : <Copy />}
                {copied ? t('detail.copied') : t('detail.copy')}
              </Button>
            )}
          </div>
        </SheetHeader>

        {event && (
          <div className="flex-1 overflow-y-auto px-4 pb-6">
            {view === 'json' ? (
              <pre className="mt-4 overflow-x-auto rounded-md bg-muted p-3 font-mono text-xs">
                {json}
              </pre>
            ) : (
              <>
                <div className="flex flex-col pt-2">
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
                </div>
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
