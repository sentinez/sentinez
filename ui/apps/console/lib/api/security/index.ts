import { RateLimit, SecRule } from '@sentinez/proto/sentinez/apps/security/v1/model';
import {
  CreateRateLimitRequest,
  CreateRateLimitResponse,
  CreateSecRuleRequest,
  CreateSecRuleResponse,
  DeleteRateLimitRequest,
  DeleteRateLimitResponse,
  DeleteSecRuleRequest,
  DeleteSecRuleResponse,
  GetRateLimitRequest,
  GetRateLimitResponse,
  GetSecRuleRequest,
  GetSecRuleResponse,
  ListRateLimitsRequest,
  ListRateLimitsResponse,
  ListSecRulesRequest,
  ListSecRulesResponse,
  StatusResponse,
  UpdateRateLimitRequest,
  UpdateRateLimitResponse,
  UpdateSecRuleRequest,
  UpdateSecRuleResponse,
} from '@sentinez/proto/sentinez/apps/security/v1/security';
import {
  ActionType,
  FieldSource,
  Operator,
  actionTypeToJSON,
  Expression,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { type Pages, paginate, toQuery } from '@/lib/api/pages';
// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

export type { Pages };

// ---- Sample data, used as fallback when the API call fails ----

const sampleRule = (
  id: string,
  name: string,
  description: string,
  priority: number,
  status: Status,
  expr: Expression,
  action: ActionType = ActionType.ACTION_TYPE_BLOCK,
): SecRule => ({
  id,
  name,
  description,
  priority,
  status,
  expr,
  action: actionTypeToJSON(action),
});

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
): RateLimit => ({ ...rule, timeWindow, maxRequests, timeout });

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

const byIds = <T extends { id: string }>(list: T[], ids?: string[]) =>
  ids?.length ? list.filter((r) => ids.includes(r.id)) : list;

// GET /security/secrules
export async function listSecRules(
  req: Partial<ListSecRulesRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRulesResponse> {
  return withFallback('listSecRules', () => {
    const list = byIds(sampleStore, req.ids);
    return { secRules: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/secrules`, {
      params: toQuery(ListSecRulesRequest.toJSON(ListSecRulesRequest.fromPartial(req))),
      // repeat `ids` key (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListSecRulesResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/secrule/{id}
export async function getSecRule(
  req: GetSecRuleRequest,
  options?: ApiOptions,
): Promise<GetSecRuleResponse> {
  return withFallback<GetSecRuleResponse>('getSecRule', () => ({
    secRule: sampleStore.find((r) => r.id === req.id) ?? sampleStore[0],
  }))(async () => {
    const endpoint = `${API_BASE_PATH}/security/secrule/${encodeURIComponent(req.id)}`;
    const resp = await axios.get(endpoint, { signal: options?.signal });
    return GetSecRuleResponse.fromJSON(resp.data ?? {});
  });
}

// POST /security/secrule
export async function createSecRule(
  req: CreateSecRuleRequest,
  options?: ApiOptions,
): Promise<CreateSecRuleResponse> {
  return withFallback('createSecRule', () => {
    const id = `sample-rule-${Date.now()}`;
    sampleStore = [...sampleStore, SecRule.fromPartial({ ...req.secRule, id })];
    return { id };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/secrule`,
      CreateSecRuleRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateSecRuleResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /security/secrule/{id}
export async function updateSecRule(
  req: UpdateSecRuleRequest,
  options?: ApiOptions,
): Promise<UpdateSecRuleResponse> {
  return withFallback<UpdateSecRuleResponse>('updateSecRule', () => {
    const next = SecRule.fromPartial({ ...req.secRule, id: req.id });
    sampleStore = sampleStore.map((r) => (r.id === req.id ? next : r));
    return { secRule: next };
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/secrule/${encodeURIComponent(req.id)}`,
      UpdateSecRuleRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateSecRuleResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /security/secrule/{id}
export async function deleteSecRule(
  req: DeleteSecRuleRequest,
  options?: ApiOptions,
): Promise<DeleteSecRuleResponse> {
  return withFallback<DeleteSecRuleResponse>('deleteSecRule', () => {
    sampleStore = sampleStore.filter((r) => r.id !== req.id);
    return {};
  })(async () => {
    const endpoint = `${API_BASE_PATH}/security/secrule/${encodeURIComponent(req.id)}`;
    const resp = await axios.delete(endpoint, { signal: options?.signal });
    return DeleteSecRuleResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/ratelimits
export async function listRateLimits(
  req: Partial<ListRateLimitsRequest> = {},
  options?: ApiOptions,
): Promise<ListRateLimitsResponse> {
  return withFallback('listRateLimits', () => {
    const list = byIds(sampleRateLimitStore, req.ids);
    return { rateLimits: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/ratelimits`, {
      params: toQuery(ListRateLimitsRequest.toJSON(ListRateLimitsRequest.fromPartial(req))),
      // repeat `ids` key (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListRateLimitsResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/ratelimit/{id}
export async function getRateLimit(
  req: GetRateLimitRequest,
  options?: ApiOptions,
): Promise<GetRateLimitResponse> {
  return withFallback<GetRateLimitResponse>('getRateLimit', () => ({
    rateLimit: sampleRateLimitStore.find((r) => r.id === req.id) ?? sampleRateLimitStore[0],
  }))(async () => {
    const endpoint = `${API_BASE_PATH}/security/ratelimit/${encodeURIComponent(req.id)}`;
    const resp = await axios.get(endpoint, { signal: options?.signal });
    return GetRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// POST /security/ratelimit
export async function createRateLimit(
  req: CreateRateLimitRequest,
  options?: ApiOptions,
): Promise<CreateRateLimitResponse> {
  return withFallback('createRateLimit', () => {
    const id = `sample-ratelimit-${Date.now()}`;
    sampleRateLimitStore = [
      ...sampleRateLimitStore,
      RateLimit.fromPartial({ ...req.rateLimit, id }),
    ];
    return { id };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/ratelimit`,
      CreateRateLimitRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /security/ratelimit/{id}
export async function updateRateLimit(
  req: UpdateRateLimitRequest,
  options?: ApiOptions,
): Promise<UpdateRateLimitResponse> {
  return withFallback<UpdateRateLimitResponse>('updateRateLimit', () => {
    const next = RateLimit.fromPartial({ ...req.rateLimit, id: req.id });
    sampleRateLimitStore = sampleRateLimitStore.map((r) => (r.id === req.id ? next : r));
    return { rateLimit: next };
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/ratelimit/${encodeURIComponent(req.id)}`,
      UpdateRateLimitRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /security/ratelimit/{id}
export async function deleteRateLimit(
  req: DeleteRateLimitRequest,
  options?: ApiOptions,
): Promise<DeleteRateLimitResponse> {
  return withFallback<DeleteRateLimitResponse>('deleteRateLimit', () => {
    sampleRateLimitStore = sampleRateLimitStore.filter((r) => r.id !== req.id);
    return {};
  })(async () => {
    const endpoint = `${API_BASE_PATH}/security/ratelimit/${encodeURIComponent(req.id)}`;
    const resp = await axios.delete(endpoint, { signal: options?.signal });
    return DeleteRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/status
export async function getSecurityStatus(options?: ApiOptions): Promise<StatusResponse> {
  return withFallback<StatusResponse>('getSecurityStatus', () => ({
    msg: 'sample: security service OK',
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/status`, { signal: options?.signal });
    return StatusResponse.fromJSON(resp.data ?? {});
  });
}
