import { SecRuleset, SecRateLimit, SecRule } from '@sentinez/proto/sentinez/apps/security/v1/model';
import {
  ListSecRulesetsRequest,
  ListCRSRuleInfosRequest,
  ListCRSRuleTagsRequest,
} from '@sentinez/proto/sentinez/apps/security/v1/security';
import { RuleInfo } from '@sentinez/proto/sentinez/types/ruleset/v1/setting';
import { providerOf } from '@/lib/type/security';
import { createCondition } from '@/lib/type/rule-expr';
import {
  ActionType,
  AndCondition,
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

import { SampleStore, byIds, byResources } from '@/lib/faker/store';
import { SAMPLE_RESOURCES, SAMPLE_TENANT_ID } from '@/lib/faker/tenant';

// Sample rows belong to the first sample resources of lib/faker/tenant
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

// Nested groups, built as AndConditions: all = AND, any = OR
const cond = (source: FieldSource, operator: Operator, value: unknown, key = '') => ({
  condition: createCondition({ source, operator, value, key }),
});

const leaf = (...args: Parameters<typeof cond>) =>
  AndCondition.fromPartial({ rules: [cond(...args)] });
const all = (...andCondition: AndCondition[]) => AndCondition.fromPartial({ andCondition });
const any = (...orCondition: AndCondition[]) => AndCondition.fromPartial({ orCondition });
const exprOf = (and: AndCondition): Expression => ({ orCondition: [and] });

// The scope of a ruleset exclusion or override, as proto JSON
const scope = (and: AndCondition) => Expression.toJSON(exprOf(and));

const SAMPLE_RULES: SecRule[] = [
  sampleRule(
    'sample-rule-1',
    'Block non-GET on /admin',
    'Sample: (Path prefix /admin AND Method != GET)',
    1,
    Status.STATUS_ACTIVE,
    Expression.fromPartial({
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
    }),
  ),
  sampleRule(
    'sample-rule-2',
    'Block bad bots',
    'Sample: User-Agent contains "curl" OR "sqlmap"',
    2,
    Status.STATUS_ACTIVE,
    Expression.fromPartial({
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
    }),
  ),
  sampleRule(
    'sample-rule-3',
    'Log requests from internal range',
    'Sample: IP in 10.0.0.0/8 (disabled)',
    3,
    Status.STATUS_DISABLE,
    Expression.fromPartial({
      orCondition: [
        {
          rules: [cond(FieldSource.FIELD_SOURCE_IP, Operator.OPERATOR_IN, ['10.0.0.0/8'])],
          orCondition: [],
          andCondition: [],
        },
      ],
    }),
    ActionType.ACTION_TYPE_LOG,
  ),
  sampleRule(
    'sample-rule-4',
    'Protect API writes',
    'Sample: /api AND (write method OR (debug query AND script UA)) AND IP not internal',
    4,
    Status.STATUS_ACTIVE,
    exprOf(
      all(
        leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/api'),
        any(
          leaf(FieldSource.FIELD_SOURCE_METHOD, Operator.OPERATOR_IN, ['POST', 'PUT', 'DELETE']),
          all(
            leaf(FieldSource.FIELD_SOURCE_QUERY, Operator.OPERATOR_EQ, '1', 'debug'),
            leaf(
              FieldSource.FIELD_SOURCE_HEADER,
              Operator.OPERATOR_MATCHES,
              '(?i)python|go-http',
              'User-Agent',
            ),
          ),
        ),
        leaf(FieldSource.FIELD_SOURCE_IP, Operator.OPERATOR_NOT_IN, ['10.0.0.0/8']),
      ),
    ),
  ),
  sampleRule(
    'sample-rule-5',
    'Admin host and login abuse',
    'Sample: 4 levels of nested groups under an OR root',
    5,
    Status.STATUS_ACTIVE,
    exprOf(
      any(
        all(
          leaf(FieldSource.FIELD_SOURCE_HOST, Operator.OPERATOR_EQ, 'admin.example.com'),
          any(
            leaf(FieldSource.FIELD_SOURCE_IP, Operator.OPERATOR_NOT_IN, [
              '10.0.0.0/8',
              '192.168.0.0/16',
            ]),
            all(
              leaf(
                FieldSource.FIELD_SOURCE_HEADER,
                Operator.OPERATOR_CONTAINS,
                ',',
                'X-Forwarded-For',
              ),
              any(
                leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_SUFFIX, '.php'),
                leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_MATCHES, '^/wp-'),
              ),
            ),
          ),
        ),
        all(
          leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_EQ, '/login'),
          leaf(FieldSource.FIELD_SOURCE_METHOD, Operator.OPERATOR_EQ, 'POST'),
          any(
            leaf(FieldSource.FIELD_SOURCE_HEADER, Operator.OPERATOR_EQ, '', 'Referer'),
            leaf(FieldSource.FIELD_SOURCE_JA4, Operator.OPERATOR_PREFIX, 't13d1516h2'),
          ),
        ),
      ),
    ),
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
  Expression.fromPartial({
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
  });

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

export const ruleStore = new SampleStore([
  ...SAMPLE_RULES,
  ...generateRules('sample-rule', SAMPLE_RULES.length + 1, 20),
]);

const sampleRateLimit = (
  rule: SecRule,
  timeWindow: string,
  maxRequests: number,
  timeout: string,
): SecRateLimit => ({ ...rule, timeWindow, maxRequests, timeout });

