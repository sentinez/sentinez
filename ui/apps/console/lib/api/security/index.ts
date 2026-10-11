import { CoreRuleset, RateLimit, SecRule } from '@sentinez/proto/sentinez/apps/security/v1/model';
import {
  CreateCoreRulesetRequest,
  CreateCoreRulesetResponse,
  CreateRateLimitRequest,
  CreateRateLimitResponse,
  CreateSecRuleRequest,
  CreateSecRuleResponse,
  DeleteCoreRulesetRequest,
  DeleteCoreRulesetResponse,
  DeleteRateLimitRequest,
  DeleteRateLimitResponse,
  DeleteSecRuleRequest,
  DeleteSecRuleResponse,
  GetCoreRulesetRequest,
  GetCoreRulesetResponse,
  GetRateLimitRequest,
  GetRateLimitResponse,
  GetRuleInfoRequest,
  GetRuleInfoResponse,
  GetSecRuleRequest,
  GetSecRuleResponse,
  ListCoreRulesetsRequest,
  ListCoreRulesetsResponse,
  ListRateLimitsRequest,
  ListRateLimitsResponse,
  ListRuleInfosRequest,
  ListRuleInfosResponse,
  ListRuleTagsRequest,
  ListRuleTagsResponse,
  ListSecRulesRequest,
  ListSecRulesResponse,
  StatusResponse,
  UpdateCoreRulesetRequest,
  UpdateCoreRulesetResponse,
  UpdateRateLimitRequest,
  UpdateRateLimitResponse,
  UpdateSecRuleRequest,
  UpdateSecRuleResponse,
} from '@sentinez/proto/sentinez/apps/security/v1/security';
import { RuleInfo } from '@sentinez/proto/sentinez/types/coreruleset/v1/setting';
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

export const SAMPLE_CORE_RULESETS: CoreRuleset[] = [
  {
    id: 'sample-coreruleset-1',
    name: 'Default protection',
    description: 'Sample: blocking, paranoia level 1, every rule group',
    content: {
      mode: 'ENGINE_MODE_ON',
      version: 'VERSION_V4_17_0',
      paranoiaLevel: 1,
      threshold: { inbound: 5, outbound: 4 },
      request: { bodyAccess: true, allowedMethods: ['GET', 'HEAD', 'POST', 'OPTIONS'] },
      status: 'STATUS_ACTIVE',
    },
  },
  {
    id: 'sample-coreruleset-2',
    name: 'Strict API',
    description: 'Sample: paranoia level 3, SQLi/XSS/RCE only, tuned exclusions and overrides',
    content: {
      mode: 'ENGINE_MODE_ON',
      version: 'VERSION_V4_17_0',
      paranoiaLevel: 3,
      detectionParanoiaLevel: 4,
      threshold: { inbound: 5, outbound: 4, earlyBlocking: true },
      categories: ['CATEGORY_RCE', 'CATEGORY_XSS', 'CATEGORY_SQLI'],
      request: {
        bodyAccess: true,
        bodyLimitBytes: '1048576',
        allowedMethods: ['GET', 'POST', 'PUT', 'DELETE'],
        allowedContentTypes: ['application/json'],
      },
      exclusions: [
        { tags: ['attack-protocol'] },
        { targets: [{ ruleId: 942100, variable: 'ARGS:password' }] },
      ],
      overrides: [
        { id: 942100, state: 'RULE_STATE_DETECTION_ONLY' },
        { id: 941100, state: 'RULE_STATE_DISABLED' },
      ],
      status: 'STATUS_ACTIVE',
    },
  },
  {
    id: 'sample-coreruleset-3',
    name: 'Staging (detection only)',
    description: 'Sample: logs matches without blocking, disabled',
    content: {
      mode: 'ENGINE_MODE_DETECTION_ONLY',
      version: 'VERSION_V4_16_0',
      paranoiaLevel: 2,
      status: 'STATUS_DISABLE',
    },
  },
].map(CoreRuleset.fromJSON);

let sampleCoreRulesetStore: CoreRuleset[] = [...SAMPLE_CORE_RULESETS];

