'use client';

import { ChevronDownIcon } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { useMemo, type ReactNode } from 'react';
import { Button } from '@sentinez/ui/components/button';
import { Checkbox } from '@sentinez/ui/components/checkbox';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { Textarea } from '@sentinez/ui/components/textarea';
import { Category } from '@sentinez/proto/sentinez/types/coreruleset/v1/setting';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { HTTP_METHODS, SelectOption } from '@/lib/type/security';
import { CoreRulesetFormValue } from './core-ruleset-form';
import { ExclusionsEditor, OverridesEditor, SwitchField } from './core-ruleset-rules';
import { useRuleInfos } from './rule-picker';
import { OptionSelect } from './sec-rule-fields';

const PARANOIA_LEVELS = [1, 2, 3, 4];

interface Props {
  value: CoreRulesetFormValue;
  onChange: (patch: Partial<CoreRulesetFormValue>) => void;
  idPrefix?: string;
}

function Section({ title, hint, children }: { title: string; hint: string; children: ReactNode }) {
  return (
    <section className="grid gap-4 border-t pt-6 mt-2">
      <div className="grid gap-1">
        <h2 className="text-base font-semibold">{title}</h2>
        <p className="text-sm text-muted-foreground">{hint}</p>
      </div>
      {children}
    </section>
  );
}

function CountField({
  id,
  label,
  hint,
  value,
  onChange,
}: {
  id: string;
  label: string;
  hint: string;
  value: number;
  onChange: (value: number) => void;
}) {
  return (
    <div className="grid gap-2 content-start">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        type="number"
        inputMode="numeric"
        min={0}
        step={1}
        className="tabular-nums"
        value={Number.isNaN(value) ? '' : value}
        onChange={(e) => onChange(e.target.valueAsNumber)}
      />
      <p className="text-sm text-muted-foreground">{hint}</p>
    </div>
  );
}

function ListField({
  id,
  label,
  hint,
  placeholder,
  value,
  onChange,
}: {
  id: string;
  label: string;
  hint: string;
  placeholder: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div className="grid gap-2 content-start">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        placeholder={placeholder}
        spellCheck={false}
        autoComplete="off"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
      <p className="text-sm text-muted-foreground">{hint}</p>
    </div>
  );
}

