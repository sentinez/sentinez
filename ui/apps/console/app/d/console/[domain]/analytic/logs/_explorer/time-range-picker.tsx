'use client';

import { useId, useState } from 'react';
import { useFormatter, useLocale, useTranslations } from 'next-intl';
import { CalendarIcon, ChevronDown } from 'lucide-react';

import { Button } from '@sentinez/ui/components/button';
import { Calendar } from '@sentinez/ui/components/calendar';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { Popover, PopoverContent, PopoverTrigger } from '@sentinez/ui/components/popover';
import {
  DATE_TIME_ZONE,
  QUICK_RANGES,
  atTime,
  timeOf,
  windowOf,
  type QuickRange,
  timeZoneOf,
  type TimeRange,
} from './time-range';

// Draft the absolute range being edited: the days (a DateRange of the
// calendar), then their times
type Draft = {
  days: { from: Date | undefined; to?: Date | undefined };
  fromTime: string;
  toTime: string;
};

// The draft of a range, its bounds at now
function draftOf(range: TimeRange): Draft {
  const now = Date.now();
  const { from, to = now } = windowOf(range, now);
  return {
    days: { from: new Date(from), to: new Date(to) },
    fromTime: timeOf(new Date(from)),
    toTime: timeOf(new Date(to)),
  };
}

interface Props {
  value: TimeRange;
  onChange: (range: TimeRange) => void;
}

/**
 * Kibana-style time picker: the quick ranges ending now, or an absolute one
 * from the calendar (days and times, local time).
 */
export function TimeRangePicker({ value, onChange }: Props) {
  const t = useTranslations('Explorer');
  const format = useFormatter();
  const locale = useLocale();
  const id = useId();
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState<Draft>(() => draftOf(value));

  const from = draft.days.from && atTime(draft.days.from, draft.fromTime);
  const to = draft.days.to && atTime(draft.days.to, draft.toTime);
  const invalid = !from || !to || from >= to;

  const label =
    'quick' in value
      ? t(`range.${value.quick}`)
      : format.dateTimeRange(new Date(value.from), new Date(value.to), DATE_TIME_ZONE);

  const openChange = (next: boolean) => {
    // Opening edits the shown range again
    if (next) setDraft(draftOf(value));
    setOpen(next);
  };
  const pickQuick = (quick: QuickRange) => {
    onChange({ quick });
    setOpen(false);
  };
  const apply = () => {
    if (invalid) return;
    onChange({ from: from.getTime(), to: to.getTime() });
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={openChange}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="max-w-full justify-start font-normal"
          aria-label={t('timeRange')}
        >
          <CalendarIcon />
          <span className="truncate">{label}</span>
          <ChevronDown className="opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-auto max-w-[calc(100vw-2rem)] p-0">
        <div className="flex max-h-[80vh] flex-col overflow-y-auto sm:flex-row">
          <div className="flex flex-col gap-1 border-b p-3 sm:w-44 sm:border-r sm:border-b-0">
            <p className="pb-1 text-xs font-medium text-muted-foreground">{t('quickRanges')}</p>
            {(Object.keys(QUICK_RANGES) as QuickRange[]).map((quick) => {
              const active = 'quick' in value && value.quick === quick;
              return (
                <Button
                  key={quick}
                  size="sm"
                  variant={active ? 'secondary' : 'ghost'}
                  aria-pressed={active}
                  className="justify-start"
                  onClick={() => pickQuick(quick)}
                >
                  {t(`range.${quick}`)}
                </Button>
              );
            })}
          </div>

          <div className="flex flex-col gap-3 p-3">
            <p className="text-xs font-medium text-muted-foreground">{t('absoluteRange')}</p>
            <Calendar
              mode="range"
              selected={draft.days}
              onSelect={(days) => setDraft((d) => ({ ...d, days: days ?? { from: undefined } }))}
              disabled={{ after: new Date() }}
              defaultMonth={draft.days.to}
              className="p-0"
            />
            <div className="grid grid-cols-2 gap-2">
              <div className="flex flex-col gap-1">
                <Label htmlFor={`${id}-from`} className="text-xs">
                  {t('startTime')} ({timeZoneOf(locale, from)})
                </Label>
                <Input
                  id={`${id}-from`}
                  type="time"
                  step={1}
                  value={draft.fromTime}
                  onChange={(e) => setDraft((d) => ({ ...d, fromTime: e.target.value }))}
                  className="h-8 font-mono text-xs"
                />
              </div>
              <div className="flex flex-col gap-1">
                <Label htmlFor={`${id}-to`} className="text-xs">
                  {t('endTime')} ({timeZoneOf(locale, to)})
                </Label>
                <Input
                  id={`${id}-to`}
                  type="time"
                  step={1}
                  value={draft.toTime}
                  onChange={(e) => setDraft((d) => ({ ...d, toTime: e.target.value }))}
                  className="h-8 font-mono text-xs"
                />
              </div>
            </div>
            <p className="min-h-4 text-xs text-muted-foreground" aria-live="polite">
              {invalid ? t('invalidRange') : format.dateTimeRange(from, to, DATE_TIME_ZONE)}
            </p>
            <Button size="sm" disabled={invalid} onClick={apply}>
              {t('apply')}
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}
