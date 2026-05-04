import axios from 'axios';

const API_BASE_PATH = process.env.SNTZ_BASE_PATH || 'http://localhost:8080';

export interface ApiOptions {
  signal?: AbortSignal;
}

export interface RuleBased {
  metadata?: {
    createdAt?: string;
    updatedAt?: string;
  };
  id?: string;
  name?: string;
  description?: string;
  node?: ApiRuleGroup;
  action?: any;
  status?: string;
  priority?: number;
}

export type Logic = 'LOGIC_UNSPECIFIED' | 'LOGIC_AND' | 'LOGIC_OR' | 'LOGIC_NOT';
export type FieldSource =
  | 'FIELD_SOURCE_UNSPECIFIED'
  | 'FIELD_SOURCE_HEADER'
  | 'FIELD_SOURCE_QUERY'
  | 'FIELD_SOURCE_PATH'
  | 'FIELD_SOURCE_BODY'
  | 'FIELD_SOURCE_IP'
  | 'FIELD_SOURCE_JA4'
  | 'FIELD_SOURCE_TLS'
  | 'FIELD_SOURCE_METHOD'
  | 'FIELD_SOURCE_HOST';

export type Operator =
  | 'OPERATOR_UNSPECIFIED'
  | 'OPERATOR_EQ'
  | 'OPERATOR_NE'
  | 'OPERATOR_CONTAINS'
  | 'OPERATOR_MATCHES'
  | 'OPERATOR_IN'
  | 'OPERATOR_PREFIX'
  | 'OPERATOR_SUFFIX'
  | 'OPERATOR_GT'
  | 'OPERATOR_GTE'
  | 'OPERATOR_LT'
  | 'OPERATOR_LTE'
  | 'OPERATOR_NOT_IN';

export interface ApiRule {
  id?: string;
  field?: FieldSource;
  key?: string;
  operator?: Operator;
  value?: string;
}

export interface ApiRuleGroup {
  id?: string;
  combinator?: Logic;
  rules?: ApiRuleNode[];
  not?: boolean;
}

export interface ApiRuleNode {
  rule?: ApiRule;
  group?: ApiRuleGroup;
}

// Mapping Logic
const reverseFieldMap: Record<string, FieldSource> = {
  'http.header': 'FIELD_SOURCE_HEADER',
  'http.query': 'FIELD_SOURCE_QUERY',
  'http.path': 'FIELD_SOURCE_PATH',
  'http.request.path': 'FIELD_SOURCE_PATH',
  'http.method': 'FIELD_SOURCE_METHOD',
  'http.host': 'FIELD_SOURCE_HOST',
  'ip.src': 'FIELD_SOURCE_IP',
  'net.ip': 'FIELD_SOURCE_IP',
};

const forwardFieldMap: Record<FieldSource, string> = {
  FIELD_SOURCE_UNSPECIFIED: '',
  FIELD_SOURCE_HEADER: 'http.header',
  FIELD_SOURCE_QUERY: 'http.query',
  FIELD_SOURCE_PATH: 'http.path',
  FIELD_SOURCE_BODY: 'http.body',
  FIELD_SOURCE_IP: 'ip.src',
  FIELD_SOURCE_JA4: 'ja4',
  FIELD_SOURCE_TLS: 'tls',
  FIELD_SOURCE_METHOD: 'http.method',
  FIELD_SOURCE_HOST: 'http.host',
};

export function transformApiToUi(apiGroup: ApiRuleGroup): any {
  if (!apiGroup) return undefined;

  const combinatorMap: any = {
    LOGIC_AND: 'and',
    LOGIC_OR: 'or',
  };

  const operatorMap: any = {
    OPERATOR_EQ: '==',
    OPERATOR_NE: '!=',
    OPERATOR_GT: '>',
    OPERATOR_LT: '<',
    OPERATOR_GTE: '>=',
    OPERATOR_LTE: '<=',
    OPERATOR_CONTAINS: 'contains',
    OPERATOR_PREFIX: 'startsWith',
    OPERATOR_SUFFIX: 'endsWith',
    OPERATOR_IN: 'in',
  };

  const traverse = (node: ApiRuleGroup): any => {
    return {
      id: node.id || Math.random().toString(36).substring(2, 9),
      combinator: combinatorMap[node.combinator || 'LOGIC_AND'] || 'and',
      not: node.not,
      rules: (node.rules || [])
        .map((rn: ApiRuleNode) => {
          if (rn.group) {
            return traverse(rn.group);
          }
          if (rn.rule) {
            const prefix = forwardFieldMap[rn.rule.field || 'FIELD_SOURCE_UNSPECIFIED'];
            const field = rn.rule.key ? `${prefix}.${rn.rule.key}` : prefix;

            return {
              id: rn.rule.id || Math.random().toString(36).substring(2, 9),
              field: field || '',
              operator: operatorMap[rn.rule.operator || 'OPERATOR_EQ'] || '==',
              value: rn.rule.value || '',
            };
          }
          return null;
        })
        .filter(Boolean),
    };
  };

  return traverse(apiGroup);
}

