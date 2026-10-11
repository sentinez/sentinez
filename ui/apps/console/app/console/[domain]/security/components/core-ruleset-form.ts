import { CoreRuleset } from '@sentinez/proto/sentinez/apps/security/v1/model';
import {
  Category,
  EngineMode,
  Exclusion,
  RuleOverride,
  RuleState,
  Version,
} from '@sentinez/proto/sentinez/types/coreruleset/v1/setting';
import { AndCondition, Expression } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import { CRS_RULE_ID_MAX, CRS_RULE_ID_MIN, CRS_UNSUPPORTED_SOURCES } from '@/lib/type/security';
import { createEmptyExpression } from './query-builder';
import type { ValidationTranslator } from './sec-rule-fields';

export interface ExclusionTargetFormValue {
  ruleId: number;
  /** Variable removed from the rule, e.g. "ARGS:password" */
  variable: string;
}

export interface ExclusionFormValue {
  /** Tags separated by commas or spaces */
  tags: string;
  targets: ExclusionTargetFormValue[];
  /** false = every request (global), true = only the requests matching `expr` */
  scoped: boolean;
  expr: Expression;
}

export interface RuleOverrideFormValue {
  id: number;
  state: RuleState;
  /** false = every request (global), true = only the requests matching `expr` */
  scoped: boolean;
  expr: Expression;
}

export interface CoreRulesetFormValue {
  name: string;
  description: string;
  status: Status;
  mode: EngineMode;
  /** VERSION_UNSPECIFIED = latest supported */
  version: Version;
  /** 0 = default (1) */
  paranoiaLevel: number;
  /** 0 = same as paranoiaLevel */
  detectionParanoiaLevel: number;
  /** 0 = CRS default */
  inboundThreshold: number;
  /** 0 = CRS default */
  outboundThreshold: number;
  earlyBlocking: boolean;
  /** true = every rule group, `categories` is ignored */
  allCategories: boolean;
  categories: Category[];
  bodyAccess: boolean;
  /** 0 = default */
  bodyLimitBytes: number;
  /** Empty = CRS default */
  allowedMethods: string[];
  /** Content types separated by commas or spaces */
  allowedContentTypes: string;
  exclusions: ExclusionFormValue[];
  overrides: RuleOverrideFormValue[];
}

const splitList = (text: string) => text.split(/[\s,]+/).filter(Boolean);

export const isCrsRuleId = (id: number) =>
  Number.isInteger(id) && id >= CRS_RULE_ID_MIN && id <= CRS_RULE_ID_MAX;

export function createEmptyExclusionTarget(): ExclusionTargetFormValue {
  return { ruleId: 0, variable: '' };
}

export function createEmptyExclusion(): ExclusionFormValue {
  return { tags: '', targets: [], scoped: false, expr: createEmptyExpression() };
}

export function createEmptyRuleOverride(): RuleOverrideFormValue {
  return {
    id: 0,
    state: RuleState.RULE_STATE_DISABLED,
    scoped: false,
    expr: createEmptyExpression(),
  };
}

export function createEmptyCoreRulesetForm(): CoreRulesetFormValue {
  return {
    name: '',
    description: '',
    status: Status.STATUS_ACTIVE,
    mode: EngineMode.ENGINE_MODE_DETECTION_ONLY,
    version: Version.VERSION_UNSPECIFIED,
    paranoiaLevel: 1,
    detectionParanoiaLevel: 0,
    inboundThreshold: 0,
    outboundThreshold: 0,
    earlyBlocking: false,
    allCategories: true,
    categories: [],
    bodyAccess: true,
    bodyLimitBytes: 0,
    allowedMethods: [],
    allowedContentTypes: '',
    exclusions: [],
    overrides: [],
  };
}

const scopeOf = (expr?: Expression) => ({
  scoped: !!expr?.orCondition.length,
  expr: expr?.orCondition.length ? expr : createEmptyExpression(),
});

