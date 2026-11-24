import { SecRuleset, SecRateLimit, SecRule } from '@sentinez/proto/sentinez/apps/security/v1/model';
import {
  CreateSecRulesetRequest,
  CreateSecRulesetResponse,
  CreateSecRateLimitRequest,
  CreateSecRateLimitResponse,
  CreateSecRuleRequest,
  CreateSecRuleResponse,
  DeleteSecRulesetRequest,
  DeleteSecRulesetResponse,
  DeleteSecRateLimitRequest,
  DeleteSecRateLimitResponse,
  DeleteSecRuleRequest,
  DeleteSecRuleResponse,
  GetSecRulesetRequest,
  GetSecRulesetResponse,
  GetSecRateLimitRequest,
  GetSecRateLimitResponse,
  GetCRSRuleInfoRequest,
  GetCRSRuleInfoResponse,
  GetConditionSchemaResponse,
  GetSecRuleRequest,
  GetSecRuleResponse,
  ListSecRulesetsRequest,
  ListSecRulesetsResponse,
  ListSecRateLimitsRequest,
  ListSecRateLimitsResponse,
  ListCRSRuleInfosRequest,
  ListCRSRuleInfosResponse,
  ListCRSRuleTagsRequest,
  ListCRSRuleTagsResponse,
  ListSecRulesRequest,
  ListSecRulesResponse,
  StatusResponse,
  UpdateSecRulesetRequest,
  UpdateSecRulesetResponse,
  UpdateSecRateLimitRequest,
  UpdateSecRateLimitResponse,
  UpdateSecRuleRequest,
  UpdateSecRuleResponse,
} from '@sentinez/proto/sentinez/apps/security/v1/security';
import { RuleInfo } from '@sentinez/proto/sentinez/types/ruleset/v1/setting';
import { providerOf } from '@/lib/type/security';
import {
  ActionType,
  FieldSource,
  Operator,
  actionTypeToJSON,
  Expression,
  ConditionSchema,
  KeyMode,
  OperatorSpec,
  SourceSpec,
  ValueKind,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import type { Owner } from '@sentinez/proto/sentinez/types/v1/model';
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { type Pages, paginate, toQuery } from '@/lib/api/pages';
import { SAMPLE_RESOURCES, SAMPLE_TENANT_ID } from '@/lib/api/tenant';
// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

export type { Pages };

// ---- Sample data, used as fallback when the API call fails ----

// Sample rows belong to the first sample resources of lib/api/tenant
const SAMPLE_RESOURCE_IDS = SAMPLE_RESOURCES.slice(0, 2).map((r) => r.id);

const ofResource = <T extends { owner?: Owner }>(item: T, n = 0): T => ({
  ...item,
  owner: {
    resourceId: SAMPLE_RESOURCE_IDS[n % SAMPLE_RESOURCE_IDS.length] ?? '',
    tenantId: SAMPLE_TENANT_ID,
  },
});

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
  exprPreview: '',
});

const cond = (source: FieldSource, operator: Operator, value: unknown, key = '') => ({
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
          andCondition: [],
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
          andCondition: [],
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
          andCondition: [],
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
          andCondition: [],
        },
      ],
    } as Expression,
    ActionType.ACTION_TYPE_LOG,
  ),
].map((r) => ofResource(r));

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
        andCondition: [],
      },
    ],
  }) as Expression;

const generateRules = (prefix: string, from: number, count: number): SecRule[] =>
  Array.from({ length: count }, (_, i) => {
    const n = from + i;
    const id = `${prefix}-${n}`;
    return ofResource(
      sampleRule(
        id,
        `Sample rule ${n}`,
        `Sample: path prefix /sample/${n}`,
        [10, 50, 100][n % 3] ?? 50,
        n % 5 === 0 ? Status.STATUS_DISABLE : Status.STATUS_ACTIVE,
        generatedExpr(id, `/sample/${n}`),
        GENERATED_ACTIONS[n % GENERATED_ACTIONS.length] ?? ActionType.ACTION_TYPE_BLOCK,
      ),
      n,
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
): SecRateLimit => ({ ...rule, timeWindow, maxRequests, timeout });

export const SAMPLE_RATE_LIMITS: SecRateLimit[] = [
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
            andCondition: [],
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
            andCondition: [],
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
            andCondition: [],
          },
          {
            rules: [cond(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/autocomplete')],
            orCondition: [],
            andCondition: [],
          },
        ],
      } as Expression,
      ActionType.ACTION_TYPE_LOG,
    ),
    '1s',
    50,
    '',
  ),
].map((r) => ofResource(r));

