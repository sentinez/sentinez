'use client';

import axios from 'axios';
import { CheckIcon } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useEffect, useState } from 'react';
import { Badge } from '@sentinez/ui/components/badge';
import { Button } from '@sentinez/ui/components/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@sentinez/ui/components/dialog';
import { Input } from '@sentinez/ui/components/input';
import { RuleTag } from '@sentinez/proto/sentinez/apps/security/v1/security';
import { Version } from '@sentinez/proto/sentinez/types/coreruleset/v1/setting';
import { useDebounced } from '@/hooks/use-debounced';
import { listRuleTags } from '@/lib/api/security';
import { toast } from '@/lib/toast';
import { PICKER_PAGE_SIZE, PickerPager } from './rule-picker';

interface Props {
  /** Catalog version, VERSION_UNSPECIFIED = latest supported */
  version: Version;
  selected: string[];
  onToggle: (tag: string) => void;
  onClose: () => void;
}

/** Dialog browsing the tags of the OWASP CRS catalog; mount it only while it is open */
export function TagPicker({ version, selected, onToggle, onClose }: Props) {
  const t = useTranslations('CoreRulesetFields');
  const tc = useTranslations('Common');
  const [search, setSearch] = useState('');
  const query = useDebounced(search);
  const [pageIndex, setPageIndex] = useState(0);
  const [tags, setTags] = useState<RuleTag[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    const request = {
      page: { index: pageIndex + 1, size: PICKER_PAGE_SIZE, total: true },
      version,
      search: query,
    };
    listRuleTags(request, { signal: controller.signal })
      .then((res) => {
        setTags(res.tags);
        setTotal(res.total);
      })
      .catch((err) => {
        if (axios.isCancel(err) || controller.signal.aborted) return;
        toast.error(t('tagPickerFailed'));
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [query, pageIndex, version, t]);

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{t('tagPickerTitle')}</DialogTitle>
          <DialogDescription>{t('tagPickerDescription')}</DialogDescription>
        </DialogHeader>

        <Input
          autoFocus
          aria-label={t('tagPickerSearch')}
          placeholder={t('tagPickerSearch')}
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            setPageIndex(0);
          }}
        />

        <div className="h-88 overflow-y-auto rounded-md border p-1">
          {tags.length === 0 ? (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              {loading ? tc('loading') : tc('noResults')}
            </div>
          ) : (
            tags.map((tag) => {
              const checked = selected.includes(tag.name);
              return (
                <button
                  key={tag.name}
                  type="button"
                  aria-pressed={checked}
                  className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-left outline-none hover:bg-accent focus-visible:bg-accent"
                  onClick={() => onToggle(tag.name)}
                >
                  <CheckIcon className={checked ? 'w-4 h-4' : 'w-4 h-4 invisible'} />
                  <span className="min-w-0 flex-1 truncate font-mono text-sm" title={tag.name}>
                    {tag.name}
                  </span>
                  <Badge variant="outline" className="tabular-nums">
                    {t('tagRuleCount', { count: tag.ruleCount })}
                  </Badge>
                </button>
              );
            })
          )}
        </div>

        <PickerPager pageIndex={pageIndex} total={total} onPageChange={setPageIndex} />

        <DialogFooter>
          <Button type="button" onClick={onClose}>
            {tc('done')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
