import { RuleBased } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { FieldSource, Operator } from '@sentinez/proto/sentinez/secure/rule/v1/engine';
import { Status } from '@sentinez/proto/sentinez/types/v1/known';
import axios from 'axios';

const API_BASE_PATH = process.env.SNTZ_BASE_PATH || 'http://localhost:8080';

export interface ApiOptions {
  signal?: AbortSignal;
}

export const MOCK_RULE: RuleBased = {
  enable: true,
  ingressFull: {
    id: 'proxy-rule-1',
    name: 'Complex Logic: (Path AND Method) OR IP',
    description: 'Mocked rule from proxy.yaml',
    priority: 1,
    status: Status.STATUS_ACTIVE,
    expr: {
      orCondition: [
        {
          rules: [
            {
              id: 'nxaxo3j',
              name: '',
              description: '',
              condition: {
                id: '0f1i7zh',
                source: FieldSource.FIELD_SOURCE_METHOD,
                key: '',
                operator: Operator.OPERATOR_EQ,
                value: '',
              },
            },
          ],
          orCondition: [],
        },
      ],
    },
  },
};

export async function listRuleBaseds(options?: ApiOptions) {
  return [MOCK_RULE];
}

export async function getRuleBased(id: string, options?: ApiOptions) {
  return MOCK_RULE;
}

export async function createRuleBased(data: RuleBased, options?: ApiOptions) {}

export async function updateRuleBased(id: string, data: RuleBased, options?: ApiOptions) {}
