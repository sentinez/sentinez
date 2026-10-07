import { SecRule } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
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
// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

export interface Pages {
  index?: number;
  size?: number;
  total?: boolean;
}

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

export interface ListSecRulesParams {
  page?: Pages;
  ids?: string[];
}

export interface ListSecRulesResult {
  secRules: SecRule[];
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

let sampleStore: SecRule[] = [...SAMPLE_RULES];

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
    return { secRules: list, total: list.length };
  })(async () => {
    const query: Record<string, unknown> = {};
    if (params?.page?.index !== undefined) query['page.index'] = params.page.index;
    if (params?.page?.size !== undefined) query['page.size'] = params.page.size;
    if (params?.page?.total !== undefined) query['page.total'] = params.page.total;
    if (params?.ids?.length) query.ids = params.ids;

    const resp = await axios.get(`${API_BASE_PATH}/security/secrules`, {
      params: query,
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
    sampleStore = [
      ...sampleStore,
      { ingressRuntime: { ...data.ingressRuntime!, id } } as SecRule,
    ];
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

// GET /security/status
export async function getSecurityStatus(options?: ApiOptions): Promise<StatusResponse> {
  return withFallback<StatusResponse>('getSecurityStatus', () => ({
    msg: 'sample: security service OK',
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/status`, { signal: options?.signal });
    return resp.data ?? {};
  });
}
