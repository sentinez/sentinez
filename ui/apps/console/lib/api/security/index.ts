import { RateLimit, SecRule } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import {
  ActionType,
  FieldSource,
  Operator,
  actionTypeFromJSON,
  actionTypeToJSON,
  Expression,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { Status, statusFromJSON, statusToJSON } from '@sentinez/proto/sentinez/types/v1/known';
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { type Pages, pageQuery as pagesQuery, paginate } from '@/lib/api/pages';
// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

export type { Pages };

/** Wire format of `v1ActionValue` in security.swagger.json */
export interface SecurityActionValue {
  strValue?: string;
  mapValue?: Record<string, string>;
}

/** Wire format of `v1SecRule` in security.swagger.json */
export interface SecuritySecRule {
  metadata?: { createdAt?: string; updatedAt?: string };
  id?: string;
  name?: string;
  description?: string;
  expr?: unknown;
  action?: string;
  actionValue?: SecurityActionValue;
  status?: string;
  priority?: number;
}

/** Wire format of `v1RateLimit` in security.swagger.json */
export interface SecurityRateLimit extends SecuritySecRule {
  timeWindow?: string;
  /** int64, serialized as a string */
  maxRequests?: string;
  timeout?: string;
}

export interface ListSecRulesParams {
  page?: Pages;
  ids?: string[];
}

export interface ListSecRulesResult {
  secRules: SecRule[];
  total: number;
}

export type ListRateLimitsParams = ListSecRulesParams;

export interface ListRateLimitsResult {
  rateLimits: RateLimit[];
  total: number;
}

export interface StatusResponse {
  msg?: string;
}

// proto Action.params <-> swagger actionValue (mapValue); a lone strValue is exposed as params.value
function actionParamsFromWire(v?: SecurityActionValue): Record<string, any> | undefined {
  if (v?.mapValue && Object.keys(v.mapValue).length) return { ...v.mapValue };
  if (v?.strValue) return { value: v.strValue };
  return undefined;
}

function actionValueToWire(params?: Record<string, any>): SecurityActionValue | undefined {
  const keys = Object.keys(params ?? {});
  if (!params || !keys.length) return undefined;
  if (keys.length === 1 && keys[0] === 'value') return { strValue: String(params.value) };
  return { mapValue: Object.fromEntries(Object.entries(params).map(([k, v]) => [k, String(v)])) };
}

function fromWire(rule: SecuritySecRule): SecRule {
  return {
    ingressRuntime: {
      id: rule.id ?? '',
      name: rule.name ?? '',
      description: rule.description ?? '',
      priority: rule.priority ?? 0,
      status: statusFromJSON(rule.status ?? 'STATUS_UNSPECIFIED'),
      expr: rule.expr ? Expression.fromJSON(rule.expr) : undefined,
      action: rule.action
        ? { type: actionTypeFromJSON(rule.action), params: actionParamsFromWire(rule.actionValue) }
        : undefined,
    },
  } as SecRule;
}

function toWire(data: SecRule): SecuritySecRule {
  const r = data.ingressRuntime;
  return {
    id: r?.id || undefined,
    name: r?.name,
    description: r?.description,
    priority: r?.priority,
    status: statusToJSON(r?.status ?? Status.STATUS_UNSPECIFIED),
    expr: r?.expr ? Expression.toJSON(r.expr) : undefined,
    action: r?.action
      ? actionTypeToJSON(r.action.type ?? ActionType.ACTION_TYPE_UNSPECIFIED)
      : undefined,
    actionValue: actionValueToWire(r?.action?.params),
  };
}

function rateLimitFromWire(rl: SecurityRateLimit): RateLimit {
  return {
    ingressRuntime: fromWire(rl).ingressRuntime,
    timeWindow: rl.timeWindow ?? '',
    maxRequests: Number(rl.maxRequests ?? 0),
    timeout: rl.timeout ?? '',
  } as RateLimit;
}

function rateLimitToWire(data: RateLimit): SecurityRateLimit {
  return {
    ...toWire({ ingressRuntime: data.ingressRuntime } as SecRule),
    timeWindow: data.timeWindow,
    maxRequests: String(data.maxRequests ?? 0),
    timeout: data.timeout || undefined,
  };
}

function pageQuery(params?: ListSecRulesParams): Record<string, unknown> {
  const query = pagesQuery(params?.page);
  if (params?.ids?.length) query.ids = params.ids;
  return query;
}

// ---- Sample data, used as fallback when the API call fails ----

const sampleRule = (
  id: string,
  name: string,
  description: string,
  priority: number,
  status: Status,
  expr: Expression,
  action: ActionType = ActionType.ACTION_TYPE_BLOCK,
): SecRule =>
  ({
    ingressRuntime: { id, name, description, priority, status, expr, action: { type: action } },
  }) as SecRule;

const cond = (source: FieldSource, operator: Operator, value: unknown, key = '') => ({
  id: Math.random().toString(36).slice(2, 9),
  expr: '',
  condition: { id: Math.random().toString(36).slice(2, 9), source, key, operator, value },
});

export const SAMPLE_RULES: SecRule[] = [
  sampleRule(
    'sample-rule-1',
    'Block non-GET on /admin',
    'Sample: (Path prefix /admin AND Method != GET)',
    1,
    Status.STATUS_ACTIVE,
    {
      orCondition: [
        {
          rules: [
            cond(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/admin'),
            cond(FieldSource.FIELD_SOURCE_METHOD, Operator.OPERATOR_NE, 'GET'),
          ],
          orCondition: [],
        },
      ],
    } as Expression,
  ),
  sampleRule(
    'sample-rule-2',
    'Block bad bots',
    'Sample: User-Agent contains "curl" OR "sqlmap"',
    2,
    Status.STATUS_ACTIVE,
    {
      orCondition: [
        {
          rules: [
            cond(FieldSource.FIELD_SOURCE_HEADER, Operator.OPERATOR_CONTAINS, 'curl', 'User-Agent'),
          ],
          orCondition: [],
        },
        {
          rules: [
            cond(
              FieldSource.FIELD_SOURCE_HEADER,
              Operator.OPERATOR_CONTAINS,
              'sqlmap',
              'User-Agent',
            ),
          ],
          orCondition: [],
        },
      ],
    } as Expression,
  ),
  sampleRule(
    'sample-rule-3',
    'Log requests from internal range',
    'Sample: IP in 10.0.0.0/8 (disabled)',
    3,
    Status.STATUS_DISABLE,
    {
      orCondition: [
        {
          rules: [cond(FieldSource.FIELD_SOURCE_IP, Operator.OPERATOR_IN, ['10.0.0.0/8'])],
          orCondition: [],
        },
      ],
    } as Expression,
    ActionType.ACTION_TYPE_LOG,
  ),
];

// Extra generated rules so the list has several pages to page through
const GENERATED_ACTIONS = [
  ActionType.ACTION_TYPE_BLOCK,
  ActionType.ACTION_TYPE_LOG,
  ActionType.ACTION_TYPE_REDIRECT,
  ActionType.ACTION_TYPE_SET_TAG,
];

const generatedExpr = (id: string, path: string): Expression =>
  ({
    orCondition: [
      {
        rules: [
          {
            id: `${id}-r`,
            expr: '',
            condition: {
              id: `${id}-c`,
              source: FieldSource.FIELD_SOURCE_PATH,
              key: '',
              operator: Operator.OPERATOR_PREFIX,
              value: path,
            },
          },
        ],
        orCondition: [],
      },
    ],
  }) as Expression;

const generateRules = (prefix: string, from: number, count: number): SecRule[] =>
  Array.from({ length: count }, (_, i) => {
    const n = from + i;
    const id = `${prefix}-${n}`;
    return sampleRule(
      id,
      `Sample rule ${n}`,
      `Sample: path prefix /sample/${n}`,
      [10, 50, 100][n % 3] ?? 50,
      n % 5 === 0 ? Status.STATUS_DISABLE : Status.STATUS_ACTIVE,
      generatedExpr(id, `/sample/${n}`),
      GENERATED_ACTIONS[n % GENERATED_ACTIONS.length] ?? ActionType.ACTION_TYPE_BLOCK,
    );
  });

let sampleStore: SecRule[] = [
  ...SAMPLE_RULES,
  ...generateRules('sample-rule', SAMPLE_RULES.length + 1, 20),
];

const sampleRateLimit = (
  rule: SecRule,
  timeWindow: string,
  maxRequests: number,
  timeout: string,
): RateLimit => ({ ingressRuntime: rule.ingressRuntime, timeWindow, maxRequests, timeout });

export const SAMPLE_RATE_LIMITS: RateLimit[] = [
  sampleRateLimit(
    sampleRule(
      'sample-ratelimit-1',
      'Throttle login attempts',
      'Sample: POST /login, 5 requests per minute, blocked for 10m',
      1,
      Status.STATUS_ACTIVE,
      {
        orCondition: [
          {
            rules: [
              cond(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_EQ, '/login'),
              cond(FieldSource.FIELD_SOURCE_METHOD, Operator.OPERATOR_EQ, 'POST'),
            ],
            orCondition: [],
          },
        ],
      } as Expression,
    ),
    '1m',
    5,
    '10m',
  ),
  sampleRateLimit(
    sampleRule(
      'sample-ratelimit-2',
      'API burst protection',
      'Sample: path prefix /api, 100 requests per 10s',
      10,
      Status.STATUS_ACTIVE,
      {
        orCondition: [
          {
            rules: [cond(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/api')],
            orCondition: [],
          },
        ],
      } as Expression,
    ),
    '10s',
    100,
    '30s',
  ),
  sampleRateLimit(
    sampleRule(
      'sample-ratelimit-3',
      'Search endpoint (log only)',
      'Sample: /search OR /autocomplete, 50 per second, disabled',
      50,
      Status.STATUS_DISABLE,
      {
        orCondition: [
          {
            rules: [cond(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/search')],
            orCondition: [],
          },
          {
            rules: [cond(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/autocomplete')],
            orCondition: [],
          },
        ],
      } as Expression,
      ActionType.ACTION_TYPE_LOG,
    ),
    '1s',
    50,
    '',
  ),
];

let sampleRateLimitStore: RateLimit[] = [
  ...SAMPLE_RATE_LIMITS,
  ...generateRules('sample-ratelimit', SAMPLE_RATE_LIMITS.length + 1, 12).map((r, i) =>
    sampleRateLimit(r, ['10s', '1m', '5m'][i % 3] ?? '1m', 50 * (i + 1), i % 2 ? '10m' : ''),
  ),
];

function withFallback<T>(label: string, fallback: () => T) {
  return async (call: () => Promise<T>): Promise<T> => {
    try {
      return await call();
    } catch (err: any) {
      if (!USE_SAMPLE || axios.isCancel(err)) throw err;
      console.warn(`[security] ${label} failed, using sample response:`, err?.message);
      return fallback();
    }
  };
}

// GET /security/secrules
export async function listSecRules(
  params?: ListSecRulesParams,
  options?: ApiOptions,
): Promise<SecRule[]> {
  const resp = await listSecRulesWithTotal(params, options);
  return resp.secRules;
}

export async function listSecRulesWithTotal(
  params?: ListSecRulesParams,
  options?: ApiOptions,
): Promise<ListSecRulesResult> {
  return withFallback('listSecRules', () => {
    const list = params?.ids?.length
      ? sampleStore.filter((r) => params.ids!.includes(r.ingressRuntime?.id ?? ''))
      : sampleStore;
    return { secRules: paginate(list, params?.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/secrules`, {
      params: pageQuery(params),
      // repeat `ids` key (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    const list: SecuritySecRule[] = resp.data?.secRules ?? [];
    return { secRules: list.map(fromWire), total: Number(resp.data?.total ?? list.length) };
  });
}

// GET /security/secrule/{id}
export async function getSecRule(id: string, options?: ApiOptions): Promise<SecRule> {
  return withFallback(
    'getSecRule',
    () => sampleStore.find((r) => r.ingressRuntime?.id === id) ?? sampleStore[0]!,
  )(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/secrule/${encodeURIComponent(id)}`, {
      signal: options?.signal,
    });
    return fromWire(resp.data?.secRule ?? {});
  });
}

// POST /security/secrule
export async function createSecRule(data: SecRule, options?: ApiOptions): Promise<string> {
  return withFallback('createSecRule', () => {
    const id = `sample-rule-${Date.now()}`;
    sampleStore = [...sampleStore, { ingressRuntime: { ...data.ingressRuntime!, id } } as SecRule];
    return id;
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/secrule`,
      { secRule: toWire(data) },
      { signal: options?.signal },
    );
    return resp.data?.id ?? '';
  });
}

// PUT /security/secrule/{id}
export async function updateSecRule(
  id: string,
  data: SecRule,
  updateMask?: string,
  options?: ApiOptions,
): Promise<SecRule> {
  const next = { ingressRuntime: { ...data.ingressRuntime!, id } } as SecRule;
  return withFallback('updateSecRule', () => {
    sampleStore = sampleStore.map((r) => (r.ingressRuntime?.id === id ? next : r));
    return next;
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/secrule/${encodeURIComponent(id)}`,
      { secRule: toWire(next), updateMask },
      { signal: options?.signal },
    );
    return fromWire(resp.data?.secRule ?? {});
  });
}

// DELETE /security/secrule/{id}
export async function deleteSecRule(id: string, options?: ApiOptions): Promise<void> {
  return withFallback<void>('deleteSecRule', () => {
    sampleStore = sampleStore.filter((r) => r.ingressRuntime?.id !== id);
  })(async () => {
    await axios.delete(`${API_BASE_PATH}/security/secrule/${encodeURIComponent(id)}`, {
      signal: options?.signal,
    });
  });
}

// GET /security/ratelimits
export async function listRateLimits(
  params?: ListRateLimitsParams,
  options?: ApiOptions,
): Promise<RateLimit[]> {
  const resp = await listRateLimitsWithTotal(params, options);
  return resp.rateLimits;
}

export async function listRateLimitsWithTotal(
  params?: ListRateLimitsParams,
  options?: ApiOptions,
): Promise<ListRateLimitsResult> {
  return withFallback('listRateLimits', () => {
    const list = params?.ids?.length
      ? sampleRateLimitStore.filter((r) => params.ids!.includes(r.ingressRuntime?.id ?? ''))
      : sampleRateLimitStore;
    return { rateLimits: paginate(list, params?.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/ratelimits`, {
      params: pageQuery(params),
      // repeat `ids` key (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    const list: SecurityRateLimit[] = resp.data?.rateLimits ?? [];
    return {
      rateLimits: list.map(rateLimitFromWire),
      total: Number(resp.data?.total ?? list.length),
    };
  });
}

// GET /security/ratelimit/{id}
export async function getRateLimit(id: string, options?: ApiOptions): Promise<RateLimit> {
  return withFallback(
    'getRateLimit',
    () => sampleRateLimitStore.find((r) => r.ingressRuntime?.id === id) ?? sampleRateLimitStore[0]!,
  )(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/ratelimit/${encodeURIComponent(id)}`, {
      signal: options?.signal,
    });
    return rateLimitFromWire(resp.data?.rateLimit ?? {});
  });
}

// POST /security/ratelimit
export async function createRateLimit(data: RateLimit, options?: ApiOptions): Promise<string> {
  return withFallback('createRateLimit', () => {
    const id = `sample-ratelimit-${Date.now()}`;
    sampleRateLimitStore = [
      ...sampleRateLimitStore,
      { ...data, ingressRuntime: { ...data.ingressRuntime!, id } },
    ];
    return id;
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/ratelimit`,
      { rateLimit: rateLimitToWire(data) },
      { signal: options?.signal },
    );
    return resp.data?.id ?? '';
  });
}

// PUT /security/ratelimit/{id}
export async function updateRateLimit(
  id: string,
  data: RateLimit,
  updateMask?: string,
  options?: ApiOptions,
): Promise<RateLimit> {
  const next: RateLimit = { ...data, ingressRuntime: { ...data.ingressRuntime!, id } };
  return withFallback('updateRateLimit', () => {
    sampleRateLimitStore = sampleRateLimitStore.map((r) =>
      r.ingressRuntime?.id === id ? next : r,
    );
    return next;
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/ratelimit/${encodeURIComponent(id)}`,
      { rateLimit: rateLimitToWire(next), updateMask },
      { signal: options?.signal },
    );
    return rateLimitFromWire(resp.data?.rateLimit ?? {});
  });
}

// DELETE /security/ratelimit/{id}
export async function deleteRateLimit(id: string, options?: ApiOptions): Promise<void> {
  return withFallback<void>('deleteRateLimit', () => {
    sampleRateLimitStore = sampleRateLimitStore.filter((r) => r.ingressRuntime?.id !== id);
  })(async () => {
    await axios.delete(`${API_BASE_PATH}/security/ratelimit/${encodeURIComponent(id)}`, {
      signal: options?.signal,
    });
  });
}

// GET /security/status
export async function getSecurityStatus(options?: ApiOptions): Promise<StatusResponse> {
  return withFallback<StatusResponse>('getSecurityStatus', () => ({
    msg: 'sample: security service OK',
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/status`, { signal: options?.signal });
    return resp.data ?? {};
  });
}