const SAMPLE_RATE_LIMITS: SecRateLimit[] = [
  sampleRateLimit(
    sampleRule(
      'sample-ratelimit-1',
      'Throttle login attempts',
      'Sample: POST /login, 5 requests per minute, blocked for 10m',
      1,
      Status.STATUS_ACTIVE,
      Expression.fromPartial({
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
      }),
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
      Expression.fromPartial({
        orCondition: [
          {
            rules: [cond(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/api')],
            orCondition: [],
            andCondition: [],
          },
        ],
      }),
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
      Expression.fromPartial({
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
      }),
      ActionType.ACTION_TYPE_LOG,
    ),
    '1s',
    50,
    '',
  ),
].map((r) => ofResource(r));

export const rateLimitStore = new SampleStore([
  ...SAMPLE_RATE_LIMITS,
  ...generateRules('sample-ratelimit', SAMPLE_RATE_LIMITS.length + 1, 12).map((r, i) =>
    sampleRateLimit(r, ['10s', '1m', '5m'][i % 3] ?? '1m', 50 * (i + 1), i % 2 ? '10m' : ''),
  ),
]);

const SAMPLE_RULESETS: SecRuleset[] = [
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
          // the search box takes SQL-like text, on the search pages only
          {
            targets: [
              { ruleId: 942100, variable: 'ARGS:q' },
              { ruleId: 942430, variable: 'ARGS:q' },
            ],
            expr: scope(
              any(
                leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/api/search'),
                leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_EQ, '/autocomplete'),
              ),
            ),
          },
          // the CMS editor posts HTML
          {
            tags: ['attack-xss'],
            expr: scope(
              all(
                leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/api/cms/pages'),
                leaf(FieldSource.FIELD_SOURCE_METHOD, Operator.OPERATOR_IN, ['POST', 'PUT']),
                leaf(FieldSource.FIELD_SOURCE_IP, Operator.OPERATOR_IN, ['10.0.0.0/8']),
              ),
            ),
          },
        ],
        overrides: [
          { id: 942100, state: 'RULE_STATE_DETECTION_ONLY' },
          { id: 941100, state: 'RULE_STATE_DISABLED' },
          // the admin tools run shell-like commands, from the office only
          {
            id: 932100,
            state: 'RULE_STATE_DISABLED',
            expr: scope(
              all(
                leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/admin/tools'),
                any(
                  leaf(FieldSource.FIELD_SOURCE_IP, Operator.OPERATOR_IN, [
                    '10.0.0.0/8',
                    '192.168.0.0/16',
                  ]),
                  leaf(FieldSource.FIELD_SOURCE_HEADER, Operator.OPERATOR_EQ, 'ops', 'X-Team'),
                ),
              ),
            ),
          },
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
  {
    id: 'sample-ruleset-4',
    name: 'WordPress',
    description: 'Sample: paranoia level 2, tuned for the WordPress admin and REST API',
    setting: {
      status: 'STATUS_ACTIVE',
      provider: 'PROVIDER_OWASP',
      crs: {
        mode: 'ENGINE_MODE_ON',
        version: 'CRS_VERSION_V4_17_0',
        paranoiaLevel: 2,
        threshold: { inbound: 7, outbound: 4 },
        request: { bodyAccess: true, allowedMethods: ['GET', 'HEAD', 'POST', 'OPTIONS'] },
        exclusions: [
          // the post editor saves HTML and scripts
          {
            tags: ['attack-xss', 'attack-rce'],
            expr: scope(
              all(
                leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_EQ, '/wp-admin/post.php'),
                leaf(FieldSource.FIELD_SOURCE_METHOD, Operator.OPERATOR_EQ, 'POST'),
              ),
            ),
          },
          {
            targets: [{ ruleId: 942430, variable: 'ARGS:content' }],
            expr: scope(
              leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_PREFIX, '/wp-json/wp/v2/'),
            ),
          },
        ],
        overrides: [
          { id: 911100, state: 'RULE_STATE_DETECTION_ONLY' },
          // the health checks of the load balancer call by IP
          {
            id: 920350,
            state: 'RULE_STATE_DISABLED',
            expr: scope(
              all(
                leaf(FieldSource.FIELD_SOURCE_PATH, Operator.OPERATOR_EQ, '/healthz'),
                leaf(
                  FieldSource.FIELD_SOURCE_HEADER,
                  Operator.OPERATOR_PREFIX,
                  'ELB',
                  'User-Agent',
                ),
              ),
            ),
          },
        ],
      },
    },
  },
].map((r) => ofResource(SecRuleset.fromJSON(r)));

export const rulesetStore = new SampleStore(SAMPLE_RULESETS);

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

export const filterRulesets = (req: Partial<ListSecRulesetsRequest>) =>
  byResources(byIds(rulesetStore.all(), req.ids), req.resourceIds).filter(
    (r) =>
      includesText(r.name, req.name) &&
      (!req.modes?.length ||
        req.modes.some((m) => m === r.setting?.crs?.mode || m === r.setting?.managed?.mode)) &&
      anyOf(req.providers, providerOf(r.setting)) &&
      anyOf(req.versions, r.setting?.crs?.version) &&
      anyOf(req.statuses, r.setting?.status) &&
      (!req.paranoiaLevel || r.setting?.crs?.paranoiaLevel === req.paranoiaLevel),
  );

export const filterCRSRuleInfos = (req: Partial<ListCRSRuleInfosRequest>) =>
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

export const sampleCRSRuleTags = (req: Partial<ListCRSRuleTagsRequest>) => {
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
