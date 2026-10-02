import { ActionType, FieldSource, Operator } from '@sentinez/proto/sentinez/secure/rule/v1/engine';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';

export const FIELD_SOURCE_OPTIONS: { label: string; value: FieldSource }[] = [
  { label: 'HTTP Method', value: FieldSource.FIELD_SOURCE_METHOD },
  { label: 'HTTP Host', value: FieldSource.FIELD_SOURCE_HOST },
  { label: 'HTTP Path', value: FieldSource.FIELD_SOURCE_PATH },
  { label: 'HTTP Header', value: FieldSource.FIELD_SOURCE_HEADER },
  { label: 'HTTP Query', value: FieldSource.FIELD_SOURCE_QUERY },
  { label: 'HTTP Body', value: FieldSource.FIELD_SOURCE_BODY },
  { label: 'Source IP', value: FieldSource.FIELD_SOURCE_IP },
  { label: 'JA4', value: FieldSource.FIELD_SOURCE_JA4 },
  { label: 'TLS', value: FieldSource.FIELD_SOURCE_TLS },
];

export const OPERATOR_OPTIONS: { label: string; value: Operator }[] = [
  { label: '==', value: Operator.OPERATOR_EQ },
  { label: '!=', value: Operator.OPERATOR_NE },
  { label: '>', value: Operator.OPERATOR_GT },
  { label: '>=', value: Operator.OPERATOR_GTE },
  { label: '<', value: Operator.OPERATOR_LT },
  { label: '<=', value: Operator.OPERATOR_LTE },
  { label: 'contains', value: Operator.OPERATOR_CONTAINS },
  { label: 'prefix', value: Operator.OPERATOR_PREFIX },
  { label: 'suffix', value: Operator.OPERATOR_SUFFIX },
  { label: 'matches', value: Operator.OPERATOR_MATCHES },
  { label: 'in', value: Operator.OPERATOR_IN },
  { label: 'not in', value: Operator.OPERATOR_NOT_IN },
];

// ─── Enum → readable label maps ──────────────────────────────────────────────

export const STATUS_LABEL: Record<Status, string> = {
  [Status.STATUS_UNSPECIFIED]: 'Unspecified',
  [Status.STATUS_ACTIVE]: 'Active',
  [Status.STATUS_DISABLE]: 'Disable',
  [Status.UNRECOGNIZED]: 'Unknown',
};

/** Returns a lowercase label used by BadgeStatus ("active" | "disable" | …) */
export function statusLabel(status: Status | number | undefined): string {
  if (status === undefined) return 'disable';
  return (STATUS_LABEL[status as Status] ?? 'unknown').toLowerCase();
}

export const FIELD_SOURCE_LABEL: Record<FieldSource, string> = {
  [FieldSource.FIELD_SOURCE_UNSPECIFIED]: 'Unspecified',
  [FieldSource.FIELD_SOURCE_HEADER]: 'HTTP Header',
  [FieldSource.FIELD_SOURCE_QUERY]: 'HTTP Query',
  [FieldSource.FIELD_SOURCE_PATH]: 'HTTP Path',
  [FieldSource.FIELD_SOURCE_BODY]: 'HTTP Body',
  [FieldSource.FIELD_SOURCE_IP]: 'Source IP',
  [FieldSource.FIELD_SOURCE_JA4]: 'JA4',
  [FieldSource.FIELD_SOURCE_TLS]: 'TLS',
  [FieldSource.FIELD_SOURCE_METHOD]: 'HTTP Method',
  [FieldSource.FIELD_SOURCE_HOST]: 'HTTP Host',
  [FieldSource.UNRECOGNIZED]: 'Unknown',
};

