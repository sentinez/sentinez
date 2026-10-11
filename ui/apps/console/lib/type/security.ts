import {
  Category,
  EngineMode,
  RuleState,
  Version,
} from '@sentinez/proto/sentinez/types/coreruleset/v1/setting';
import { ActionType, FieldSource, Operator } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';

// Labels and descriptions live in the "Enum" messages namespace; translate them
// with useSecurityOptions() from '@/hooks/use-security-options'.

/** Selectable field sources, in display order (key into Enum.fieldSource) */
export const FIELD_SOURCES = [
  { key: 'method', value: FieldSource.FIELD_SOURCE_METHOD },
  { key: 'host', value: FieldSource.FIELD_SOURCE_HOST },
  { key: 'path', value: FieldSource.FIELD_SOURCE_PATH },
  { key: 'header', value: FieldSource.FIELD_SOURCE_HEADER },
  { key: 'query', value: FieldSource.FIELD_SOURCE_QUERY },
  { key: 'body', value: FieldSource.FIELD_SOURCE_BODY },
  { key: 'ip', value: FieldSource.FIELD_SOURCE_IP },
  { key: 'ja4', value: FieldSource.FIELD_SOURCE_JA4 },
  { key: 'tls', value: FieldSource.FIELD_SOURCE_TLS },
] as const;

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

// ─── Enum → label / message key maps ─────────────────────────────────────────

/** Stable English names used by statusLabel() / BadgeStatus, not for display */
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

/** Key into Enum.actionType / Enum.actionTypeDescription */
export const ACTION_TYPE_KEY = {
  [ActionType.ACTION_TYPE_UNSPECIFIED]: 'unspecified',
  [ActionType.ACTION_TYPE_BLOCK]: 'block',
  [ActionType.ACTION_TYPE_LOG]: 'log',
  [ActionType.ACTION_TYPE_MODIFY_HEADER]: 'modifyHeader',
  [ActionType.ACTION_TYPE_REDIRECT]: 'redirect',
  [ActionType.ACTION_TYPE_SET_TAG]: 'setTag',
  [ActionType.ACTION_TYPE_ROUTE_TO]: 'routeTo',
  [ActionType.UNRECOGNIZED]: 'unknown',
} as const satisfies Record<ActionType, string>;

/** shadcn Badge variant per action, used in rule tables */
export const ACTION_TYPE_BADGE_VARIANT: Record<
  ActionType,
  'default' | 'secondary' | 'destructive' | 'outline'
> = {
  [ActionType.ACTION_TYPE_UNSPECIFIED]: 'outline',
  [ActionType.ACTION_TYPE_BLOCK]: 'destructive',
  [ActionType.ACTION_TYPE_LOG]: 'secondary',
  [ActionType.ACTION_TYPE_MODIFY_HEADER]: 'outline',
  [ActionType.ACTION_TYPE_REDIRECT]: 'outline',
  [ActionType.ACTION_TYPE_SET_TAG]: 'outline',
  [ActionType.ACTION_TYPE_ROUTE_TO]: 'default',
  [ActionType.UNRECOGNIZED]: 'outline',
};

export interface SelectOption<T> {
  label: string;
  value: T;
  description: string;
}

/** Preset priorities (key into Enum.priority / Enum.priorityDescription) */
export const PRIORITIES = [
  { key: 'critical', value: 1 },
  { key: 'high', value: 10 },
  { key: 'medium', value: 50 },
  { key: 'low', value: 100 },
] as const;

/** Selectable statuses (key into Enum.status / Enum.statusDescription) */
export const STATUSES = [
  { key: 'active', value: Status.STATUS_ACTIVE },
  { key: 'disable', value: Status.STATUS_DISABLE },
] as const;

/** Actions that take no extra value */
export const ACTIONS_WITHOUT_VALUE: ActionType[] = [
  ActionType.ACTION_TYPE_BLOCK,
  ActionType.ACTION_TYPE_LOG,
];

/** Selectable actions, in display order */
export const ACTION_TYPES: ActionType[] = [
  ActionType.ACTION_TYPE_BLOCK,
  ActionType.ACTION_TYPE_LOG,
  ActionType.ACTION_TYPE_MODIFY_HEADER,
  ActionType.ACTION_TYPE_REDIRECT,
  ActionType.ACTION_TYPE_SET_TAG,
  ActionType.ACTION_TYPE_ROUTE_TO,
];

// ─── OWASP Core Rule Set ─────────────────────────────────────────────────────

/** Selectable engine modes (key into Enum.engineMode / Enum.engineModeDescription) */
export const ENGINE_MODES = [
  { key: 'on', value: EngineMode.ENGINE_MODE_ON },
  { key: 'detectionOnly', value: EngineMode.ENGINE_MODE_DETECTION_ONLY },
  { key: 'off', value: EngineMode.ENGINE_MODE_OFF },
] as const;