export const SAMPLE_RULE_INFOS: RuleInfo[] = [
  {
    id: 901100,
    category: 'CATEGORY_UNSPECIFIED',
    msg: 'Initialization',
    system: true,
  },
  {
    id: 911100,
    category: 'CATEGORY_METHOD_ENFORCEMENT',
    paranoiaLevel: 1,
    severity: 'CRITICAL',
    msg: 'Method is not allowed by policy',
    tags: ['attack-generic', 'paranoia-level/1', 'OWASP_CRS'],
  },
  {
    id: 920350,
    category: 'CATEGORY_PROTOCOL_ENFORCEMENT',
    paranoiaLevel: 1,
    severity: 'WARNING',
    msg: 'Host header is a numeric IP address',
    tags: ['attack-protocol', 'paranoia-level/1', 'OWASP_CRS'],
  },
  {
    id: 930120,
    category: 'CATEGORY_LFI',
    paranoiaLevel: 1,
    severity: 'CRITICAL',
    msg: 'OS File Access Attempt',
    tags: ['attack-lfi', 'paranoia-level/1', 'OWASP_CRS'],
  },
  {
    id: 932100,
    category: 'CATEGORY_RCE',
    paranoiaLevel: 1,
    severity: 'CRITICAL',
    msg: 'Remote Command Execution: Unix Command Injection',
    tags: ['attack-rce', 'paranoia-level/1', 'OWASP_CRS'],
  },
  {
    id: 941100,
    category: 'CATEGORY_XSS',
    paranoiaLevel: 1,
    severity: 'CRITICAL',
    msg: 'XSS Attack Detected via libinjection',
    tags: ['attack-xss', 'paranoia-level/1', 'OWASP_CRS'],
  },
  {
    id: 942100,
    category: 'CATEGORY_SQLI',
    paranoiaLevel: 1,
    severity: 'CRITICAL',
    msg: 'SQL Injection Attack Detected via libinjection',
    tags: ['attack-sqli', 'paranoia-level/1', 'OWASP_CRS'],
  },
  {
    id: 942430,
    category: 'CATEGORY_SQLI',
    paranoiaLevel: 2,
    severity: 'WARNING',
    msg: 'Restricted SQL Character Anomaly Detection (args): # of special characters exceeded (12)',
    tags: ['attack-sqli', 'paranoia-level/2', 'OWASP_CRS'],
  },
  {
    id: 951110,
    category: 'CATEGORY_DATA_LEAKAGES_SQL',
    paranoiaLevel: 1,
    severity: 'CRITICAL',
    msg: 'Microsoft Access SQL Information Leakage',
    tags: ['attack-disclosure', 'paranoia-level/1', 'OWASP_CRS'],
  },
].map(RuleInfo.fromJSON);

const anyOf = <T>(filter: T[] | undefined, value: T) => !filter?.length || filter.includes(value);

const includesText = (text: string, part?: string) =>
  !part || text.toLowerCase().includes(part.toLowerCase());

const filterCoreRulesets = (req: Partial<ListCoreRulesetsRequest>) =>
  byIds(sampleCoreRulesetStore, req.ids).filter(
    (r) =>
      includesText(r.name, req.name) &&
      anyOf(req.modes, r.content?.mode) &&
      anyOf(req.versions, r.content?.version) &&
      anyOf(req.statuses, r.content?.status) &&
      (!req.paranoiaLevel || r.content?.paranoiaLevel === req.paranoiaLevel),
  );

const filterRuleInfos = (req: Partial<ListRuleInfosRequest>) =>
  SAMPLE_RULE_INFOS.filter(
    (r) =>
      (req.includeSystem || !r.system) &&
      anyOf(req.ids, r.id) &&
      anyOf(req.categories, r.category) &&
      anyOf(req.paranoiaLevels, r.paranoiaLevel) &&
      anyOf(
        req.severities?.map((s) => s.toUpperCase()),
        r.severity,
      ) &&
      (req.tags ?? []).every((tag) => r.tags.includes(tag)) &&
      (includesText(r.msg, req.search) ||
        r.tags.some((tag) => includesText(tag, req.search)) ||
        String(r.id).startsWith(req.search ?? '')),
  );

const sampleRuleTags = (req: Partial<ListRuleTagsRequest>) => {
  const counts = new Map<string, number>();
  for (const rule of SAMPLE_RULE_INFOS) {
    if (rule.system && !req.includeSystem) continue;
    for (const tag of rule.tags) {
      if (includesText(tag, req.search)) counts.set(tag, (counts.get(tag) ?? 0) + 1);
    }
  }
  return [...counts]
    .map(([name, ruleCount]) => ({ name, ruleCount }))
    .sort((a, b) => (a.name < b.name ? -1 : 1));
};

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

