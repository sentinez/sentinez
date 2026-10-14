'use client';

import { PlusIcon, SearchIcon, Trash2Icon } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useState, type ReactNode } from 'react';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { Switch } from '@sentinez/ui/components/switch';
import { Version } from '@sentinez/proto/sentinez/types/coreruleset/v1/setting';
import { Expression } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { CRS_RULE_ID_MAX, CRS_RULE_ID_MIN } from '@/lib/type/security';
import {
  ExclusionFormValue,
  ExclusionTargetFormValue,
  RuleOverrideFormValue,
  createEmptyExclusion,
  createEmptyExclusionTarget,
  createEmptyRuleOverride,
  isCrsRuleId,
  splitList,
  toggleTag,
} from './core-ruleset-form';
import { QueryBuilder, QueryBuilderEditor } from './query-builder';
import { RuleCatalog, RulePicker } from './rule-picker';
import { OptionSelect } from './sec-rule-fields';
import { TagPicker } from './tag-picker';

/** A labelled switch with a hint, laid out as one row */
export function SwitchField({
  id,
  label,
  hint,
  checked,
  onChange,
}: {
  id: string;
  label: string;
  hint?: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  return (
    <div className="flex items-start justify-between gap-4">
      <div className="grid gap-1">
        <Label htmlFor={id}>{label}</Label>
        {hint && <p className="text-sm text-muted-foreground">{hint}</p>}
      </div>
      <Switch id={id} checked={checked} onCheckedChange={onChange} />
    </div>
  );
}

function RuleIdField({
  id,
  value,
  onChange,
  catalog,
  version,
}: {
  id: string;
  value: number;
  onChange: (ruleId: number) => void;
  catalog: RuleCatalog;
  version: Version;
}) {
  const t = useTranslations('CoreRulesetFields');
  const [picking, setPicking] = useState(false);
  const info = catalog.infos.get(value);
  const unknown = !info && catalog.ready && isCrsRuleId(value);

  return (
    <div className="grid gap-2 content-start">
      <Label htmlFor={id}>{t('ruleId')}</Label>
      <div className="flex gap-2">
        <Input
          id={id}
          type="number"
          inputMode="numeric"
          min={CRS_RULE_ID_MIN}
          max={CRS_RULE_ID_MAX}
          step={1}
          placeholder="942100"
          className="tabular-nums"
          value={value || ''}
          onChange={(e) => onChange(e.target.valueAsNumber)}
        />
        <Button type="button" variant="outline" onClick={() => setPicking(true)}>
          <SearchIcon className="w-4 h-4" />
          {t('browse')}
        </Button>
      </div>
      {(info || unknown) && (
        <p className="truncate text-sm text-muted-foreground" title={info?.msg}>
          {info ? info.msg : t('unknownRule')}
        </p>
      )}
      {picking && (
        <RulePicker
          version={version}
          onClose={() => setPicking(false)}
          onSelect={(rule) => {
            onChange(rule.id);
            setPicking(false);
          }}
        />
      )}
    </div>
  );
}

function TagsField({
  id,
  value,
  onChange,
  version,
}: {
  id: string;
  value: string;
  onChange: (tags: string) => void;
  version: Version;
}) {
  const t = useTranslations('CoreRulesetFields');
  const [picking, setPicking] = useState(false);

  return (
    <div className="grid gap-2">
      <Label htmlFor={id}>{t('tags')}</Label>
      <div className="flex gap-2">
        <Input
          id={id}
          placeholder={t('tagsPlaceholder')}
          spellCheck={false}
          autoComplete="off"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
        <Button type="button" variant="outline" onClick={() => setPicking(true)}>
          <SearchIcon className="w-4 h-4" />
          {t('browse')}
        </Button>
      </div>
      <p className="text-sm text-muted-foreground">{t('tagsHint')}</p>
      {picking && (
        <TagPicker
          version={version}
          selected={splitList(value)}
          onToggle={(tag) => onChange(toggleTag(value, tag))}
          onClose={() => setPicking(false)}
        />
      )}
    </div>
  );
}

function ScopeField({
  id,
  scoped,
  expr,
  onChange,
}: {
  id: string;
  scoped: boolean;
  expr: Expression;
  onChange: (patch: { scoped?: boolean; expr?: Expression }) => void;
}) {
  const t = useTranslations('CoreRulesetFields');
  return (
    <div className="grid gap-2">
      <SwitchField
        id={id}
        label={t('scoped')}
        hint={t('scopedHint')}
        checked={scoped}
        onChange={(checked) => onChange({ scoped: checked })}
      />
      {scoped && (
        <div className="-mx-1">
          <QueryBuilder value={expr} onValueChange={(next) => onChange({ expr: next })}>
            <QueryBuilderEditor />
          </QueryBuilder>
        </div>
      )}
    </div>
  );
}

function ItemCard({
  title,
  onRemove,
  children,
}: {
  title: string;
  onRemove: () => void;
  children: ReactNode;
}) {
  const tc = useTranslations('Common');
  return (
    <div className="grid gap-4 rounded-md border p-4">
      <div className="flex items-center justify-between gap-2">
        <span className="text-sm font-medium">{title}</span>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label={`${tc('remove')}: ${title}`}
          onClick={onRemove}
        >
          <Trash2Icon className="w-4 h-4" />
        </Button>
      </div>
      {children}
    </div>
  );
}

interface ListProps<T> {
  value: T[];
  onChange: (value: T[]) => void;
  catalog: RuleCatalog;
  version: Version;
  idPrefix: string;
}

const patchAt = <T,>(list: T[], index: number, patch: Partial<T>) =>
  list.map((item, i) => (i === index ? { ...item, ...patch } : item));

const removeAt = <T,>(list: T[], index: number) => list.filter((_, i) => i !== index);

function ExclusionTargets({
  value,
  onChange,
  catalog,
  version,
  idPrefix,
}: ListProps<ExclusionTargetFormValue>) {
  const t = useTranslations('CoreRulesetFields');
  const tc = useTranslations('Common');
  return (
    <div className="grid gap-2">
      <Label>{t('targets')}</Label>
      {value.map((target, i) => (
        <div key={i} className="grid grid-cols-1 sm:grid-cols-[1fr_1fr_auto] gap-2 items-start">
          <RuleIdField
            id={`${idPrefix}-${i}-rule`}
            value={target.ruleId}
            onChange={(ruleId) => onChange(patchAt(value, i, { ruleId }))}
            catalog={catalog}
            version={version}
          />
          <div className="grid gap-2 content-start">
            <Label htmlFor={`${idPrefix}-${i}-variable`}>{t('variable')}</Label>
            <Input
              id={`${idPrefix}-${i}-variable`}
              placeholder={t('variablePlaceholder')}
              spellCheck={false}
              autoComplete="off"
              value={target.variable}
              onChange={(e) => onChange(patchAt(value, i, { variable: e.target.value }))}
            />
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="sm:mt-5.5"
            aria-label={tc('remove')}
            onClick={() => onChange(removeAt(value, i))}
          >
            <Trash2Icon className="w-4 h-4" />
          </Button>
        </div>
      ))}
      <div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => onChange([...value, createEmptyExclusionTarget()])}
        >
          <PlusIcon className="w-4 h-4" />
          {t('addTarget')}
        </Button>
      </div>
    </div>
  );
}

