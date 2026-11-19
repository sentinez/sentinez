import {
  Condition,
  ConditionSchema,
  FieldSource,
  KeyMode,
  Operator,
  OperatorSpec,
  SourceSpec,
  SourceVariant,
  ValueKind,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';

// Lookups in the condition schema of GET /security/condition-schema. The
// server rejects a condition outside of it, so the editor only offers what
// it lists. Without a schema (sample mode) every lookup returns undefined
// and the editor offers everything, as before.

/** The keys the edge reads as no key, for the sources with a presence variant */
const PRESENCE_KEYS: Partial<Record<FieldSource, string>> = {
  [FieldSource.FIELD_SOURCE_HEADER]: 'header',
  [FieldSource.FIELD_SOURCE_QUERY]: 'query',
};

function sourceSpecOf(
  schema: ConditionSchema | undefined,
  source: FieldSource,
): SourceSpec | undefined {
  return schema?.sources.find((s) => s.source === source);
}

/** Whether the source takes a key in one of its variants (header, query) */
export function acceptsKey(spec: SourceSpec | undefined): boolean {
  return !!spec?.variants.some((v) => v.keyMode === KeyMode.KEY_MODE_REQUIRED);
}

/** Whether an empty key is allowed too: it checks the presence of fields */
export function keyIsOptional(spec: SourceSpec | undefined): boolean {
  return !!spec?.variants.some((v) => v.keyMode === KeyMode.KEY_MODE_NONE);
}

/** The variant a condition reads: the one of its key mode */
function variantOf(spec: SourceSpec | undefined, key: string): SourceVariant | undefined {
  const none = key === '' || (!!spec && key === PRESENCE_KEYS[spec.source]);
  const mode = none ? KeyMode.KEY_MODE_NONE : KeyMode.KEY_MODE_REQUIRED;
  return spec?.variants.find((v) => v.keyMode === mode);
}

function operatorSpecOf(
  variant: SourceVariant | undefined,
  operator: Operator,
): OperatorSpec | undefined {
  return variant?.operators.find((o) => o.operator === operator);
}

/** The spec of the condition: its source, variant and operator */
export function conditionSpec(schema: ConditionSchema | undefined, condition: Condition) {
  const source = sourceSpecOf(schema, condition.source);
  const variant = variantOf(source, condition.key);
  const operator = operatorSpecOf(variant, condition.operator);
  return { source, variant, operator };
}

/**
 * The condition made to fit the schema after its source, key or operator
 * changed: a key the source does not take is dropped, an operator of
 * another variant is replaced by the first of the variant, and the value
 * is turned into a list or a single value as the operator takes.
 */
export function conformCondition(
  schema: ConditionSchema | undefined,
  condition: Condition,
): Condition {
  const source = sourceSpecOf(schema, condition.source);
  if (!source) return condition;

  const key = acceptsKey(source) ? condition.key : '';
  const variant = variantOf(source, key);
  const first = variant?.operators[0];
  const operator = operatorSpecOf(variant, condition.operator) ?? first;
  if (!operator) return { ...condition, key };

  return {
    ...condition,
    key,
    operator: operator.operator,
    value: conformValue(operator, condition.value),
  };
}

function conformValue(spec: OperatorSpec, value: unknown): string | string[] {
  const items = (Array.isArray(value) ? value : [value]).filter(
    (v): v is string => typeof v === 'string' && v !== '',
  );
  // a list is typed: an item may be a value being typed (GE for GET),
  // the server rejects what is left outside of the enum
  if (spec.list) return items;

  const first = items[0] ?? '';
  const isEnum = spec.valueKind === ValueKind.VALUE_KIND_ENUM;
  return isEnum && !spec.enumValues.includes(first) ? '' : first;
}

/** A list value as the text of its input, and back */
export function listToText(value: unknown): string {
  return Array.isArray(value) ? value.join(', ') : typeof value === 'string' ? value : '';
}

export function textToList(text: string): string[] {
  return text
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
}
