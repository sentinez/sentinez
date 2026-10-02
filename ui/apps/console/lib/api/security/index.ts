import { RuleBased } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import {
  ActionType,
  FieldSource,
  Operator,
  actionTypeFromJSON,
  actionTypeToJSON,
  Expression,
} from '@sentinez/proto/sentinez/secure/rule/v1/engine';
import { Status, statusFromJSON, statusToJSON } from '@sentinez/proto/sentinez/types/v1/known';
import axios from 'axios';

const API_BASE_PATH = process.env.NEXT_PUBLIC_BASE_PATH || 'http://localhost:8080';
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

/** Wire format of `v1RuleBased` in security.swagger.json */
export interface SecurityRuleBased {
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

export interface ListRuleBasedsParams {
  page?: Pages;
  ids?: string[];
}

export interface ListRuleBasedsResult {
  ruleBaseds: RuleBased[];
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

function fromWire(rule: SecurityRuleBased): RuleBased {
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
  } as RuleBased;
}

function toWire(data: RuleBased): SecurityRuleBased {
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
): RuleBased =>
  ({
    ingressRuntime: { id, name, description, priority, status, expr, action: { type: action } },
  }) as RuleBased;

const cond = (source: FieldSource, operator: Operator, value: unknown, key = '') => ({
  id: Math.random().toString(36).slice(2, 9),
  expr: '',
  condition: { id: Math.random().toString(36).slice(2, 9), source, key, operator, value },
});

export const SAMPLE_RULES: RuleBased[] = [
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

let sampleStore: RuleBased[] = [...SAMPLE_RULES];

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

// GET /security/rulebaseds
export async function listRuleBaseds(
  params?: ListRuleBasedsParams,
  options?: ApiOptions,
): Promise<RuleBased[]> {
  const resp = await listRuleBasedsWithTotal(params, options);
  return resp.ruleBaseds;
}

export async function listRuleBasedsWithTotal(
  params?: ListRuleBasedsParams,
  options?: ApiOptions,
): Promise<ListRuleBasedsResult> {
  return withFallback('listRuleBaseds', () => {
    const list = params?.ids?.length
      ? sampleStore.filter((r) => params.ids!.includes(r.ingressRuntime?.id ?? ''))
      : sampleStore;
    return { ruleBaseds: list, total: list.length };
  })(async () => {
    const query: Record<string, unknown> = {};
    if (params?.page?.index !== undefined) query['page.index'] = params.page.index;
    if (params?.page?.size !== undefined) query['page.size'] = params.page.size;
    if (params?.page?.total !== undefined) query['page.total'] = params.page.total;
    if (params?.ids?.length) query.ids = params.ids;

    const resp = await axios.get(`${API_BASE_PATH}/security/rulebaseds`, {
      params: query,
      // repeat `ids` key (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    const list: SecurityRuleBased[] = resp.data?.ruleBaseds ?? [];
    return { ruleBaseds: list.map(fromWire), total: Number(resp.data?.total ?? list.length) };
  });
}

// GET /security/rulebased/{id}
export async function getRuleBased(id: string, options?: ApiOptions): Promise<RuleBased> {
  return withFallback(
    'getRuleBased',
    () => sampleStore.find((r) => r.ingressRuntime?.id === id) ?? sampleStore[0]!,
  )(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/rulebased/${encodeURIComponent(id)}`, {
      signal: options?.signal,
    });
    return fromWire(resp.data?.ruleBased ?? {});
  });
}

// POST /security/rulebased
export async function createRuleBased(data: RuleBased, options?: ApiOptions): Promise<string> {
  return withFallback('createRuleBased', () => {
    const id = `sample-rule-${Date.now()}`;
    sampleStore = [
      ...sampleStore,
      { ingressRuntime: { ...data.ingressRuntime!, id } } as RuleBased,
    ];
    return id;
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/rulebased`,
      { ruleBased: toWire(data) },
      { signal: options?.signal },
    );
    return resp.data?.id ?? '';
  });
}

// PUT /security/rulebased/{id}
export async function updateRuleBased(
  id: string,
  data: RuleBased,
  updateMask?: string,
  options?: ApiOptions,
): Promise<RuleBased> {
  const next = { ingressRuntime: { ...data.ingressRuntime!, id } } as RuleBased;
  return withFallback('updateRuleBased', () => {
    sampleStore = sampleStore.map((r) => (r.ingressRuntime?.id === id ? next : r));
    return next;
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/rulebased/${encodeURIComponent(id)}`,
      { ruleBased: toWire(next), updateMask },
      { signal: options?.signal },
    );
    return fromWire(resp.data?.ruleBased ?? {});
  });
}

// DELETE /security/rulebased/{id}
export async function deleteRuleBased(id: string, options?: ApiOptions): Promise<void> {
  return withFallback<void>('deleteRuleBased', () => {
    sampleStore = sampleStore.filter((r) => r.ingressRuntime?.id !== id);
  })(async () => {
    await axios.delete(`${API_BASE_PATH}/security/rulebased/${encodeURIComponent(id)}`, {
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