let sampleRateLimitStore: SecRateLimit[] = [
  ...SAMPLE_RATE_LIMITS,
  ...generateRules('sample-ratelimit', SAMPLE_RATE_LIMITS.length + 1, 12).map((r, i) =>
    sampleRateLimit(r, ['10s', '1m', '5m'][i % 3] ?? '1m', 50 * (i + 1), i % 2 ? '10m' : ''),
  ),
];

export const SAMPLE_RULESETS: SecRuleset[] = [
  {
    id: 'sample-ruleset-1',
    name: 'Default protection',
    description: 'Sample: blocking, paranoia level 1, every rule group',
    setting: {
      status: 'STATUS_ACTIVE',
      provider: 'PROVIDER_OWASP',
      crs: {
        mode: 'ENGINE_MODE_ON',
        version: 'CRS_VERSION_V4_17_0',
        paranoiaLevel: 1,
        threshold: { inbound: 5, outbound: 4 },
        request: { bodyAccess: true, allowedMethods: ['GET', 'HEAD', 'POST', 'OPTIONS'] },
      },
    },
  },
  {
    id: 'sample-ruleset-2',
    name: 'Strict API',
    description: 'Sample: paranoia level 3, SQLi/XSS/RCE only, tuned exclusions and overrides',
    setting: {
      status: 'STATUS_ACTIVE',
      provider: 'PROVIDER_OWASP',
      crs: {
        mode: 'ENGINE_MODE_ON',
        version: 'CRS_VERSION_V4_17_0',
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
      },
    },
  },
  {
    id: 'sample-ruleset-3',
    name: 'Staging (detection only)',
    description: 'Sample: logs matches without blocking, disabled',
    setting: {
      status: 'STATUS_DISABLE',
      provider: 'PROVIDER_OWASP',
      crs: {
        mode: 'ENGINE_MODE_DETECTION_ONLY',
        version: 'CRS_VERSION_V4_16_0',
        paranoiaLevel: 2,
      },
    },
  },
].map((r) => ofResource(SecRuleset.fromJSON(r)));

let sampleRulesetStore: SecRuleset[] = [...SAMPLE_RULESETS];

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

const filterRulesets = (req: Partial<ListSecRulesetsRequest>) =>
  byResources(byIds(sampleRulesetStore, req.ids), req.resourceIds).filter(
    (r) =>
      includesText(r.name, req.name) &&
      (!req.modes?.length ||
        req.modes.some((m) => m === r.setting?.crs?.mode || m === r.setting?.managed?.mode)) &&
      anyOf(req.providers, providerOf(r.setting)) &&
      anyOf(req.versions, r.setting?.crs?.version) &&
      anyOf(req.statuses, r.setting?.status) &&
      (!req.paranoiaLevel || r.setting?.crs?.paranoiaLevel === req.paranoiaLevel),
  );

const filterCRSRuleInfos = (req: Partial<ListCRSRuleInfosRequest>) =>
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