export function transformUiToApi(uiGroup: any): ApiRuleGroup {
  if (!uiGroup)
    return {
      id: 'root',
      combinator: 'LOGIC_AND',
      rules: [],
    };

  const combinatorMap: any = {
    and: 'LOGIC_AND',
    or: 'LOGIC_OR',
  };

  const operatorMap: any = {
    '==': 'OPERATOR_EQ',
    '!=': 'OPERATOR_NE',
    '>': 'OPERATOR_GT',
    '<': 'OPERATOR_LT',
    '>=': 'OPERATOR_GTE',
    '<=': 'OPERATOR_LTE',
    contains: 'OPERATOR_CONTAINS',
    startsWith: 'OPERATOR_PREFIX',
    endsWith: 'OPERATOR_SUFFIX',
    in: 'OPERATOR_IN',
  };

  const traverse = (node: any): ApiRuleGroup => {
    return {
      id: node.id,
      combinator: combinatorMap[node.combinator] || 'LOGIC_AND',
      not: node.not,
      rules: (node.rules || []).map((uiRule: any) => {
        if (uiRule.combinator) {
          return {
            group: traverse(uiRule),
          };
        }

        // Parse legacy string fields back to source + key
        let fieldSource: FieldSource = 'FIELD_SOURCE_UNSPECIFIED';
        let key = '';

        if (uiRule.field) {
          // Attempt to find the longest matching prefix from our map
          const parts = uiRule.field.split('.');
          for (let i = parts.length; i > 0; i--) {
            const prefix = parts.slice(0, i).join('.');
            if (reverseFieldMap[prefix]) {
              fieldSource = reverseFieldMap[prefix];
              key = parts.slice(i).join('.');
              break;
            }
          }
        }

        return {
          rule: {
            id: uiRule.id,
            field: fieldSource,
            key: key,
            operator: operatorMap[uiRule.operator] || 'OPERATOR_EQ',
            value: uiRule.value,
          },
        };
      }),
    };
  };

  return traverse(uiGroup);
}

export function transformApiToUiRuleBased(apiData: any): RuleBased {
  if (!apiData) return apiData;

  let action = apiData.action;
  if (typeof action === 'string' && action.trim().startsWith('{')) {
    try {
      action = JSON.parse(action);
    } catch (e) {
      // Keep as string if parsing fails
    }
  }

  return {
    ...apiData,
    node: apiData.node ? transformApiToUi(apiData.node) : undefined,
    action: action,
  };
}

export function transformUiToApiRuleBased(uiData: any): RuleBased {
  if (!uiData) return uiData;

  let action = uiData.action;
  if (action && typeof action !== 'string') {
    action = JSON.stringify(action);
  }

  return {
    ...uiData,
    node: transformUiToApi(uiData.node),
    action: action,
  };
}

export const MOCK_RULE: RuleBased = {
  id: 'proxy-rule-1',
  name: 'Complex Logic: (Path AND Method) OR IP',
  description: 'Mocked rule from proxy.yaml',
  status: 'STATUS_ACTIVE',
  priority: 1,
  node: {
    id: 'root',
    combinator: 'LOGIC_OR',
    rules: [
      {
        rule: {
          id: 'rule-1',
          field: 'FIELD_SOURCE_PATH',
          key: '',
          operator: 'OPERATOR_EQ',
          value: '/block',
        },
      },
      {
        group: {
          id: 'group-1',
          combinator: 'LOGIC_AND',
          rules: [
            {
              rule: {
                id: 'rule-2',
                field: 'FIELD_SOURCE_IP',
                key: '',
                operator: 'OPERATOR_EQ',
                value: '1.1.1.1',
              },
            },
            {
              rule: {
                id: 'rule-3',
                field: 'FIELD_SOURCE_METHOD',
                key: '',
                operator: 'OPERATOR_EQ',
                value: 'POST',
              },
            },
          ],
        },
      },
    ],
  },
  action: {
    type: 'ACTION_TYPE_BLOCK',
  },
};

export async function listRuleBaseds(options?: ApiOptions) {
  try {
    const endpoint = `${API_BASE_PATH}/security/rulebaseds`;
    const resp = await axios.get(endpoint, {
      signal: options?.signal,
    });
    const ruleBaseds = resp.data?.ruleBaseds || [MOCK_RULE];
    return ruleBaseds.map(transformApiToUiRuleBased);
  } catch (err: any) {
    if (axios.isCancel(err)) throw err;
    console.error('Error listing rulebaseds:', err.message);
    return [MOCK_RULE].map(transformApiToUiRuleBased);
  }
}

export async function getRuleBased(id: string, options?: ApiOptions) {
  try {
    const endpoint = `${API_BASE_PATH}/security/rulebased/${id}`;
    const resp = await axios.get(endpoint, {
      signal: options?.signal,
    });
    return transformApiToUiRuleBased(resp.data?.ruleBased || MOCK_RULE);
  } catch (err: any) {
    if (axios.isCancel(err)) throw err;
    console.error('Error getting rulebased:', err.message);
    return transformApiToUiRuleBased(MOCK_RULE);
  }
}

export async function createRuleBased(data: RuleBased, options?: ApiOptions) {
  try {
    const endpoint = `${API_BASE_PATH}/security/rulebased`;
    const payload = transformUiToApiRuleBased(data);
    const resp = await axios.post(endpoint, { ruleBased: payload }, { signal: options?.signal });
    return resp.data;
  } catch (err: any) {
    if (axios.isCancel(err)) throw err;
    console.error('Error creating rule based:', err.message);
    throw err;
  }
}

export async function updateRuleBased(id: string, data: RuleBased, options?: ApiOptions) {
  try {
    const endpoint = `${API_BASE_PATH}/security/rulebased/${id}`;
    const payload = transformUiToApiRuleBased(data);
    const resp = await axios.put(endpoint, { ruleBased: payload }, { signal: options?.signal });
    return resp.data;
  } catch (err: any) {
    if (axios.isCancel(err)) throw err;
    console.error('Error updating rule based:', err.message);
    throw err;
  }
}