// GET /security/corerulesets
export async function listCoreRulesets(
  req: Partial<ListCoreRulesetsRequest> = {},
  options?: ApiOptions,
): Promise<ListCoreRulesetsResponse> {
  return withFallback('listCoreRulesets', () => {
    const list = filterCoreRulesets(req);
    return { coreRulesets: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/corerulesets`, {
      params: toQuery(ListCoreRulesetsRequest.toJSON(ListCoreRulesetsRequest.fromPartial(req))),
      // repeat the keys of repeated filters (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListCoreRulesetsResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/coreruleset/{id}
export async function getCoreRuleset(
  req: GetCoreRulesetRequest,
  options?: ApiOptions,
): Promise<GetCoreRulesetResponse> {
  return withFallback<GetCoreRulesetResponse>('getCoreRuleset', () => ({
    coreRuleset: sampleCoreRulesetStore.find((r) => r.id === req.id) ?? sampleCoreRulesetStore[0],
  }))(async () => {
    const endpoint = `${API_BASE_PATH}/security/coreruleset/${encodeURIComponent(req.id)}`;
    const resp = await axios.get(endpoint, { signal: options?.signal });
    return GetCoreRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// POST /security/coreruleset
export async function createCoreRuleset(
  req: CreateCoreRulesetRequest,
  options?: ApiOptions,
): Promise<CreateCoreRulesetResponse> {
  return withFallback('createCoreRuleset', () => {
    const id = `sample-coreruleset-${Date.now()}`;
    sampleCoreRulesetStore = [
      ...sampleCoreRulesetStore,
      CoreRuleset.fromPartial({ ...req.coreRuleset, id }),
    ];
    return { id };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/coreruleset`,
      CreateCoreRulesetRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateCoreRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /security/coreruleset/{id}
export async function updateCoreRuleset(
  req: UpdateCoreRulesetRequest,
  options?: ApiOptions,
): Promise<UpdateCoreRulesetResponse> {
  return withFallback<UpdateCoreRulesetResponse>('updateCoreRuleset', () => {
    const next = CoreRuleset.fromPartial({ ...req.coreRuleset, id: req.id });
    sampleCoreRulesetStore = sampleCoreRulesetStore.map((r) => (r.id === req.id ? next : r));
    return { coreRuleset: next };
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/coreruleset/${encodeURIComponent(req.id)}`,
      UpdateCoreRulesetRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateCoreRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /security/coreruleset/{id}
export async function deleteCoreRuleset(
  req: DeleteCoreRulesetRequest,
  options?: ApiOptions,
): Promise<DeleteCoreRulesetResponse> {
  return withFallback<DeleteCoreRulesetResponse>('deleteCoreRuleset', () => {
    sampleCoreRulesetStore = sampleCoreRulesetStore.filter((r) => r.id !== req.id);
    return {};
  })(async () => {
    const endpoint = `${API_BASE_PATH}/security/coreruleset/${encodeURIComponent(req.id)}`;
    const resp = await axios.delete(endpoint, { signal: options?.signal });
    return DeleteCoreRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/crs/rules
export async function listRuleInfos(
  req: Partial<ListRuleInfosRequest> = {},
  options?: ApiOptions,
): Promise<ListRuleInfosResponse> {
  return withFallback('listRuleInfos', () => {
    const list = filterRuleInfos(req);
    return { rules: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/crs/rules`, {
      params: toQuery(ListRuleInfosRequest.toJSON(ListRuleInfosRequest.fromPartial(req))),
      // repeat the keys of repeated filters (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListRuleInfosResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/crs/tags
export async function listRuleTags(
  req: Partial<ListRuleTagsRequest> = {},
  options?: ApiOptions,
): Promise<ListRuleTagsResponse> {
  return withFallback('listRuleTags', () => {
    const list = sampleRuleTags(req);
    return { tags: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/crs/tags`, {
      params: toQuery(ListRuleTagsRequest.toJSON(ListRuleTagsRequest.fromPartial(req))),
      signal: options?.signal,
    });
    return ListRuleTagsResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/crs/rules/{id}
export async function getRuleInfo(
  req: Partial<GetRuleInfoRequest> & Pick<GetRuleInfoRequest, 'id'>,
  options?: ApiOptions,
): Promise<GetRuleInfoResponse> {
  return withFallback<GetRuleInfoResponse>('getRuleInfo', () => ({
    rule: SAMPLE_RULE_INFOS.find((r) => r.id === req.id),
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/crs/rules/${req.id}`, {
      // `id` is the path param: zero it so toJSON leaves it out of the query
      params: toQuery(GetRuleInfoRequest.toJSON(GetRuleInfoRequest.fromPartial({ ...req, id: 0 }))),
      signal: options?.signal,
    });
    return GetRuleInfoResponse.fromJSON(resp.data ?? {});
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