const sampleCRSRuleTags = (req: Partial<ListCRSRuleTagsRequest>) => {
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

const byResources = <T extends { owner?: Owner }>(list: T[], ids?: string[]) =>
  ids?.length ? list.filter((r) => ids.includes(r.owner?.resourceId ?? '')) : list;

// GET /security/secrules
export async function listSecRules(
  req: Partial<ListSecRulesRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRulesResponse> {
  return withFallback('listSecRules', () => {
    const list = byResources(byIds(sampleStore, req.ids), req.resourceIds);
    return { secrules: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/secrules`, {
      params: toQuery(ListSecRulesRequest.toJSON(ListSecRulesRequest.fromPartial(req))),
      // repeat the `ids` / `resourceIds` keys (collectionFormat: multi)
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
    secrule: sampleStore.find((r) => r.id === req.id) ?? sampleStore[0],
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
    sampleStore = [...sampleStore, SecRule.fromPartial({ ...req.secrule, id })];
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
    // The owner of a row never changes
    const owner = sampleStore.find((r) => r.id === req.id)?.owner;
    const next = SecRule.fromPartial({ ...req.secrule, id: req.id, owner });
    sampleStore = sampleStore.map((r) => (r.id === req.id ? next : r));
    return { secrule: next };
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

// GET /security/secratelimits
export async function listSecRateLimits(
  req: Partial<ListSecRateLimitsRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRateLimitsResponse> {
  return withFallback('listSecRateLimits', () => {
    const list = byResources(byIds(sampleRateLimitStore, req.ids), req.resourceIds);
    return { secratelimits: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/secratelimits`, {
      params: toQuery(ListSecRateLimitsRequest.toJSON(ListSecRateLimitsRequest.fromPartial(req))),
      // repeat the `ids` / `resourceIds` keys (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListSecRateLimitsResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/secratelimit/{id}
export async function getSecRateLimit(
  req: GetSecRateLimitRequest,
  options?: ApiOptions,
): Promise<GetSecRateLimitResponse> {
  return withFallback<GetSecRateLimitResponse>('getSecRateLimit', () => ({
    secratelimit: sampleRateLimitStore.find((r) => r.id === req.id) ?? sampleRateLimitStore[0],
  }))(async () => {
    const endpoint = `${API_BASE_PATH}/security/secratelimit/${encodeURIComponent(req.id)}`;
    const resp = await axios.get(endpoint, { signal: options?.signal });
    return GetSecRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// POST /security/secratelimit
export async function createSecRateLimit(
  req: CreateSecRateLimitRequest,
  options?: ApiOptions,
): Promise<CreateSecRateLimitResponse> {
  return withFallback('createSecRateLimit', () => {
    const id = `sample-ratelimit-${Date.now()}`;
    sampleRateLimitStore = [
      ...sampleRateLimitStore,
      SecRateLimit.fromPartial({ ...req.secratelimit, id }),
    ];
    return { id };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/secratelimit`,
      CreateSecRateLimitRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateSecRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /security/secratelimit/{id}
export async function updateSecRateLimit(
  req: UpdateSecRateLimitRequest,
  options?: ApiOptions,
): Promise<UpdateSecRateLimitResponse> {
  return withFallback<UpdateSecRateLimitResponse>('updateSecRateLimit', () => {
    const owner = sampleRateLimitStore.find((r) => r.id === req.id)?.owner;
    const next = SecRateLimit.fromPartial({ ...req.secratelimit, id: req.id, owner });
    sampleRateLimitStore = sampleRateLimitStore.map((r) => (r.id === req.id ? next : r));
    return { secratelimit: next };
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/secratelimit/${encodeURIComponent(req.id)}`,
      UpdateSecRateLimitRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateSecRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /security/secratelimit/{id}
export async function deleteSecRateLimit(
  req: DeleteSecRateLimitRequest,
  options?: ApiOptions,
): Promise<DeleteSecRateLimitResponse> {
  return withFallback<DeleteSecRateLimitResponse>('deleteSecRateLimit', () => {
    sampleRateLimitStore = sampleRateLimitStore.filter((r) => r.id !== req.id);
    return {};
  })(async () => {
    const endpoint = `${API_BASE_PATH}/security/secratelimit/${encodeURIComponent(req.id)}`;
    const resp = await axios.delete(endpoint, { signal: options?.signal });
    return DeleteSecRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/secrulesets
export async function listSecRulesets(
  req: Partial<ListSecRulesetsRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRulesetsResponse> {
  return withFallback('listSecRulesets', () => {
    const list = filterRulesets(req);
    return { secrulesets: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/secrulesets`, {
      params: toQuery(ListSecRulesetsRequest.toJSON(ListSecRulesetsRequest.fromPartial(req))),
      // repeat the keys of repeated filters (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListSecRulesetsResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/secruleset/{id}
export async function getSecRuleset(
  req: GetSecRulesetRequest,
  options?: ApiOptions,
): Promise<GetSecRulesetResponse> {
  return withFallback<GetSecRulesetResponse>('getSecRuleset', () => ({
    secruleset: sampleRulesetStore.find((r) => r.id === req.id) ?? sampleRulesetStore[0],
  }))(async () => {
    const endpoint = `${API_BASE_PATH}/security/secruleset/${encodeURIComponent(req.id)}`;
    const resp = await axios.get(endpoint, { signal: options?.signal });
    return GetSecRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// POST /security/secruleset
export async function createSecRuleset(
  req: CreateSecRulesetRequest,
  options?: ApiOptions,
): Promise<CreateSecRulesetResponse> {
  return withFallback('createSecRuleset', () => {
    const id = `sample-ruleset-${Date.now()}`;
    sampleRulesetStore = [...sampleRulesetStore, SecRuleset.fromPartial({ ...req.secruleset, id })];
    return { id };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/secruleset`,
      CreateSecRulesetRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateSecRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /security/secruleset/{id}
export async function updateSecRuleset(
  req: UpdateSecRulesetRequest,
  options?: ApiOptions,
): Promise<UpdateSecRulesetResponse> {
  return withFallback<UpdateSecRulesetResponse>('updateSecRuleset', () => {
    const owner = sampleRulesetStore.find((r) => r.id === req.id)?.owner;
    const next = SecRuleset.fromPartial({ ...req.secruleset, id: req.id, owner });
    sampleRulesetStore = sampleRulesetStore.map((r) => (r.id === req.id ? next : r));
    return { secruleset: next };
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/secruleset/${encodeURIComponent(req.id)}`,
      UpdateSecRulesetRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateSecRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /security/secruleset/{id}
export async function deleteSecRuleset(
  req: DeleteSecRulesetRequest,
  options?: ApiOptions,
): Promise<DeleteSecRulesetResponse> {
  return withFallback<DeleteSecRulesetResponse>('deleteSecRuleset', () => {
    sampleRulesetStore = sampleRulesetStore.filter((r) => r.id !== req.id);
    return {};
  })(async () => {
    const endpoint = `${API_BASE_PATH}/security/secruleset/${encodeURIComponent(req.id)}`;
    const resp = await axios.delete(endpoint, { signal: options?.signal });
    return DeleteSecRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/crs/rules
export async function listCRSRuleInfos(
  req: Partial<ListCRSRuleInfosRequest> = {},
  options?: ApiOptions,
): Promise<ListCRSRuleInfosResponse> {
  return withFallback('listCRSRuleInfos', () => {
    const list = filterCRSRuleInfos(req);
    return { rules: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/crs/rules`, {
      params: toQuery(ListCRSRuleInfosRequest.toJSON(ListCRSRuleInfosRequest.fromPartial(req))),
      // repeat the keys of repeated filters (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListCRSRuleInfosResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/crs/tags
export async function listCRSRuleTags(
  req: Partial<ListCRSRuleTagsRequest> = {},
  options?: ApiOptions,
): Promise<ListCRSRuleTagsResponse> {
  return withFallback('listCRSRuleTags', () => {
    const list = sampleCRSRuleTags(req);
    return { tags: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/crs/tags`, {
      params: toQuery(ListCRSRuleTagsRequest.toJSON(ListCRSRuleTagsRequest.fromPartial(req))),
      signal: options?.signal,
    });
    return ListCRSRuleTagsResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/crs/rules/{id}
export async function getCRSRuleInfo(
  req: Partial<GetCRSRuleInfoRequest> & Pick<GetCRSRuleInfoRequest, 'id'>,
  options?: ApiOptions,
): Promise<GetCRSRuleInfoResponse> {
  return withFallback<GetCRSRuleInfoResponse>('getCRSRuleInfo', () => ({
    rule: SAMPLE_RULE_INFOS.find((r) => r.id === req.id),
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/crs/rules/${req.id}`, {
      // `id` is the path param: zero it so toJSON leaves it out of the query
      params: toQuery(
        GetCRSRuleInfoRequest.toJSON(GetCRSRuleInfoRequest.fromPartial({ ...req, id: 0 })),
      ),
      signal: options?.signal,
    });
    return GetCRSRuleInfoResponse.fromJSON(resp.data ?? {});
  });
}

// ---- Sample condition schema: what core/szrule/rule_schema.go serves ----

const spec = (
  operator: Operator,
  valueKind: ValueKind,
  extra: Partial<OperatorSpec> = {},
): OperatorSpec => ({ operator, valueKind, list: false, enumValues: [], ...extra });

const TEXT_OPERATORS = [
  spec(Operator.OPERATOR_EQ, ValueKind.VALUE_KIND_STRING),
  spec(Operator.OPERATOR_NE, ValueKind.VALUE_KIND_STRING),
  spec(Operator.OPERATOR_CONTAINS, ValueKind.VALUE_KIND_STRING),
  spec(Operator.OPERATOR_PREFIX, ValueKind.VALUE_KIND_STRING),
  spec(Operator.OPERATOR_SUFFIX, ValueKind.VALUE_KIND_STRING),
  spec(Operator.OPERATOR_MATCHES, ValueKind.VALUE_KIND_REGEX),
];

const NUMBER_OPERATORS = [
  Operator.OPERATOR_GT,
  Operator.OPERATOR_GTE,
  Operator.OPERATOR_LT,
  Operator.OPERATOR_LTE,
].map((op) => spec(op, ValueKind.VALUE_KIND_NUMBER));

/** EQ and NE to a value, IN and NOT_IN to a list of values */
const membership = (valueKind: ValueKind, enumValues: string[] = []): OperatorSpec[] => [
  spec(Operator.OPERATOR_EQ, valueKind, { enumValues }),
  spec(Operator.OPERATOR_NE, valueKind, { enumValues }),
  spec(Operator.OPERATOR_IN, valueKind, { enumValues, list: true }),
  spec(Operator.OPERATOR_NOT_IN, valueKind, { enumValues, list: true }),
];

/** A source without a key: it has a single variant */
const plainSource = (source: FieldSource, operators: OperatorSpec[]): SourceSpec => ({
  source,
  variants: [{ name: 'value', keyMode: KeyMode.KEY_MODE_NONE, operators }],
});

/** Header and query: the value of a named field, or the presence of fields */
const keyedSource = (source: FieldSource): SourceSpec => ({
  source,
  variants: [
    {
      name: 'value',
      keyMode: KeyMode.KEY_MODE_REQUIRED,
      operators: [
        ...TEXT_OPERATORS,
        ...NUMBER_OPERATORS,
        spec(Operator.OPERATOR_IN, ValueKind.VALUE_KIND_STRING, { list: true }),
        spec(Operator.OPERATOR_NOT_IN, ValueKind.VALUE_KIND_STRING, { list: true }),
      ],
    },
    {
      name: 'presence',
      keyMode: KeyMode.KEY_MODE_NONE,
      operators: membership(ValueKind.VALUE_KIND_NAME),
    },
  ],
});

export const SAMPLE_CONDITION_SCHEMA: ConditionSchema = {
  sources: [
    keyedSource(FieldSource.FIELD_SOURCE_HEADER),
    keyedSource(FieldSource.FIELD_SOURCE_QUERY),
    plainSource(FieldSource.FIELD_SOURCE_PATH, TEXT_OPERATORS),
    plainSource(FieldSource.FIELD_SOURCE_BODY, TEXT_OPERATORS),
    plainSource(FieldSource.FIELD_SOURCE_IP, membership(ValueKind.VALUE_KIND_IP)),
    plainSource(FieldSource.FIELD_SOURCE_JA4, TEXT_OPERATORS),
    plainSource(FieldSource.FIELD_SOURCE_TLS, [
      spec(Operator.OPERATOR_EQ, ValueKind.VALUE_KIND_ENUM, { enumValues: ['true', 'false'] }),
      spec(Operator.OPERATOR_NE, ValueKind.VALUE_KIND_ENUM, { enumValues: ['true', 'false'] }),
    ]),
    plainSource(
      FieldSource.FIELD_SOURCE_METHOD,
      membership(ValueKind.VALUE_KIND_ENUM, [
        'GET',
        'HEAD',
        'POST',
        'PUT',
        'PATCH',
        'DELETE',
        'OPTIONS',
        'CONNECT',
        'TRACE',
      ]),
    ),
    plainSource(FieldSource.FIELD_SOURCE_HOST, TEXT_OPERATORS),
  ],
};

// GET /security/condition-schema
// The conditions a rule expression accepts: the editor offers what it lists.
export async function getConditionSchema(
  options?: ApiOptions,
): Promise<GetConditionSchemaResponse> {
  return withFallback<GetConditionSchemaResponse>('getConditionSchema', () => ({
    schema: SAMPLE_CONDITION_SCHEMA,
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/condition-schema`, {
      signal: options?.signal,
    });
    return GetConditionSchemaResponse.fromJSON(resp.data ?? {});
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