export const OPERATOR_LABEL: Record<Operator, string> = {
  [Operator.OPERATOR_UNSPECIFIED]: 'Unspecified',
  [Operator.OPERATOR_EQ]: '==',
  [Operator.OPERATOR_NE]: '!=',
  [Operator.OPERATOR_CONTAINS]: 'contains',
  [Operator.OPERATOR_MATCHES]: 'matches',
  [Operator.OPERATOR_IN]: 'in',
  [Operator.OPERATOR_PREFIX]: 'prefix',
  [Operator.OPERATOR_SUFFIX]: 'suffix',
  [Operator.OPERATOR_GT]: '>',
  [Operator.OPERATOR_GTE]: '>=',
  [Operator.OPERATOR_LT]: '<',
  [Operator.OPERATOR_LTE]: '<=',
  [Operator.OPERATOR_NOT_IN]: 'not in',
  [Operator.UNRECOGNIZED]: 'Unknown',
};

export const ACTION_TYPE_LABEL: Record<ActionType, string> = {
  [ActionType.ACTION_TYPE_UNSPECIFIED]: 'Unspecified',
  [ActionType.ACTION_TYPE_BLOCK]: 'Block',
  [ActionType.ACTION_TYPE_LOG]: 'Log',
  [ActionType.ACTION_TYPE_MODIFY_HEADER]: 'Modify Header',
  [ActionType.ACTION_TYPE_REDIRECT]: 'Redirect',
  [ActionType.ACTION_TYPE_SET_TAG]: 'Set Tag',
  [ActionType.ACTION_TYPE_ROUTE_TO]: 'Route To',
  [ActionType.UNRECOGNIZED]: 'Unknown',
};

export interface SelectOption<T> {
  label: string;
  value: T;
  description: string;
}

export const PRIORITY_OPTIONS: SelectOption<number>[] = [
  { label: 'Critical (1)', value: 1, description: 'Evaluated before all other rules.' },
  { label: 'High (10)', value: 10, description: 'Evaluated early, after critical rules.' },
  { label: 'Medium (50)', value: 50, description: 'Default for most rules.' },
  { label: 'Low (100)', value: 100, description: 'Evaluated last, after all other rules.' },
];

export const STATUS_OPTIONS: SelectOption<Status>[] = [
  {
    label: STATUS_LABEL[Status.STATUS_ACTIVE],
    value: Status.STATUS_ACTIVE,
    description: 'The rule is evaluated against incoming requests.',
  },
  {
    label: STATUS_LABEL[Status.STATUS_DISABLE],
    value: Status.STATUS_DISABLE,
    description: 'The rule is kept but skipped during evaluation.',
  },
];

export const ACTION_TYPE_DESCRIPTION: Partial<Record<ActionType, string>> = {
  [ActionType.ACTION_TYPE_BLOCK]: 'Reject the request when the condition matches.',
  [ActionType.ACTION_TYPE_LOG]: 'Allow the request and record a log entry.',
  [ActionType.ACTION_TYPE_MODIFY_HEADER]: 'Set a request header (name and value) when matched.',
  [ActionType.ACTION_TYPE_REDIRECT]: 'Redirect the client to the given URL.',
  [ActionType.ACTION_TYPE_SET_TAG]: 'Attach a tag to the request for downstream use.',
  [ActionType.ACTION_TYPE_ROUTE_TO]: 'Route the request to the given upstream/target.',
};

/** Actions that take no extra value */
export const ACTIONS_WITHOUT_VALUE: ActionType[] = [
  ActionType.ACTION_TYPE_BLOCK,
  ActionType.ACTION_TYPE_LOG,
];

export const ACTION_TYPE_OPTIONS: SelectOption<ActionType>[] = [
  ActionType.ACTION_TYPE_BLOCK,
  ActionType.ACTION_TYPE_LOG,
  ActionType.ACTION_TYPE_MODIFY_HEADER,
  ActionType.ACTION_TYPE_REDIRECT,
  ActionType.ACTION_TYPE_SET_TAG,
  ActionType.ACTION_TYPE_ROUTE_TO,
].map((value) => ({
  label: ACTION_TYPE_LABEL[value],
  value,
  description: ACTION_TYPE_DESCRIPTION[value] ?? '',
}));