/** CoreRuleset (API) -> form value */
export function coreRulesetFormOf(r: CoreRuleset): CoreRulesetFormValue {
  const c = r.content;
  return {
    name: r.name,
    description: r.description,
    status: c?.status || Status.STATUS_ACTIVE,
    // unspecified is treated as off by the edge
    mode: c?.mode || EngineMode.ENGINE_MODE_OFF,
    version: c?.version ?? Version.VERSION_UNSPECIFIED,
    paranoiaLevel: c?.paranoiaLevel ?? 0,
    detectionParanoiaLevel: c?.detectionParanoiaLevel ?? 0,
    inboundThreshold: c?.threshold?.inbound ?? 0,
    outboundThreshold: c?.threshold?.outbound ?? 0,
    earlyBlocking: c?.threshold?.earlyBlocking ?? false,
    allCategories: !c?.categories.length,
    categories: c?.categories ?? [],
    bodyAccess: c?.request?.bodyAccess ?? false,
    bodyLimitBytes: c?.request?.bodyLimitBytes ?? 0,
    allowedMethods: (c?.request?.allowedMethods ?? []).map((m) => m.toUpperCase()),
    allowedContentTypes: (c?.request?.allowedContentTypes ?? []).join(', '),
    exclusions: (c?.exclusions ?? []).map((e) => ({
      tags: e.tags.join(', '),
      targets: e.targets.map((t) => ({ ruleId: t.ruleId, variable: t.variable })),
      ...scopeOf(e.expr),
    })),
    overrides: (c?.overrides ?? []).map((o) => ({
      id: o.id,
      state: o.state || RuleState.RULE_STATE_DISABLED,
      ...scopeOf(o.expr),
    })),
  };
}

const exclusionOf = (e: ExclusionFormValue): Exclusion => ({
  tags: splitList(e.tags),
  targets: e.targets.map((t) => ({ ruleId: t.ruleId, variable: t.variable.trim() })),
  expr: e.scoped ? e.expr : undefined,
});

const ruleOverrideOf = (o: RuleOverrideFormValue): RuleOverride => ({
  id: o.id,
  state: o.state,
  expr: o.scoped ? o.expr : undefined,
});

/** Form value -> CoreRuleset (API) */
export function coreRulesetOf(f: CoreRulesetFormValue, id = ''): CoreRuleset {
  return {
    id,
    name: f.name.trim(),
    description: f.description,
    content: {
      mode: f.mode,
      version: f.version,
      paranoiaLevel: f.paranoiaLevel,
      detectionParanoiaLevel: f.detectionParanoiaLevel,
      threshold: {
        inbound: f.inboundThreshold,
        outbound: f.outboundThreshold,
        earlyBlocking: f.earlyBlocking,
      },
      categories: f.allCategories ? [] : f.categories,
      request: {
        bodyAccess: f.bodyAccess,
        bodyLimitBytes: f.bodyLimitBytes,
        allowedMethods: f.allowedMethods,
        allowedContentTypes: splitList(f.allowedContentTypes),
      },
      exclusions: f.exclusions.map(exclusionOf),
      overrides: f.overrides.map(ruleOverrideOf),
      status: f.status,
    },
  };
}

const isCount = (n: number) => Number.isInteger(n) && n >= 0;

const hasUnsupportedSource = (groups: AndCondition[]): boolean =>
  groups.some(
    (g) =>
      g.rules.some((r) => !!r.condition && CRS_UNSUPPORTED_SOURCES.includes(r.condition.source)) ||
      hasUnsupportedSource(g.orCondition),
  );

function validateExclusion(e: ExclusionFormValue, n: number, t: ValidationTranslator) {
  if (!splitList(e.tags).length && !e.targets.length) return t('exclusionEmpty', { n });
  if (!e.targets.every((target) => isCrsRuleId(target.ruleId))) return t('crsRuleId');
  if (!e.targets.every((target) => target.variable.trim())) return t('exclusionVariable', { n });
  return null;
}

function validateRuleOverride(o: RuleOverrideFormValue, n: number, t: ValidationTranslator) {
  if (!isCrsRuleId(o.id)) return t('crsRuleId');
  if (o.scoped && o.state !== RuleState.RULE_STATE_DISABLED) return t('overrideScope', { n });
  return null;
}

/** Returns an error message, or null when the form is valid */
export function validateCoreRulesetForm(
  f: CoreRulesetFormValue,
  t: ValidationTranslator,
): string | null {
  if (!f.name.trim()) return t('nameRequired');
  if (!isCount(f.inboundThreshold) || !isCount(f.outboundThreshold)) return t('anomalyThreshold');
  if (!f.allCategories && !f.categories.length) return t('categoriesRequired');
  if (!isCount(f.bodyLimitBytes)) return t('bodyLimit');
  for (const [i, e] of f.exclusions.entries()) {
    const error = validateExclusion(e, i + 1, t);
    if (error) return error;
  }
  for (const [i, o] of f.overrides.entries()) {
    const error = validateRuleOverride(o, i + 1, t);
    if (error) return error;
  }
  const scopes = [...f.exclusions, ...f.overrides].filter((s) => s.scoped);
  if (scopes.some((s) => hasUnsupportedSource(s.expr.orCondition))) return t('scopeSource');
  return null;
}
