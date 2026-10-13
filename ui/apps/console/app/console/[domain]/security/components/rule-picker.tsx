'use client';

import axios from 'axios';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useEffect, useMemo, useState } from 'react';
import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@sentinez/ui/components/dialog';
import { Input } from '@sentinez/ui/components/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import { Category, RuleInfo, Version } from '@sentinez/proto/sentinez/types/coreruleset/v1/setting';
import { useDebounced } from '@/hooks/use-debounced';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { listRuleInfos } from '@/lib/api/security';
import { toast } from '@/lib/toast';
import { isCrsRuleId } from './core-ruleset-form';
import { cn } from '@sentinez/ui/lib/utils';

export const PICKER_PAGE_SIZE = 8;

const severityVariants = new Map<string, string>([
  ['CRITICAL', 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300'],
  ['WARNING', 'bg-yellow-50 text-yellow-700 dark:bg-yellow-950 dark:text-yellow-300'],
  ['INFO', 'bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300'],
]);

export interface RuleCatalog {
  infos: Map<number, RuleInfo>;
  /** true once the catalog answered for the current rule IDs */
  ready: boolean;
}

/** Catalog entries of the given rule IDs, to show what a typed ID refers to */
export function useRuleInfos(ids: number[], version: Version): RuleCatalog {
  const key = useMemo(
    () => [...new Set(ids.filter(isCrsRuleId))].sort((a, b) => a - b).join(','),
    [ids],
  );
  const [loaded, setLoaded] = useState({ key: '', infos: new Map<number, RuleInfo>() });

  useEffect(() => {
    if (!key) return;
    const controller = new AbortController();
    const request = { ids: key.split(',').map(Number), version, includeSystem: true };
    listRuleInfos(request, { signal: controller.signal })
      .then((res) => setLoaded({ key, infos: new Map(res.rules.map((r) => [r.id, r])) }))
      // the description is a hint only: leave it out when the catalog is unreachable
      .catch(() => {});
    return () => controller.abort();
  }, [key, version]);

  return { infos: loaded.infos, ready: loaded.key === key };
}

/** Range and previous / next buttons under the list of a picker dialog */
export function PickerPager({
  pageIndex,
  total,
  onPageChange,
}: {
  pageIndex: number;
  total: number;
  onPageChange: (pageIndex: number) => void;
}) {
  const tp = useTranslations('Pagination');
  const from = total === 0 ? 0 : pageIndex * PICKER_PAGE_SIZE + 1;
  const to = Math.min(total, (pageIndex + 1) * PICKER_PAGE_SIZE);

  return (
    <div className="flex items-center justify-between gap-4">
      <div className="text-sm text-muted-foreground tabular-nums">
        {tp('range', { from, to, total })}
      </div>
      <div className="flex items-center gap-1">
        <Button
          variant="outline"
          size="icon"
          className="h-8 w-8"
          aria-label={tp('previous')}
          disabled={pageIndex === 0}
          onClick={() => onPageChange(pageIndex - 1)}
        >
          <ChevronLeft />
        </Button>
        <Button
          variant="outline"
          size="icon"
          className="h-8 w-8"
          aria-label={tp('next')}
          disabled={to >= total}
          onClick={() => onPageChange(pageIndex + 1)}
        >
          <ChevronRight />
        </Button>
      </div>
    </div>
  );
}

interface Props {
  /** Catalog version, VERSION_UNSPECIFIED = latest supported */
  version: Version;
  onSelect: (rule: RuleInfo) => void;
  onClose: () => void;
}

/** Dialog searching the OWASP CRS catalog; mount it only while it is open */
export function RulePicker({ version, onSelect, onClose }: Props) {
  const t = useTranslations('CoreRulesetFields');
  const tc = useTranslations('Common');
  const { crsCategoryOptions } = useSecurityOptions();
  const [search, setSearch] = useState('');
  const query = useDebounced(search);
  const [category, setCategory] = useState(Category.CATEGORY_UNSPECIFIED);
  const [pageIndex, setPageIndex] = useState(0);
  const [rules, setRules] = useState<RuleInfo[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    const request = {
      page: { index: pageIndex + 1, size: PICKER_PAGE_SIZE, total: true },
      version,
      search: query,
      categories: category ? [category] : [],
    };
    listRuleInfos(request, { signal: controller.signal })
      .then((res) => {
        setRules(res.rules);
        setTotal(res.total);
      })
      .catch((err) => {
        if (axios.isCancel(err) || controller.signal.aborted) return;
        toast.error(t('pickerFailed'));
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [query, category, pageIndex, version, t]);

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t('pickerTitle')}</DialogTitle>
          <DialogDescription>{t('pickerDescription')}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col sm:flex-row gap-2">
          <Input
            autoFocus
            aria-label={t('pickerSearch')}
            placeholder={t('pickerSearch')}
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPageIndex(0);
            }}
          />
          <Select
            value={String(category)}
            onValueChange={(v) => {
              setCategory(Number(v) as Category);
              setPageIndex(0);
            }}
          >
            <SelectTrigger className="w-full sm:w-64" aria-label={t('ruleGroups')}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={String(Category.CATEGORY_UNSPECIFIED)}>
                {t('allCategories')}
              </SelectItem>
              {crsCategoryOptions.map((o) => (
                <SelectItem key={o.value} value={String(o.value)}>
                  {o.code} · {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="h-88 overflow-y-auto rounded-md border p-1">
          {rules.length === 0 ? (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              {loading ? tc('loading') : tc('noResults')}
            </div>
          ) : (
            rules.map((r) => (
              <button
                key={r.id}
                type="button"
                className="flex w-full items-start gap-3 rounded-md px-3 py-2 text-left outline-none hover:bg-accent focus-visible:bg-accent"
                onClick={() => onSelect(r)}
              >
                <span className="font-mono text-sm tabular-nums">{r.id}</span>
                <span className="grid min-w-0 flex-1 gap-0.5">
                  <span className="truncate text-sm" title={r.msg}>
                    {r.msg}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">
                    {r.tags.join(' · ')}
                  </span>
                </span>
                {r.paranoiaLevel > 0 && (
                  <Badge variant="outline">{t('paranoiaShort', { level: r.paranoiaLevel })}</Badge>
                )}
                {r.severity && (
                  <Badge className={cn(severityVariants.get(r.severity))}>{r.severity}</Badge>
                )}
              </button>
            ))
          )}
        </div>

        <PickerPager pageIndex={pageIndex} total={total} onPageChange={setPageIndex} />
      </DialogContent>
    </Dialog>
  );
}