/** Editor for CoreRuleset.exclusions */
export function ExclusionsEditor({
  value,
  onChange,
  catalog,
  version,
  idPrefix,
}: ListProps<ExclusionFormValue>) {
  const t = useTranslations('CoreRulesetFields');
  return (
    <div className="grid gap-4">
      {value.length === 0 && <p className="text-sm text-muted-foreground">{t('noExclusions')}</p>}
      {value.map((exclusion, i) => (
        <ItemCard
          key={i}
          title={t('exclusionN', { n: i + 1 })}
          onRemove={() => onChange(removeAt(value, i))}
        >
          <TagsField
            id={`${idPrefix}-${i}-tags`}
            value={exclusion.tags}
            onChange={(tags) => onChange(patchAt(value, i, { tags }))}
            version={version}
          />
          <ExclusionTargets
            value={exclusion.targets}
            onChange={(targets) => onChange(patchAt(value, i, { targets }))}
            catalog={catalog}
            version={version}
            idPrefix={`${idPrefix}-${i}-target`}
          />
          <ScopeField
            id={`${idPrefix}-${i}-scoped`}
            scoped={exclusion.scoped}
            expr={exclusion.expr}
            onChange={(patch) => onChange(patchAt(value, i, patch))}
          />
        </ItemCard>
      ))}
      <div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => onChange([...value, createEmptyExclusion()])}
        >
          <PlusIcon className="w-4 h-4" />
          {t('addExclusion')}
        </Button>
      </div>
    </div>
  );
}

/** Editor for CoreRuleset.overrides */
export function OverridesEditor({
  value,
  onChange,
  catalog,
  version,
  idPrefix,
}: ListProps<RuleOverrideFormValue>) {
  const t = useTranslations('CoreRulesetFields');
  const { ruleStateOptions } = useSecurityOptions();
  return (
    <div className="grid gap-4">
      {value.length === 0 && <p className="text-sm text-muted-foreground">{t('noOverrides')}</p>}
      {value.map((override, i) => (
        <ItemCard
          key={i}
          title={t('overrideN', { n: i + 1 })}
          onRemove={() => onChange(removeAt(value, i))}
        >
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <RuleIdField
              id={`${idPrefix}-${i}-rule`}
              value={override.id}
              onChange={(id) => onChange(patchAt(value, i, { id }))}
              catalog={catalog}
              version={version}
            />
            <OptionSelect
              label={t('state')}
              value={override.state}
              options={ruleStateOptions}
              onChange={(state) => onChange(patchAt(value, i, { state }))}
            />
          </div>
          <ScopeField
            id={`${idPrefix}-${i}-scoped`}
            scoped={override.scoped}
            expr={override.expr}
            onChange={(patch) => onChange(patchAt(value, i, patch))}
          />
        </ItemCard>
      ))}
      <div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => onChange([...value, createEmptyRuleOverride()])}
        >
          <PlusIcon className="w-4 h-4" />
          {t('addOverride')}
        </Button>
      </div>
    </div>
  );
}