/** Multi-select over HTTP methods; methods outside HTTP_METHODS stay selectable */
function MethodSelect({
  id,
  label,
  hint,
  placeholder,
  value,
  onChange,
}: {
  id: string;
  label: string;
  hint: string;
  placeholder: string;
  value: string[];
  onChange: (value: string[]) => void;
}) {
  const methods = [...HTTP_METHODS, ...value.filter((m) => !HTTP_METHODS.includes(m))];
  // keep the display order whatever the order of selection
  const toggle = (method: string, checked: boolean) =>
    onChange(methods.filter((m) => (m === method ? checked : value.includes(m))));

  return (
    <div className="grid gap-2 content-start">
      <Label htmlFor={id}>{label}</Label>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            id={id}
            type="button"
            variant="outline"
            className="w-full justify-between font-normal"
          >
            <span className={value.length ? 'truncate' : 'truncate text-muted-foreground'}>
              {value.length ? methods.filter((m) => value.includes(m)).join(', ') : placeholder}
            </span>
            <ChevronDownIcon className="size-4 opacity-50" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-(--radix-dropdown-menu-trigger-width)">
          {methods.map((method) => (
            <DropdownMenuCheckboxItem
              key={method}
              checked={value.includes(method)}
              onCheckedChange={(checked) => toggle(method, checked === true)}
              // stay open to pick several methods
              onSelect={(e) => e.preventDefault()}
            >
              {method}
            </DropdownMenuCheckboxItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
      <p className="text-sm text-muted-foreground">{hint}</p>
    </div>
  );
}

function CategoryGroup({
  title,
  options,
  selected,
  onToggle,
  idPrefix,
}: {
  title: string;
  options: { label: string; value: Category; code: number }[];
  selected: Category[];
  onToggle: (category: Category, checked: boolean) => void;
  idPrefix: string;
}) {
  return (
    <fieldset className="grid gap-2">
      <legend className="mb-2 text-sm font-medium text-muted-foreground">{title}</legend>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-2">
        {options.map((o) => (
          <div key={o.value} className="flex items-center gap-2">
            <Checkbox
              id={`${idPrefix}-category-${o.value}`}
              checked={selected.includes(o.value)}
              onCheckedChange={(checked) => onToggle(o.value, checked === true)}
            />
            <Label htmlFor={`${idPrefix}-category-${o.value}`} className="font-normal">
              <span className="font-mono text-xs text-muted-foreground tabular-nums">{o.code}</span>
              {o.label}
            </Label>
          </div>
        ))}
      </div>
    </fieldset>
  );
}

/** Form fields matching v1CoreRuleset in security.swagger.json */
export function CoreRulesetFields({ value, onChange, idPrefix = 'ruleset' }: Props) {
  const t = useTranslations('CoreRulesetFields');
  const options = useSecurityOptions();

  const levels = PARANOIA_LEVELS.map((level) => t('paranoiaN', { level }));
  const levelOptions = (zeroLabel: string, description: string): SelectOption<number>[] =>
    [zeroLabel, ...levels].map((label, level) => ({ label, value: level, description }));

  const ruleIds = useMemo(
    () => [
      ...value.overrides.map((o) => o.id),
      ...value.exclusions.flatMap((e) => e.targets.map((target) => target.ruleId)),
    ],
    [value.overrides, value.exclusions],
  );
  const catalog = useRuleInfos(ruleIds, value.version);

  const toggleCategory = (category: Category, checked: boolean) =>
    onChange({
      categories: checked
        ? [...value.categories, category]
        : value.categories.filter((c) => c !== category),
    });

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-name`}>{t('name')}</Label>
        <Input
          id={`${idPrefix}-name`}
          placeholder={t('namePlaceholder')}
          value={value.name}
          onChange={(e) => onChange({ name: e.target.value })}
        />
      </div>
      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-description`}>{t('description')}</Label>
        <Textarea
          id={`${idPrefix}-description`}
          placeholder={t('description')}
          value={value.description}
          onChange={(e) => onChange({ description: e.target.value })}
        />
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <OptionSelect
          label={t('status')}
          value={value.status}
          options={options.statusOptions}
          onChange={(status) => onChange({ status })}
        />
        <OptionSelect
          label={t('mode')}
          value={value.mode}
          options={options.engineModeOptions}
          onChange={(mode) => onChange({ mode })}
        />
      </div>

      <Section title={t('engine')} hint={t('engineHint')}>
        <OptionSelect
          label={t('version')}
          value={value.version}
          options={options.crsVersionOptions}
          onChange={(version) => onChange({ version })}
        />
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <OptionSelect
            label={t('paranoiaLevel')}
            value={value.paranoiaLevel}
            options={levelOptions(t('paranoiaDefault'), t('paranoiaHint'))}
            onChange={(paranoiaLevel) => onChange({ paranoiaLevel })}
          />
          <OptionSelect
            label={t('detectionParanoiaLevel')}
            value={value.detectionParanoiaLevel}
            options={levelOptions(t('detectionSame'), t('detectionParanoiaHint'))}
            onChange={(detectionParanoiaLevel) => onChange({ detectionParanoiaLevel })}
          />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <CountField
            id={`${idPrefix}-inbound`}
            label={t('inboundThreshold')}
            hint={t('thresholdHint')}
            value={value.inboundThreshold}
            onChange={(inboundThreshold) => onChange({ inboundThreshold })}
          />
          <CountField
            id={`${idPrefix}-outbound`}
            label={t('outboundThreshold')}
            hint={t('thresholdHint')}
            value={value.outboundThreshold}
            onChange={(outboundThreshold) => onChange({ outboundThreshold })}
          />
        </div>
        <SwitchField
          id={`${idPrefix}-early-blocking`}
          label={t('earlyBlocking')}
          hint={t('earlyBlockingHint')}
          checked={value.earlyBlocking}
          onChange={(earlyBlocking) => onChange({ earlyBlocking })}
        />
      </Section>

      <Section title={t('ruleGroups')} hint={t('ruleGroupsHint')}>
        <SwitchField
          id={`${idPrefix}-all-categories`}
          label={t('allGroups')}
          checked={value.allCategories}
          onChange={(allCategories) => onChange({ allCategories })}
        />
        {!value.allCategories && (
          <>
            <CategoryGroup
              title={t('requestPhase')}
              options={options.crsCategoryOptions.filter((o) => !o.response)}
              selected={value.categories}
              onToggle={toggleCategory}
              idPrefix={idPrefix}
            />
            <CategoryGroup
              title={t('responsePhase')}
              options={options.crsCategoryOptions.filter((o) => o.response)}
              selected={value.categories}
              onToggle={toggleCategory}
              idPrefix={idPrefix}
            />
          </>
        )}
      </Section>

      <Section title={t('request')} hint={t('requestHint')}>
        <SwitchField
          id={`${idPrefix}-body-access`}
          label={t('bodyAccess')}
          hint={t('bodyAccessHint')}
          checked={value.bodyAccess}
          onChange={(bodyAccess) => onChange({ bodyAccess })}
        />
        <CountField
          id={`${idPrefix}-body-limit`}
          label={t('bodyLimit')}
          hint={t('bodyLimitHint')}
          value={value.bodyLimitBytes}
          onChange={(bodyLimitBytes) => onChange({ bodyLimitBytes })}
        />
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <MethodSelect
            id={`${idPrefix}-methods`}
            label={t('allowedMethods')}
            hint={t('allowedMethodsHint')}
            placeholder={t('allowedMethodsDefault')}
            value={value.allowedMethods}
            onChange={(allowedMethods) => onChange({ allowedMethods })}
          />
          <ListField
            id={`${idPrefix}-content-types`}
            label={t('allowedContentTypes')}
            hint={t('allowedContentTypesHint')}
            placeholder="application/json"
            value={value.allowedContentTypes}
            onChange={(allowedContentTypes) => onChange({ allowedContentTypes })}
          />
        </div>
      </Section>

      <Section title={t('exclusions')} hint={t('exclusionsHint')}>
        <ExclusionsEditor
          value={value.exclusions}
          onChange={(exclusions) => onChange({ exclusions })}
          catalog={catalog}
          version={value.version}
          idPrefix={`${idPrefix}-exclusion`}
        />
      </Section>

      <Section title={t('overrides')} hint={t('overridesHint')}>
        <OverridesEditor
          value={value.overrides}
          onChange={(overrides) => onChange({ overrides })}
          catalog={catalog}
          version={value.version}
          idPrefix={`${idPrefix}-override`}
        />
      </Section>
    </div>
  );
}