/** shadcn Badge variant per engine mode, used in the rulesets table */
export const ENGINE_MODE_BADGE_VARIANT: Record<EngineMode, 'default' | 'secondary' | 'outline'> = {
  [EngineMode.ENGINE_MODE_UNSPECIFIED]: 'outline',
  [EngineMode.ENGINE_MODE_OFF]: 'outline',
  [EngineMode.ENGINE_MODE_DETECTION_ONLY]: 'secondary',
  [EngineMode.ENGINE_MODE_ON]: 'default',
  [EngineMode.UNRECOGNIZED]: 'outline',
};

/** Pinned CRS versions, newest first; VERSION_UNSPECIFIED means "latest supported" */
export const CRS_VERSIONS = [
  { label: '4.17.0', value: Version.VERSION_V4_17_0 },
  { label: '4.16.0', value: Version.VERSION_V4_16_0 },
] as const;

/** CRS rule groups in rule-ID order (key into Enum.crsCategory); `code` is the ID prefix */
export const CRS_CATEGORIES = [
  { key: 'commonExceptions', code: 905, value: Category.CATEGORY_COMMON_EXCEPTIONS },
  { key: 'methodEnforcement', code: 911, value: Category.CATEGORY_METHOD_ENFORCEMENT },
  { key: 'scannerDetection', code: 913, value: Category.CATEGORY_SCANNER_DETECTION },
  { key: 'protocolEnforcement', code: 920, value: Category.CATEGORY_PROTOCOL_ENFORCEMENT },
  { key: 'protocolAttack', code: 921, value: Category.CATEGORY_PROTOCOL_ATTACK },
  { key: 'multipartAttack', code: 922, value: Category.CATEGORY_MULTIPART_ATTACK },
  { key: 'lfi', code: 930, value: Category.CATEGORY_LFI },
  { key: 'rfi', code: 931, value: Category.CATEGORY_RFI },
  { key: 'rce', code: 932, value: Category.CATEGORY_RCE },
  { key: 'php', code: 933, value: Category.CATEGORY_PHP },
  { key: 'generic', code: 934, value: Category.CATEGORY_GENERIC },
  { key: 'xss', code: 941, value: Category.CATEGORY_XSS },
  { key: 'sqli', code: 942, value: Category.CATEGORY_SQLI },
  { key: 'sessionFixation', code: 943, value: Category.CATEGORY_SESSION_FIXATION },
  { key: 'java', code: 944, value: Category.CATEGORY_JAVA },
  { key: 'dataLeakages', code: 950, value: Category.CATEGORY_DATA_LEAKAGES },
  { key: 'dataLeakagesSql', code: 951, value: Category.CATEGORY_DATA_LEAKAGES_SQL },
  { key: 'dataLeakagesJava', code: 952, value: Category.CATEGORY_DATA_LEAKAGES_JAVA },
  { key: 'dataLeakagesPhp', code: 953, value: Category.CATEGORY_DATA_LEAKAGES_PHP },
  { key: 'dataLeakagesIis', code: 954, value: Category.CATEGORY_DATA_LEAKAGES_IIS },
  { key: 'webShells', code: 955, value: Category.CATEGORY_WEB_SHELLS },
  { key: 'dataLeakagesRuby', code: 956, value: Category.CATEGORY_DATA_LEAKAGES_RUBY },
] as const;

/** HTTP methods offered for RequestPolicy.allowed_methods */
export const HTTP_METHODS = [
  'GET',
  'HEAD',
  'POST',
  'PUT',
  'PATCH',
  'DELETE',
  'OPTIONS',
  'CONNECT',
  'TRACE',
];

/** Rule groups from this ID prefix on inspect the response */
export const CRS_RESPONSE_CODE = 950;

/** Selectable rule states (key into Enum.ruleState / Enum.ruleStateDescription) */
export const RULE_STATES = [
  { key: 'disabled', value: RuleState.RULE_STATE_DISABLED },
  { key: 'detectionOnly', value: RuleState.RULE_STATE_DETECTION_ONLY },
  { key: 'enabled', value: RuleState.RULE_STATE_ENABLED },
] as const;

/** Bounds of a CRS rule ID (buf.validate on ExclusionTarget.rule_id / RuleOverride.id) */
export const CRS_RULE_ID_MIN = 900000;
export const CRS_RULE_ID_MAX = 999999;

/** Expression sources the edge cannot translate to SecLang for a scoped exclusion / override */
export const CRS_UNSUPPORTED_SOURCES: FieldSource[] = [
  FieldSource.FIELD_SOURCE_BODY,
  FieldSource.FIELD_SOURCE_JA4,
  FieldSource.FIELD_SOURCE_TLS,
];
